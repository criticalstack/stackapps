package controllers

import (
	"context"

	featuresv1alpha1 "github.com/criticalstack/stackapps/api/v1alpha1"
	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// StackAppReconciler reconciles a StackApp object
type StackAppReconciler struct {
	client.Client
	Log      logr.Logger
	Scheme   *runtime.Scheme
	recorder record.EventRecorder
}

func (r *StackAppReconciler) SetupWithManager(mgr ctrl.Manager) error {
	err := ctrl.NewControllerManagedBy(mgr).
		For(&featuresv1alpha1.StackApp{}).
		Owns(&featuresv1alpha1.StackRelease{}).
		Complete(r)
	if err != nil {
		return err
	}

	r.recorder = mgr.GetEventRecorderFor("stackapps-controller")
	return nil
}

func (r *StackAppReconciler) Reconcile(request ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	reqLogger := r.Log.WithValues("stackapp", request.Name)

	reqLogger.Info("Reconciling StackApp")
	stackApp := &featuresv1alpha1.StackApp{}
	if err := r.Client.Get(ctx, request.NamespacedName, stackApp); err != nil {
		return reconcile.Result{}, client.IgnoreNotFound(err)
	}
	config := featuresv1alpha1.StackAppConfig{}
	if err := r.Get(ctx, client.ObjectKey{Name: stackApp.Name}, &config); err != nil {
		if err := r.setErrorStatus(ctx, stackApp, "StackAppConfig", err); err != nil {
			reqLogger.Error(err, "failed to update status")
		}
		return reconcile.Result{}, err
	}
	var curr featuresv1alpha1.StackRelease
	if err := r.Client.Get(ctx, client.ObjectKey{Name: stackApp.Name}, &curr); err != nil {
		if !apierrors.IsNotFound(err) {
			return reconcile.Result{}, err
		}
	}
	//if stackRelease has issued rollback, revert to previous stackapp
	if curr.Status.State == featuresv1alpha1.StackReleaseStateRollback {
		stackApp.Spec = config.Spec.Releases.RollbackRevision
		if err := r.Client.Update(ctx, stackApp); err != nil {
			return reconcile.Result{}, err
		}
	}

	stackApp.Status.CurrentRelease = featuresv1alpha1.CurrentReleaseState{
		Name:               curr.Name,
		Namespace:          curr.Namespace,
		StackReleaseStatus: curr.Status,
	}
	if err := r.Status().Update(ctx, stackApp); err != nil {

		return ctrl.Result{}, err
	}
	sr := &featuresv1alpha1.StackRelease{}
	sr.SetName(stackApp.Name)
	operationalResult, err := controllerutil.CreateOrUpdate(ctx, r.Client, sr, func() error {
		sr.Spec.AppName = stackApp.Name
		sr.Spec.Config = config.Spec.Releases
		sr.Spec.Config.ProxyNamespace = config.Spec.AppNamespace
		sr.Spec.AppRevision = stackApp.Spec.AppRevision
		sr.Spec.AppRevision.Config = config.Spec.AppRevisions
		sr.Spec.AppRevision.Config.AppNamespace = config.Spec.AppNamespace
		return controllerutil.SetControllerReference(stackApp, sr, r.Scheme)
	})
	if err != nil {
		if err := r.setErrorStatus(ctx, stackApp, "StackRelease", err); err != nil {
			reqLogger.Error(err, "failed to update status")
		}
		return reconcile.Result{}, err
	}
	if operationalResult != controllerutil.OperationResultNone {
		stackApp.Status.State = featuresv1alpha1.StackAppStateCreating
		if err := r.Status().Update(ctx, stackApp); err != nil {
			return ctrl.Result{}, err
		}
	}
	//if everything is healthy store the current stackApp in the configuration
	//for rollback and recovery
	if curr.Status.State == featuresv1alpha1.StackReleaseStateReady {
		if stackApp.Status.State == featuresv1alpha1.StackAppStateReady {
			if curr.Status.CurrentRevision.Healthy == corev1.ConditionTrue {
				config.Spec.Releases.RollbackRevision = stackApp.Spec
				reqLogger.Info("updating config.RollbackRevision")
				if err := r.Client.Update(context.Background(), &config); err != nil {
					reqLogger.Error(err, "failed to update config with rollBackVersion")
					return reconcile.Result{}, err
				}
			}
		}
	}
	stackApp.Status = featuresv1alpha1.StackAppStatus{
		CurrentRelease: featuresv1alpha1.CurrentReleaseState{
			Name:               sr.Name,
			StackReleaseStatus: sr.Status,
		},
		State: featuresv1alpha1.StackAppStateReady,
	}
	if err := r.Status().Update(ctx, stackApp); err != nil {
		return ctrl.Result{}, err
	}
	return ctrl.Result{}, nil
}

func (r *StackAppReconciler) setErrorStatus(ctx context.Context, sa *featuresv1alpha1.StackApp, reason string, err error) error {
	sa.Status.State = featuresv1alpha1.StackAppStateError
	sa.Status.Reason = reason
	sa.Status.Message = err.Error()
	return r.Status().Update(ctx, sa)
}
