package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path"
	"time"

	featuresv1alpha1 "github.com/criticalstack/stackapps/api/v1alpha1"
	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	promv1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/yaml"
)

// StackReleaseReconciler reconciles a StackRelease object
type StackReleaseReconciler struct {
	client.Client
	Log     logr.Logger
	Scheme  *runtime.Scheme
	Metrics promv1.API
}

func (r *StackReleaseReconciler) isServiceRateGreaterThanZero(ctx context.Context, service, code string) (bool, error) {
	vector, err := r.serviceQuery(ctx, service, code)
	if err != nil {
		return false, err
	}
	for _, rate := range vector {
		if !rate.Value.Equal(model.ZeroSample.Value) {
			return true, nil
		}
	}
	return false, nil
}

func (r *StackReleaseReconciler) serviceQuery(ctx context.Context, service, code string) (model.Vector, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	query := fmt.Sprintf(`rate(traefik_service_requests_total{service="%s",code=~"%s"}[30s])`, service, code)
	result, _, err := r.Metrics.Query(ctx, query, time.Now())
	if err != nil {
		return nil, err
	}
	vector, ok := result.(model.Vector)
	if !ok {
		return nil, errors.Errorf("query result was %T, not model.Vector", result)
	}
	return vector, nil
}

func (r *StackReleaseReconciler) SetupWithManager(mgr ctrl.Manager) error {
	err := ctrl.NewControllerManagedBy(mgr).
		For(&featuresv1alpha1.StackRelease{}).
		Owns(&featuresv1alpha1.AppRevision{}).
		Complete(r)
	if err != nil {
		return err
	}

	if err := mgr.GetFieldIndexer().IndexField(context.Background(), &featuresv1alpha1.AppRevision{}, "metadata.ownerReferences.name", func(o runtime.Object) []string {
		names := make([]string, 0)
		for _, ref := range o.(*featuresv1alpha1.AppRevision).GetOwnerReferences() {
			names = append(names, ref.Name)
		}
		return names
	}); err != nil {
		return err
	}
	return nil
}

func (r *StackReleaseReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	reqLogger := r.Log.WithValues("Request.Name", req.Name)
	reqLogger.Info("Reconciling StackRelease")
	stackRelease := &featuresv1alpha1.StackRelease{}
	if err := r.Client.Get(ctx, req.NamespacedName, stackRelease); err != nil {
		if apierrors.IsNotFound(err) {
			return reconcile.Result{}, nil
		}
		reqLogger.Info("Failed to retrieve StackRelease before reconciling", err)
		return reconcile.Result{}, err
	}

	switch stackRelease.Spec.Config.BackendType {
	case featuresv1alpha1.TraefikBackend:
		requeue, cond := r.proxyExists(ctx, stackRelease.Spec.Config.ProxyNamespace)
		r.setCondition(stackRelease, &cond)
		if requeue {
			if err := r.Client.Status().Update(ctx, stackRelease); err != nil {
				return ctrl.Result{}, err
			}
			return ctrl.Result{Requeue: true}, nil
		}
	case featuresv1alpha1.NoReleaseBackend:
	default:
		err := errors.Errorf("invalid backend type %q", stackRelease.Spec.Config.BackendType)
		r.setCondition(stackRelease, &featuresv1alpha1.StackReleaseCondition{
			Type:    featuresv1alpha1.StackReleaseError,
			Status:  corev1.ConditionTrue,
			Message: err.Error(),
		})
		if err := r.Client.Status().Update(ctx, stackRelease); err != nil {
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, err
	}
	requeue, cond := r.createRevision(ctx, stackRelease)
	r.setCondition(stackRelease, &cond)
	if cond.Type != featuresv1alpha1.StackReleaseError {
		// no error occurred, clear error
		r.setCondition(stackRelease, &featuresv1alpha1.StackReleaseCondition{Type: featuresv1alpha1.StackReleaseError, Status: corev1.ConditionFalse})
	}
	if err := r.Status().Update(ctx, stackRelease); err != nil {
		return ctrl.Result{}, err
	}
	if requeue {
		return ctrl.Result{RequeueAfter: 10 * time.Second}, nil
	}
	return ctrl.Result{Requeue: requeue}, nil
}

func (r *StackReleaseReconciler) proxyExists(ctx context.Context, ns string) (bool, featuresv1alpha1.StackReleaseCondition) {
	var deps appsv1.DeploymentList
	if err := r.List(ctx, &deps, client.InNamespace(ns), client.MatchingLabels{"stackreleaseproxy": "true"}); err != nil {
		return false, buildCondition(featuresv1alpha1.StackReleaseError, corev1.ConditionTrue, fmt.Sprintf("failed listing deployments in app namespace: %v", err))
	}

	switch n := len(deps.Items); n {
	case 1:
		// good, continue
	case 0:
		if err := deployDirectory(ctx, r.Client, ns, "manifests/"); err != nil {
			return false, buildCondition(featuresv1alpha1.StackReleaseError, corev1.ConditionTrue, err.Error())
		}
		return true, buildCondition(featuresv1alpha1.StackReleaseInstalling, corev1.ConditionTrue, "starting proxy deployment")
	default: // > 1
		return false, buildCondition(featuresv1alpha1.StackReleaseError, corev1.ConditionTrue, fmt.Sprintf("%d proxy deployments found, should never exceed one", n))
	}
	if deps.Items[0].Status.AvailableReplicas < 1 {
		return true, buildCondition(featuresv1alpha1.StackReleaseInstalling, corev1.ConditionTrue, "waiting for available proxy replicas")
	}
	return false, buildCondition(featuresv1alpha1.StackReleaseInstalling, corev1.ConditionTrue, "Proxy ready, starting app install")

}

func deployYaml(ctx context.Context, c client.Client, ns string, fileName string) error {
	rawManifest, err := ioutil.ReadFile(fileName)
	if err != nil {
		return err
	}
	uns := unstructured.Unstructured{}
	yaml.Unmarshal(rawManifest, &uns.Object)
	uns.SetNamespace(ns)
	if err != nil {
		return err
	}
	if err := createOrUpdate(ctx, c, &uns); err != nil {
		return err
	}
	return nil
}

func createOrUpdate(ctx context.Context, c client.Client, u *unstructured.Unstructured) error {
	ok := client.ObjectKey{
		Name:      u.GetName(),
		Namespace: u.GetNamespace(),
	}
	var found unstructured.Unstructured
	found.SetKind(u.GetKind())
	found.SetAPIVersion(u.GetAPIVersion())
	if err := c.Get(ctx, ok, &found); err != nil {
		if apierrors.IsNotFound(err) {
			return c.Create(ctx, u)
		}
		return err
	}
	u.Object["metadata"] = found.Object["metadata"]
	return c.Update(ctx, u)
}

func (r *StackReleaseReconciler) canaryController(ctx context.Context, sr *featuresv1alpha1.StackRelease) (bool, featuresv1alpha1.StackReleaseCondition) {
	irNsN := types.NamespacedName{Name: sr.ObjectMeta.Name,
		Namespace: sr.Spec.Config.ProxyNamespace,
	}
	ir := unstructured.Unstructured{}
	ir.SetKind("IngressRoute")
	ir.SetAPIVersion("traefik.containo.us/v1alpha1")
	if err := r.Get(ctx, irNsN, &ir); err != nil {
		if !apierrors.IsNotFound(err) {
			return false, buildCondition(featuresv1alpha1.StackReleaseError,
				"true",
				"Failed to get Traefik IngressRoute")
		}

		// no routing exists from load balancer, define routing to initial revision
		var services []map[string]interface{}
		serviceName := fmt.Sprintf("%s-r%d", sr.Spec.AppName, sr.Spec.AppRevision.Revision)
		services = append(services, defineTraefikService(serviceName, 8080, 100))
		if err := r.deployTraefikIngress(ctx, sr, services); err != nil {
			return false, buildCondition(featuresv1alpha1.StackReleaseError,
				"true",
				"Failed to Deploy Traefik IngressRoute")
		}
		return true, buildCondition(featuresv1alpha1.StackReleaseInstalling,
			"true",
			"Deployed ingress")
	}
	services, err := getTraefikServices(ir, sr.Spec.Config.HostName)
	if err != nil {
		return false, buildCondition(featuresv1alpha1.StackReleaseError,
			"true",
			"Failed get services from IngressRoute")
	}
	switch len(services) {
	case 1:
		svcName := services[0].Name
		if svcName == revisionedName(sr) {
			// stackRelease is happy
			return false, buildCondition(featuresv1alpha1.StackReleaseDeploying,
				"false",
				"deployment finished")
		}
		revisionHealthy, err := r.checkAppRevisionHealth(ctx, sr)
		if err != nil {
			return false, buildCondition(featuresv1alpha1.StackReleaseError,
				"true",
				"unable to retrieve AppRevision Health")
		}
		switch revisionHealthy {
		case corev1.ConditionTrue:
			if err := r.setRollBackService(ctx, sr, svcName); err != nil {
				return false, buildCondition(featuresv1alpha1.StackReleaseError,
					"true",
					"unable to set RollBackService")
			}
			return r.startCanaryDeployment(ctx, sr)
		default:
			return true, buildCondition(featuresv1alpha1.StackReleaseDeploying,
				"true",
				"waiting for AppRecvision to become Healthy")
		}
	case 2: // ongoing deployment
		rollback, condition := r.rollBack(ctx, sr)
		if rollback || sr.Status.State == featuresv1alpha1.StackReleaseStateRollback {
			return false, condition
		}
		deploying, err := r.deployCanaryStep(ctx, sr)
		if err != nil {
			return false, buildCondition(featuresv1alpha1.StackReleaseError,
				"true",
				"Error Deploying canary step")
		}
		if !deploying {
			return false, buildCondition(featuresv1alpha1.StackReleaseDeploying,
				"false",
				"Deployment finished.")
		}
	default:
		return false, buildCondition(featuresv1alpha1.StackReleaseError,
			"true",
			"Unknown configuration of IngressRoute")
	}
	return true, buildCondition(featuresv1alpha1.StackReleaseDeploying,
		"true",
		"Canary Deployment in progress")
}

func (r *StackReleaseReconciler) checkAppRevisionHealth(ctx context.Context, sr *featuresv1alpha1.StackRelease) (corev1.ConditionStatus, error) {
	ar := featuresv1alpha1.AppRevision{}
	if err := r.Get(ctx, types.NamespacedName{Name: sr.ObjectMeta.Name, Namespace: revisionedName(sr)}, &ar); err != nil {
		return corev1.ConditionFalse, err
	}
	for _, c := range ar.Status.Conditions {
		if c.Type == featuresv1alpha1.AppRevisionReady {
			return c.Status, nil
		}
	}
	return corev1.ConditionFalse, nil
}

func (r *StackReleaseReconciler) startCanaryDeployment(ctx context.Context, sr *featuresv1alpha1.StackRelease) (bool, featuresv1alpha1.StackReleaseCondition) {
	firstStep := sr.Spec.Config.ReleaseStages[0]
	switch sr.Spec.Config.BackendType {
	case featuresv1alpha1.TraefikBackend:
		var services []map[string]interface{}
		stable := defineTraefikService(sr.Spec.RollBackService, 8080, 100-1*firstStep.CanaryWeight)
		canary := defineTraefikService(revisionedName(sr), 8080, 1*firstStep.CanaryWeight)
		services = append(services, stable, canary)
		sr.Status.CurrentCanaryWeight = firstStep
		sr.Status.CurrentCanaryWeight.NextStep = &metav1.Time{Time: metav1.Now().Add(firstStep.StepDuration.Duration)}
		if err := r.deployTraefikIngress(ctx, sr, services); err != nil {
			return false, buildCondition(featuresv1alpha1.StackReleaseError, "true", "Error Deploying Traefik IngressRoute")
		}
		return true, buildCondition(featuresv1alpha1.StackReleaseDeploying, "true", "Canary Deployment in progress")
	default:
		return false, buildCondition(featuresv1alpha1.StackReleaseError, "true", "unknown backend type")
	}
}
func (r *StackReleaseReconciler) checkRollBack(ctx context.Context, sr *featuresv1alpha1.StackRelease) error {
	healthy, err := r.checkAppRevisionHealth(ctx, sr)
	if err != nil {
		return errors.Wrap(err, "error checking appRevision health")
	}
	if healthy == corev1.ConditionFalse {
		return errors.Errorf("AppRevision is no longer healthy")
	}

	serviceName := fmt.Sprintf(`%s-%s-%s@kubernetescrd`, sr.Spec.Config.ProxyNamespace, revisionedName(sr), "8080")
	serverError, err := r.isServiceRateGreaterThanZero(ctx, serviceName, "5..")
	if err != nil {
		return errors.Wrap(err, "error checking http response codes")
	}
	if serverError {
		return errors.Errorf("New revision encountered internal server error")

	}
	return nil
}

func (r *StackReleaseReconciler) rollBack(ctx context.Context, sr *featuresv1alpha1.StackRelease) (bool, featuresv1alpha1.StackReleaseCondition) {
	if err := r.checkRollBack(ctx, sr); err != nil {
		// there is cause for rollBack
		switch sr.Spec.Config.BackendType {
		case featuresv1alpha1.TraefikBackend:
			var services []map[string]interface{}
			stable := defineTraefikService(sr.Spec.RollBackService, 8080, 100)
			services = append(services, stable)
			if err := r.deployTraefikIngress(ctx, sr, services); err != nil {
				return false, buildCondition(featuresv1alpha1.StackReleaseError, "true", "Error Deploying Traefik IngressRoute")
			}
			sr.Status.State = featuresv1alpha1.StackReleaseStateRollback
			return false, buildCondition(featuresv1alpha1.StackReleaseDeploying, "false", "Canary was unhealthy, rollback occured")
		default:
			return false, buildCondition(featuresv1alpha1.StackReleaseError, "true", "unknown backend type")
		}
	}
	return false, featuresv1alpha1.StackReleaseCondition{}
}

func (r *StackReleaseReconciler) setRollBackService(ctx context.Context, sr *featuresv1alpha1.StackRelease, rbs string) error {
	_, err := controllerutil.CreateOrUpdate(ctx, r.Client, sr, func() error {
		sr.Spec.RollBackService = rbs
		return nil
	})
	return err
}

func (r *StackReleaseReconciler) deployCanaryStep(ctx context.Context, sr *featuresv1alpha1.StackRelease) (bool, error) {
	currentStage := sr.Status.CurrentCanaryWeight
	releaseStages := sr.Spec.Config.ReleaseStages
	currentStage = sr.Status.CurrentCanaryWeight
	for i, releaseStage := range releaseStages {
		if releaseStage.CanaryWeight == currentStage.CanaryWeight {
			if metav1.Now().After(currentStage.NextStep.Time) {
				if i+1 < len(releaseStages) {
					duration := releaseStages[i+1].StepDuration.Duration
					releaseStages[i+1].NextStep = &metav1.Time{Time: metav1.Now().Add(duration)}
					currentStage = releaseStages[i+1]
					break
				} else {
					var services []map[string]interface{}
					stable := defineTraefikService(revisionedName(sr), 8080, 100)
					services = append(services, stable)
					if err := r.deployTraefikIngress(ctx, sr, services); err != nil {
						return true, errors.Wrap(err, "error deploying Traefik ingressRoute")
					}
					sr.Status.CurrentCanaryWeight = featuresv1alpha1.ReleaseStage{}
					return false, nil // deployment finished

				}
			} else {
				return true, nil // waiting for next step time
			}
		}
	}
	sr.Status.CurrentCanaryWeight = currentStage
	sr.Status.CurrentCanaryWeight.NextStep = &metav1.Time{Time: metav1.Now().Add(currentStage.StepDuration.Duration)}
	switch sr.Spec.Config.BackendType {
	case featuresv1alpha1.TraefikBackend:
		var services []map[string]interface{}
		stable := defineTraefikService(sr.Spec.RollBackService, 8080, 100-currentStage.CanaryWeight)
		canary := defineTraefikService(revisionedName(sr), 8080, currentStage.CanaryWeight)
		services = append(services, stable, canary)
		if err := r.deployTraefikIngress(ctx, sr, services); err != nil {
			return true, errors.Wrap(err, "error deploying Traefik ingressRoute")
		}
	default:
		return false, nil
	}
	return true, nil // deployment ongoing
}

func revisionedName(sr *featuresv1alpha1.StackRelease) string {
	return fmt.Sprintf("%s-r%d", sr.Spec.AppName, sr.Spec.AppRevision.Revision)
}

func (r *StackReleaseReconciler) deployTraefikIngress(ctx context.Context, sr *featuresv1alpha1.StackRelease, services []map[string]interface{}) error {
	u := unstructured.Unstructured{}
	u.SetName(sr.ObjectMeta.Name)
	u.SetNamespace(sr.Spec.Config.ProxyNamespace)
	u.SetKind("IngressRoute")
	u.SetAPIVersion("traefik.containo.us/v1alpha1")
	_, err := controllerutil.CreateOrUpdate(ctx, r.Client, &u, func() error {
		ts := defineTraefikSpec(services, sr.Spec.Config.HostName)
		b, err := json.Marshal(ts)
		if err != nil {
			return err
		}
		var m map[string]interface{}
		if err := json.Unmarshal(b, &m); err != nil {
			return err
		}
		unstructured.SetNestedMap(u.Object, m, "spec")
		return controllerutil.SetControllerReference(sr, &u, r.Scheme)
	})
	return err
}

func defineTraefikSpec(services []map[string]interface{}, host string) map[string]interface{} {
	spec := make(map[string]interface{})
	spec["entrypoints"] = []string{"web"}
	spec["routes"] = defineTraefikRoute(services, host)
	return spec
}

func defineTraefikService(sn string, port int32, weight uint8) map[string]interface{} {
	ts := map[string]interface{}{
		"kind":   "Service",
		"name":   sn,
		"port":   port,
		"weight": weight,
	}
	return ts
}

type service struct {
	Kind   string `json:"kind"`
	Name   string `json:"name"`
	Port   uint16 `json:"port"`
	Weight int    `json:"weight,omitempty"`
}
type route struct {
	Kind     string    `json:"kind"`
	Match    string    `json:"match"`
	Services []service `json:"services"`
}

func getTraefikServices(u unstructured.Unstructured, host string) ([]service, error) {
	var currentRoutes []route
	routes, _, _ := unstructured.NestedSlice(u.Object, "spec", "routes")
	routesJson, _ := json.Marshal(routes)
	err := json.Unmarshal(routesJson, &currentRoutes)
	if err != nil {
		return nil, err
	}
	for _, route := range currentRoutes {
		if route.Match == fmt.Sprintf("Host(`%s`)", host) {
			return route.Services, nil
		}
	}
	return nil, errors.Errorf("Host %s not found in current ingressRoute", host)
}

func defineTraefikRoute(services []map[string]interface{}, host string) []map[string]interface{} {
	var routes []map[string]interface{}
	routes = append(routes, map[string]interface{}{
		"kind":     "Rule",
		"match":    fmt.Sprintf("Host(`%s`)", host),
		"services": services,
	},
	)

	return routes
}

func (r *StackReleaseReconciler) createRevision(ctx context.Context, sr *featuresv1alpha1.StackRelease) (bool, featuresv1alpha1.StackReleaseCondition) {
	var ns string
	revisionName := fmt.Sprintf("%s-r%d", sr.ObjectMeta.Name, sr.Spec.AppRevision.Revision)
	switch sr.Spec.Config.BackendType {
	case featuresv1alpha1.TraefikBackend:
		// Create namespace for new revision
		ns = revisionName
		newNs := &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: ns,
			},
		}
		if err := r.Get(ctx, types.NamespacedName{Name: ns}, newNs); err != nil {
			if apierrors.IsNotFound(err) {
				if err := r.Create(ctx, newNs); err != nil {
					return false, buildCondition(featuresv1alpha1.StackReleaseError,
						corev1.ConditionTrue,
						fmt.Sprintf("failed to create revision namespace: %v", err))
				}
				return true, buildCondition(featuresv1alpha1.StackReleaseInstalling,
					corev1.ConditionTrue,
					"Building Namespace")
			}
			return false, buildCondition(featuresv1alpha1.StackReleaseError,
				corev1.ConditionTrue,
				fmt.Sprintf("failed to discover revision namespace: %v", err))
		}
		if err := controllerutil.SetControllerReference(sr, newNs, r.Scheme); err != nil {
			return false, buildCondition(featuresv1alpha1.StackReleaseError,
				corev1.ConditionTrue,
				fmt.Sprintf("failure creating namespace %v", err))
		}
		// ExternalName service to go to new revision namespace
		if err := r.defineRevisonService(ctx, sr, ns); err != nil {
			return false, buildCondition(featuresv1alpha1.StackReleaseError,
				corev1.ConditionTrue,
				fmt.Sprintf("failure creating service %v", err))
		}
		// Service in new namespace as ingress to new revision
		if err := r.defineRevisionIngressService(ctx, ns, sr); err != nil {
			return false, buildCondition(featuresv1alpha1.StackReleaseError,
				corev1.ConditionTrue,
				fmt.Sprintf("failure creating service %v", err))
		}
		requeue, cond := r.canaryController(ctx, sr)
		if requeue {
			if err := r.Client.Status().Update(ctx, sr); err != nil {
				return false, buildCondition(featuresv1alpha1.StackReleaseError,
					corev1.ConditionTrue,
					fmt.Sprintf("failed in canary controller %v", err))
			}
			return true, cond
		}
	case featuresv1alpha1.NoReleaseBackend:
		ns = sr.Spec.AppRevision.Config.AppNamespace
	default:
		return false, buildCondition(featuresv1alpha1.StackReleaseError,
			corev1.ConditionTrue,
			fmt.Sprintf("invalid backend type %q", sr.Spec.Config.BackendType))
	}
	// deploy appRevision into new namespace
	ar := &featuresv1alpha1.AppRevision{}
	ar.SetName(sr.Name)
	ar.SetNamespace(ns)
	if _, err := controllerutil.CreateOrUpdate(ctx, r.Client, ar, func() error {
		ar.Spec = sr.Spec.AppRevision
		return controllerutil.SetControllerReference(sr, ar, r.Scheme)
	}); err != nil {
		return false, buildCondition(featuresv1alpha1.StackReleaseError, corev1.ConditionTrue, fmt.Sprintf("failed to deploy AppRevision: %q", err))
	}
	sr.Status.CurrentRevision = featuresv1alpha1.CurrentRevisionState{
		Name:      ar.Name,
		Namespace: ar.Namespace,
		Revision:  ar.Spec.Revision,
	}
	for _, cond := range ar.Status.Conditions {
		if cond.Type == featuresv1alpha1.AppRevisionHealthy {
			sr.Status.CurrentRevision.Healthy = cond.Status
			break
		}
	}
	return false, buildCondition(featuresv1alpha1.StackReleaseInstalling, corev1.ConditionFalse, "")
}

func (r *StackReleaseReconciler) defineRevisonService(ctx context.Context, sr *featuresv1alpha1.StackRelease, newNs string) error {
	var svc corev1.Service
	svc.SetName(fmt.Sprintf("%s-r%v", sr.ObjectMeta.Name, sr.Spec.AppRevision.Revision))
	svc.SetNamespace(sr.Spec.Config.ProxyNamespace)
	_, err := controllerutil.CreateOrUpdate(ctx, r.Client, &svc, func() error {
		svc.Spec.Ports = []corev1.ServicePort{
			{
				Name:     "http",
				Port:     8080,
				Protocol: "TCP",
			},
		}
		svc.Spec.SessionAffinity = "None"
		svc.Spec.Type = "ExternalName"
		svc.Spec.ExternalName = fmt.Sprintf("%s.%s.svc.cluster.local", sr.Spec.AppName, newNs)
		return controllerutil.SetControllerReference(sr, &svc, r.Scheme)
	})
	return err
}

func (r *StackReleaseReconciler) defineRevisionIngressService(ctx context.Context, nsname string, sr *featuresv1alpha1.StackRelease) error {
	var svc corev1.Service
	svc.SetName(sr.Spec.AppName)
	svc.SetNamespace(nsname)
	_, err := controllerutil.CreateOrUpdate(ctx, r.Client, &svc, func() error {
		svc.Spec.Ports = []corev1.ServicePort{
			{
				Name:       "http",
				Port:       8080,
				TargetPort: intstr.IntOrString{IntVal: sr.Spec.Config.IngressPort},
				Protocol:   "TCP",
			},
		}
		svc.Spec.Selector = map[string]string{"stackreleaseingress": "true"}
		svc.Spec.SessionAffinity = "None"
		return controllerutil.SetControllerReference(sr, &svc, r.Scheme)
	})
	return err
}

func deployDirectory(ctx context.Context, c client.Client, ns string, dir string) error {
	items, err := ioutil.ReadDir(dir)
	if err != nil {
		return err
	}
	for _, item := range items {
		if err := deployYaml(ctx, c, ns, path.Join(dir, item.Name())); err != nil {
			return err
		}
	}
	return nil
}

func buildCondition(t featuresv1alpha1.StackReleaseConditionType, s corev1.ConditionStatus, m string) featuresv1alpha1.StackReleaseCondition {
	return featuresv1alpha1.StackReleaseCondition{
		Type:    t,
		Status:  s,
		Message: m,
	}
}

func (r *StackReleaseReconciler) setCondition(sr *featuresv1alpha1.StackRelease, src *featuresv1alpha1.StackReleaseCondition) {
	found := false
	for i, c := range sr.Status.Conditions {
		if c.Type == src.Type {
			src.LastTransitionTime = c.LastTransitionTime
			if c.Status != src.Status {
				src.LastTransitionTime = metav1.Now()
			}
			sr.Status.Conditions[i] = *src
			found = true
			break
		}
	}
	if !found {
		src.LastTransitionTime = metav1.Now()
		sr.Status.Conditions = append(sr.Status.Conditions, *src)
	}
	if sr.Status.State != featuresv1alpha1.StackReleaseStateRollback {
		sr.Status.State = featuresv1alpha1.StackReleaseStateReady
	}
	sr.Status.Reason = ""
	for _, c := range sr.Status.Conditions {
		if c.Status != corev1.ConditionTrue {
			continue
		}
		switch c.Type {
		case featuresv1alpha1.StackReleaseInstalling:
			sr.Status.State = featuresv1alpha1.StackReleaseStateCreating
		case featuresv1alpha1.StackReleaseDeploying:
			sr.Status.State = featuresv1alpha1.StackReleaseStateDeploying
		case featuresv1alpha1.StackReleaseError:
			sr.Status.State = featuresv1alpha1.StackReleaseStateError
			sr.Status.Reason = c.Message
			return
		}
	}
}
