package controllers

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"time"

	featuresv1alpha1 "github.com/criticalstack/stackapps/api/v1alpha1"
	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

// AppRevisionReconciler reconciles a AppRevision object
type AppRevisionReconciler struct {
	client.Client
	Log      logr.Logger
	Scheme   *runtime.Scheme
	recorder record.EventRecorder

	mu       sync.Mutex
	counters map[schema.GroupVersionKind]struct{}
	ctrl     controller.Controller
}

func (r *AppRevisionReconciler) SetupWithManager(mgr ctrl.Manager) error {
	c, err := ctrl.NewControllerManagedBy(mgr).
		For(&featuresv1alpha1.AppRevision{}).
		WithEventFilter(predicate.GenerationChangedPredicate{}).
		Build(r)
	if err != nil {
		return err
	}

	r.recorder = mgr.GetEventRecorderFor("apprevisions-controller")
	r.counters = make(map[schema.GroupVersionKind]struct{})
	r.ctrl = c
	return nil
}

func eventLogPredicate(log logr.Logger) predicate.Funcs {
	return predicate.Funcs{
		CreateFunc: func(ev event.CreateEvent) bool {
			log.Info("create event", "kind", ev.Object.GetObjectKind().GroupVersionKind().Kind, "name", ev.Meta.GetName())
			return true
		},
		DeleteFunc: func(ev event.DeleteEvent) bool {
			log.Info("delete event", "kind", ev.Object.GetObjectKind().GroupVersionKind().Kind, "name", ev.Meta.GetName())
			return true
		},
		UpdateFunc: func(ev event.UpdateEvent) bool {
			log.Info("update event", "kind", ev.ObjectOld.GetObjectKind().GroupVersionKind().Kind, "name", ev.MetaOld.GetName())
			return true
		},
		GenericFunc: func(ev event.GenericEvent) bool {
			log.Info("generic event", "kind", ev.Object.GetObjectKind().GroupVersionKind().Kind, "name", ev.Meta.GetName())
			return true
		},
	}
}

func (r *AppRevisionReconciler) createInformerIfNotExists(gvk schema.GroupVersionKind, since string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	var u unstructured.Unstructured
	u.SetGroupVersionKind(gvk)
	preds := make([]predicate.Predicate, 0)
	if since != "" {
		preds = append(preds, predicate.Funcs{CreateFunc: func(ev event.CreateEvent) bool {
			return ev.Meta.GetResourceVersion() > since
		}})
	}
	if os.Getenv("DEBUG_WATCH_EVENTS") != "" {
		preds = append(preds, eventLogPredicate(r.Log))
	}
	if _, ok := r.counters[gvk]; !ok {
		if err := r.ctrl.Watch(&source.Kind{Type: &u}, &handler.EnqueueRequestForOwner{
			OwnerType:    &featuresv1alpha1.AppRevision{},
			IsController: true,
		}, preds...); err != nil {
			return err
		}
		r.counters[gvk] = struct{}{}
	}
	return nil
}

func (r *AppRevisionReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	reqLogger := r.Log.WithValues("Request.Namespace", req.Namespace, "Request.Name", req.Name)
	reqLogger.Info("Reconciling AppRevision")

	// Fetch the AppRevision instance
	appRevision := &featuresv1alpha1.AppRevision{}

	if err := r.Get(ctx, req.NamespacedName, appRevision); err != nil {
		return reconcile.Result{}, client.IgnoreNotFound(err)
	}

	appRevision.Status.Default()

	var cm corev1.ConfigMap
	if err := r.Get(ctx, types.NamespacedName{Name: appRevision.Spec.Manifests, Namespace: appRevision.Spec.Config.AppNamespace}, &cm); err != nil {
		r.setErrorStatus(ctx, appRevision, string(apierrors.ReasonForError(err)), "Failed to get ConfigMap: "+err.Error())
		return reconcile.Result{}, err
	}

	if !appRevision.Spec.Config.Signing.Optional && len(appRevision.Spec.Signatures) == 0 {
		r.setErrorStatus(ctx, appRevision, "InvalidSigning", "Signatures are required but none are present")
		return reconcile.Result{}, nil
	}
	if !appRevision.Spec.Config.Signing.InsecureSkipVerification {
		for keyname, enc := range appRevision.Spec.Signatures {
			var vk featuresv1alpha1.VerificationKey
			if err := r.Client.Get(ctx, client.ObjectKey{Name: keyname}, &vk); err != nil {
				reason := "InvalidSigning"
				if apierrors.IsNotFound(errors.Cause(err)) {
					reason = "KeyNotFound"
				}
				r.setErrorStatus(ctx, appRevision, reason, "Failed to get VerificationKey: "+err.Error())
				return reconcile.Result{}, nil
			}
			if err := vk.VerifyConfigMapSignature(enc, &cm); err != nil {
				r.setErrorStatus(ctx, appRevision, "InvalidSigning", err.Error())
				return reconcile.Result{}, nil
			}
		}
	}

	type key struct {
		GVK  schema.GroupVersionKind
		Name string
	}
	oldResources := appRevision.Status.Resources
	inManifests := make(map[key]bool)
	currentResourceMap := make(map[key]bool)
	newResources := make(resourceList, 0)
	toCreate := make([]*unstructured.Unstructured, 0)
	toUpdate := make([]*unstructured.Unstructured, 0)
	for fname, b := range cm.Data {
		dec := yaml.NewYAMLOrJSONDecoder(strings.NewReader(b), 100)
		for {
			var u unstructured.Unstructured
			if err := dec.Decode(&u); err != nil {
				if err == io.EOF {
					break
				}
				r.setErrorStatus(ctx, appRevision, "InvalidManifest", "Failed to decode manifests: "+err.Error())
				return reconcile.Result{}, err
			}
			u.SetNamespace(req.Namespace)
			labels := u.GetLabels()
			if labels == nil {
				labels = make(map[string]string)
			}
			labels["stackapps.criticalstack.com/app-name"] = appRevision.Name
			u.SetLabels(labels)

			anno := u.GetAnnotations()
			if anno == nil {
				anno = make(map[string]string)
			}
			anno["stackapps.criticalstack.com/manifest-map"] = appRevision.Spec.Manifests
			anno["stackapps.criticalstack.com/manifest"] = fname
			u.SetAnnotations(anno)

			gvk := u.GroupVersionKind()
			inManifests[key{GVK: gvk, Name: u.GetName()}] = true

			if err := controllerutil.SetControllerReference(appRevision, &u, r.Scheme); err != nil {
				r.setErrorStatus(ctx, appRevision, "InvalidManifest", "Failed to set controller reference: "+err.Error())
				return reconcile.Result{}, nil
			}

			var found unstructured.Unstructured
			found.SetGroupVersionKind(gvk)

			if err := r.Client.Get(ctx, types.NamespacedName{Name: u.GetName(), Namespace: req.Namespace}, &found); err != nil {
				if apierrors.IsNotFound(err) {
					if appRevision.Spec.Config.DevMode && resourceList(appRevision.Status.OriginalResources).find(u.GetName(), gvk) != nil {
						// resource was deleted in dev mode, ignore
						continue
					}
					r.createInformerIfNotExists(gvk, "")
					toCreate = append(toCreate, &u)
					continue
				}
				r.setErrorStatus(ctx, appRevision, string(apierrors.ReasonForError(err)), "Failed checking for object: "+err.Error())
				return reconcile.Result{}, err
			}

			r.createInformerIfNotExists(gvk, found.GetResourceVersion())

			if ownerRef := metav1.GetControllerOf(&found); ownerRef != nil {
				if ownerRef.Kind != "AppRevision" || ownerRef.Name != appRevision.Name {
					// TODO(ktravis): should this be an error?
					reqLogger.Info("Resource exists and is managed elsewhere", "ownerRef", ownerRef)
					continue
				}
			}
			currentResourceMap[key{GVK: gvk, Name: found.GetName()}] = true

			// XXX(ktravis): evaluate if this is the right thing to check
			if found.GetAnnotations()["stackapps.criticalstack.com/manifest-map"] != appRevision.Spec.Manifests {
				toUpdate = append(toUpdate, &u)
				continue
			}
			newResources = append(newResources, featuresv1alpha1.AppRevisionResource{Unstructured: found})
		}
	}

	// Update to make sure we are at the latest version before setting status
	if err := r.Get(ctx, req.NamespacedName, appRevision); err != nil {
		return reconcile.Result{}, err
	}
	if len(toUpdate) > 0 {
		msg := fmt.Sprintf("Updating %d resources", len(toUpdate))
		r.setCondition(&appRevision.Status, featuresv1alpha1.AppRevisionCondition{
			Type:    featuresv1alpha1.AppRevisionUpdating,
			Status:  corev1.ConditionTrue,
			Reason:  "ManifestChange",
			Message: msg,
		})
		if len(toCreate) > 0 {
			msg += fmt.Sprintf(", creating %d new resources", len(toCreate))
		}
		r.setCondition(&appRevision.Status, featuresv1alpha1.AppRevisionCondition{
			Type:    featuresv1alpha1.AppRevisionReady,
			Status:  corev1.ConditionFalse,
			Reason:  "ManifestChange",
			Message: msg,
		})
	} else if len(toCreate) > 0 {
		r.setCondition(&appRevision.Status, featuresv1alpha1.AppRevisionCondition{
			Type:   featuresv1alpha1.AppRevisionUpdating,
			Status: corev1.ConditionFalse,
		})
		r.setCondition(&appRevision.Status, featuresv1alpha1.AppRevisionCondition{
			Type:    featuresv1alpha1.AppRevisionReady,
			Status:  corev1.ConditionFalse,
			Reason:  "ResourceCreation",
			Message: fmt.Sprintf("Creating %d new resources", len(toCreate)),
		})
	} else {
		r.setCondition(&appRevision.Status, featuresv1alpha1.AppRevisionCondition{
			Type:   featuresv1alpha1.AppRevisionUpdating,
			Status: corev1.ConditionFalse,
		})
		r.setCondition(&appRevision.Status, featuresv1alpha1.AppRevisionCondition{
			Type:   featuresv1alpha1.AppRevisionReady,
			Status: corev1.ConditionTrue,
		})
	}
	appRevision.Status.Resources = newResources
	r.aggregateResourceConditions(&appRevision.Status)
	health := featuresv1alpha1.AppRevisionCondition{
		Type:   featuresv1alpha1.AppRevisionHealthy,
		Status: corev1.ConditionUnknown,
	}
	for _, hc := range appRevision.Spec.HealthChecks {
		if err := checkHealth(hc, appRevision, &health); err != nil {
			reqLogger.Error(err, "failed to execute health check", "healthCheck", hc)
			health.Status = corev1.ConditionFalse
			health.Reason = "HealthCheck"
			if hc.Name != "" {
				health.Reason = hc.Name
			}
			health.Message = fmt.Sprintf("health check failed: %v", err)
			continue
		}
	}
	if health.Status == corev1.ConditionTrue {
		health.Message = "health check(s) passed"
	}
	r.setCondition(&appRevision.Status, health)

	errs := make([]string, 0)

	for _, res := range toCreate {
		r.recorder.Eventf(appRevision, corev1.EventTypeNormal, "AppRevisionReconcile", "Creating resource: %s/%s", res.GetKind(), res.GetName())
		if err := r.Client.Create(ctx, res); err != nil {
			errs = append(errs, errors.Wrapf(err, "object creation failed").Error())
			break
		}
		r.addObjectReference(appRevision, res)
	}

	for _, res := range toUpdate {
		r.recorder.Eventf(appRevision, corev1.EventTypeNormal, "AppRevisionReconcile", "Updating resource: %s/%s", res.GetKind(), res.GetName())
		force := true
		if err := r.Patch(ctx, res, client.Apply, &client.PatchOptions{FieldManager: "apprevisions-controller", Force: &force}); err != nil {
			errs = append(errs, errors.Wrapf(err, "object patch failed").Error())
			break
		}
		r.addObjectReference(appRevision, res)
	}

	cond := featuresv1alpha1.AppRevisionCondition{
		Type:   featuresv1alpha1.AppRevisionDeploymentFailed,
		Status: corev1.ConditionFalse,
	}
	if len(errs) > 0 {
		cond.Status = corev1.ConditionTrue
		cond.Reason = "Error"
		cond.Message = strings.Join(errs, ", ")

		r.recorder.Eventf(appRevision, corev1.EventTypeWarning, "AppRevisionReconcile", cond.Message)
	}

	for _, res := range oldResources {
		if !inManifests[key{GVK: res.GroupVersionKind(), Name: res.GetName()}] {
			r.removeObjectReference(appRevision, &res)
		}
		if currentResourceMap[key{GVK: res.GroupVersionKind(), Name: res.GetName()}] {
			continue
		}

		reqLogger.Info("Orphaned resource", "kind", res.GetKind(), "name", res.GetName())
		// TODO(ktravis): check AppRevision.Spec.OrphanPolicy
		r.recorder.Eventf(appRevision, corev1.EventTypeNormal, "AppRevisionReconcile", "Deleting orphaned resource: %v/%v", res.GetKind(), res.GetName())
		if err := client.IgnoreNotFound(r.Client.Delete(ctx, &res)); err != nil {
			r.recorder.Eventf(appRevision, corev1.EventTypeWarning, "AppRevisionReconcile", "Failed to delete %v/%v: %v", res.GetKind(), res.GetName(), err)
			reqLogger.Error(err, "Failed delete", "kind", res.GetKind(), "name", res.GetName())
			errs = append(errs, errors.Wrapf(err, "object delete failed").Error())
			continue
		}
	}

	r.setCondition(&appRevision.Status, cond)
	if err := r.Status().Update(ctx, appRevision); err != nil {
		return reconcile.Result{}, errors.Wrap(err, "Failed to update AppRevision status")
	}

	if len(errs) > 0 {
		return reconcile.Result{RequeueAfter: time.Second * 5}, nil
	}

	return ctrl.Result{}, nil
}

func (r *AppRevisionReconciler) aggregateResourceConditions(status *featuresv1alpha1.AppRevisionStatus) {
	conditionsByType := make(map[featuresv1alpha1.ResourceConditionType]*featuresv1alpha1.ResourceCondition)
	for _, res := range status.Resources {
		conds, ok, _ := unstructured.NestedSlice(res.Object, "status", "conditions")
		if !ok {
			continue
		}
		apiGroup := res.GroupVersionKind().Group
		localRef := corev1.TypedLocalObjectReference{
			APIGroup: &apiGroup,
			Kind:     res.GetKind(),
			Name:     res.GetName(),
		}
		for _, cond := range conds {
			m, ok := cond.(map[string]interface{})
			if !ok {
				continue
			}
			t, ok := m["type"].(string)
			if !ok {
				continue
			}
			s, ok := m["status"].(string)
			if !ok {
				continue
			}
			st := corev1.ConditionStatus(s)
			rct := featuresv1alpha1.ResourceConditionType(t)
			rc, ok := conditionsByType[rct]
			if !ok {
				rc = &featuresv1alpha1.ResourceCondition{
					Type:   rct,
					Status: st,
				}
				conditionsByType[rct] = rc
			}
			if st == corev1.ConditionFalse {
				rc.Status = st
			}
			rci := featuresv1alpha1.ResourceConditionInstance{
				Status:   st,
				Resource: localRef,
			}
			rci.Reason, _ = m["reason"].(string)
			rc.Instances = append(rc.Instances, rci)
		}
	}
	status.ResourceConditions = make([]featuresv1alpha1.ResourceCondition, 0)
	for _, cond := range conditionsByType {
		status.ResourceConditions = append(status.ResourceConditions, *cond)
	}
}

func (r *AppRevisionReconciler) setCondition(status *featuresv1alpha1.AppRevisionStatus, cond featuresv1alpha1.AppRevisionCondition) (prev *featuresv1alpha1.AppRevisionCondition) {
	for i, c := range status.Conditions {
		c := c
		if c.Type == cond.Type {
			prev = &c
			if prev.Status != cond.Status {
				cond.LastTransitionTime = metav1.Now()
			} else {
				cond.LastTransitionTime = prev.LastTransitionTime
			}
			status.Conditions[i] = cond
			break
		}
	}
	if prev == nil {
		cond.LastTransitionTime = metav1.Now()
		status.Conditions = append(status.Conditions, cond)
	}
	return prev
}

func (r *AppRevisionReconciler) setErrorStatus(ctx context.Context, appRevision *featuresv1alpha1.AppRevision, reason, msg string) error {
	r.recorder.Eventf(appRevision, corev1.EventTypeWarning, reason, msg)
	r.setCondition(&appRevision.Status, featuresv1alpha1.AppRevisionCondition{
		Type:    featuresv1alpha1.AppRevisionDeploymentFailed,
		Status:  corev1.ConditionTrue,
		Reason:  reason,
		Message: msg,
	})
	r.setCondition(&appRevision.Status, featuresv1alpha1.AppRevisionCondition{
		Type:    featuresv1alpha1.AppRevisionReady,
		Status:  corev1.ConditionFalse,
		Reason:  reason,
		Message: msg,
	})
	if err := r.Status().Update(ctx, appRevision); err != nil {
		r.Log.Error(err, "updating status failed")
		return err
	}
	return nil
}

type resourceList []featuresv1alpha1.AppRevisionResource

func (r resourceList) find(name string, gvk schema.GroupVersionKind) *featuresv1alpha1.AppRevisionResource {
	apiVersion, kind := gvk.ToAPIVersionAndKind()
	for i, ref := range r {
		if ref.GetName() == name && ref.GetAPIVersion() == apiVersion && ref.GetKind() == kind {
			return &r[i]
		}
	}
	return nil
}

// TODO(ktravis): this was originally storing references rather than entire objects, revisit later
func (r *AppRevisionReconciler) addObjectReference(ar *featuresv1alpha1.AppRevision, res *unstructured.Unstructured) {
	for i, ref := range ar.Status.OriginalResources {
		if !(ref.GetName() == res.GetName() && ref.GetAPIVersion() == res.GetAPIVersion() && ref.GetKind() == res.GetKind()) {
			continue
		}
		ar.Status.OriginalResources[i] = featuresv1alpha1.AppRevisionResource{Unstructured: *res}
		return
	}
	ar.Status.OriginalResources = append(ar.Status.OriginalResources, featuresv1alpha1.AppRevisionResource{Unstructured: *res})
}

func (r *AppRevisionReconciler) removeObjectReference(ar *featuresv1alpha1.AppRevision, res *featuresv1alpha1.AppRevisionResource) {
	for i := range ar.Status.OriginalResources {
		ref := ar.Status.OriginalResources[i]
		if ref.GetName() == res.GetName() && ref.GetAPIVersion() == res.GetAPIVersion() && ref.GetKind() == res.GetKind() {
			ar.Status.OriginalResources = append(ar.Status.OriginalResources[:i], ar.Status.OriginalResources[i+1:]...)
			return
		}
	}
}
