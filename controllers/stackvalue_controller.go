package controllers

import (
	"context"
	"fmt"

	featuresv1alpha1 "github.com/criticalstack/stackapps/api/v1alpha1"
	provider "github.com/criticalstack/stackapps/controllers/provider"
	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	controllers "sigs.k8s.io/controller-runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// StackValueReconciler reconciles a StackValue object
type StackValueReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

func (r *StackValueReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return controllers.NewControllerManagedBy(mgr).
		For(&featuresv1alpha1.StackValue{}).
		Owns(&corev1.ConfigMap{}).
		Owns(&corev1.Secret{}).
		Complete(r)
}

// TODO(ktravis): remove me when not needed
func getConfig(ctx context.Context, c client.Client, app string) (*featuresv1alpha1.StackAppConfig, error) {
	config := featuresv1alpha1.StackAppConfig{}
	if err := c.Get(ctx, client.ObjectKey{Name: app}, &config); err != nil {
		if apierrors.IsNotFound(err) {
			return nil, errors.Wrap(err, fmt.Sprintf("Cannot find stackapps config %q", app))
		}
		return nil, err
	}
	return &config, nil
}

func (r *StackValueReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	reqLogger := r.Log.WithValues("Request.Namespace", req.Namespace, "Request.Name", req.Name)
	reqLogger.Info("Reconciling StackValue")

	ctx := context.Background()
	stackValue := &featuresv1alpha1.StackValue{}
	// try to retireve stackValues details
	if err := r.Client.Get(ctx, req.NamespacedName, stackValue); err != nil {
		return reconcile.Result{}, client.IgnoreNotFound(err)
	}
	config, err := getConfig(ctx, r.Client, stackValue.Spec.AppName)
	if err != nil {
		r.setErrorStatus(ctx, stackValue, "GetStackAppConfig", err.Error())
		return reconcile.Result{}, err
	}

	if !config.Spec.StackValues.Enabled {
		reqLogger.Info("StackValues are not enabled, enable in StackAppsConfig resource")
		r.setErrorStatus(ctx, stackValue, "StackValues not enabled", "")
		return reconcile.Result{}, nil
	}
	var secret corev1.Secret
	secretKey := client.ObjectKey{
		Name:      config.Spec.StackValues.Secret.Name,
		Namespace: config.Spec.StackValues.Secret.Namespace,
	}
	if secretKey.Namespace == "" {
		secretKey.Namespace = config.Spec.AppNamespace
	}

	if err := r.Get(ctx, secretKey, &secret); err != nil {
		if apierrors.IsNotFound(err) {
			err = errors.Wrap(err, fmt.Sprintf("Cannot find secret containing access tokens %q", secretKey))
		}
		r.setErrorStatus(ctx, stackValue, "GetCredentials", err.Error())
		return reconcile.Result{}, err
	}
	mapTokens(secret.Data, config)

	var u unstructured.Unstructured
	u.SetGroupVersionKind(schema.GroupVersionKind{
		Version: "v1",
		Kind:    stackValue.Spec.ObjectType,
	})
	// try to get object to ensure it doesnt exist
	if err := r.Get(ctx, types.NamespacedName{Name: stackValue.ObjectMeta.Name, Namespace: req.Namespace}, &u); err != nil {
		if !apierrors.IsNotFound(err) {
			r.setErrorStatus(ctx, stackValue, "GetObject", err.Error())
			return reconcile.Result{}, err
		}
		src := config.Spec.StackValues.Source(stackValue.Spec.SourceType)
		if src == nil {
			r.setErrorStatus(ctx, stackValue, "GetSource", fmt.Sprintf("source not found %q", stackValue.Spec.SourceType))
			return reconcile.Result{}, err
		}
		// object doesn't exist, try to create it
		p := provider.New(src, stackValue.Spec.Path)
		values, err := p.Values()
		if err != nil {
			r.setErrorStatus(ctx, stackValue, "FetchValues", err.Error())
			return reconcile.Result{}, err
		}
		u.SetName(stackValue.Name)
		u.SetNamespace(stackValue.Namespace)
		u.Object["data"] = values
		if err := controllerutil.SetControllerReference(stackValue, &u, r.Scheme); err != nil {
			r.setErrorStatus(ctx, stackValue, "CreateObject", err.Error())
			return reconcile.Result{}, errors.Wrap(err, "StackValues controller failed to set controller reference")
		}

		if err := r.Create(ctx, &u); err != nil {
			r.setErrorStatus(ctx, stackValue, "CreateObject", err.Error())
			return reconcile.Result{}, errors.Wrapf(err, "stackValues Controller failed to create object %q", stackValue.Spec.ObjectType)
		}
		return ctrl.Result{}, nil
	}

	r.setCondition(&stackValue.Status, featuresv1alpha1.StackValueCondition{
		Type:   featuresv1alpha1.StackValueFailed,
		Status: corev1.ConditionFalse,
	})
	r.setCondition(&stackValue.Status, featuresv1alpha1.StackValueCondition{
		Type:   featuresv1alpha1.StackValueReady,
		Status: corev1.ConditionTrue,
	})
	if err := r.Status().Update(ctx, stackValue); err != nil {
		return ctrl.Result{}, err
	}
	return ctrl.Result{}, nil
}

func mapTokens(a map[string][]byte, c *featuresv1alpha1.StackAppConfig) {
	for _, src := range c.Spec.StackValues.Sources {
		for tokenName, token := range a {
			if tokenName == src.Name {
				src.Token = token
				break
			}
		}
	}
}

func (r *StackValueReconciler) setCondition(status *featuresv1alpha1.StackValueStatus, cond featuresv1alpha1.StackValueCondition) (prev *featuresv1alpha1.StackValueCondition) {
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

func (r *StackValueReconciler) setErrorStatus(ctx context.Context, stackValue *featuresv1alpha1.StackValue, reason, msg string) {
	r.setCondition(&stackValue.Status, featuresv1alpha1.StackValueCondition{
		Type:    featuresv1alpha1.StackValueFailed,
		Status:  corev1.ConditionTrue,
		Reason:  reason,
		Message: msg,
	})
	r.setCondition(&stackValue.Status, featuresv1alpha1.StackValueCondition{
		Type:   featuresv1alpha1.StackValueReady,
		Status: corev1.ConditionFalse,
		Reason: "Failed",
	})
	r.Status().Update(ctx, stackValue)
}
