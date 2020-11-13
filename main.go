package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/prometheus/client_golang/api"
	promv1 "github.com/prometheus/client_golang/api/prometheus/v1"

	"go.uber.org/zap/zapcore"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	featuresv1alpha1 "github.com/criticalstack/stackapps/api/v1alpha1"
	"github.com/criticalstack/stackapps/controllers"
	// +kubebuilder:scaffold:imports
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

const (
	DefaultPrometheusEndpoint = "http://cs-prometheus-server.critical-stack.svc.cluster.local:9090"
)

func init() {
	_ = clientgoscheme.AddToScheme(scheme)

	_ = featuresv1alpha1.AddToScheme(scheme)
	// +kubebuilder:scaffold:scheme
}

func main() {
	var metricsAddr string
	var enableLeaderElection bool
	flag.StringVar(&metricsAddr, "metrics-addr", ":8080", "The address the metric endpoint binds to.")
	flag.BoolVar(&enableLeaderElection, "enable-leader-election", false,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")
	flag.Parse()

	promAddr := os.Getenv("PROMETHEUS_ENDPOINT")
	var config api.Config
	if promAddr != "" {
		config.Address = promAddr
	} else {
		config.Address = DefaultPrometheusEndpoint
	}
	client, err := api.NewClient(config)
	if err != nil {
		setupLog.Error(err, "unable to start prometheus client")
		os.Exit(1)
	}

	ctrl.SetLogger(zap.New(zap.StacktraceLevel(zapcore.FatalLevel), zap.UseDevMode(true)))

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:             scheme,
		MetricsBindAddress: metricsAddr,
		Port:               9443,
		LeaderElection:     enableLeaderElection,
		LeaderElectionID:   "7830725e.criticalstack.com",
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	reconcilers := []interface {
		SetupWithManager(ctrl.Manager) error
	}{
		&controllers.StackAppReconciler{
			Client: mgr.GetClient(),
			Log:    ctrl.Log.WithName("controllers").WithName("StackApps"),
			Scheme: mgr.GetScheme(),
		},
		&controllers.AppRevisionReconciler{
			Client: mgr.GetClient(),
			Log:    ctrl.Log.WithName("controllers").WithName("AppRevisions"),
			Scheme: mgr.GetScheme(),
		},
		&controllers.StackValueReconciler{
			Client: mgr.GetClient(),
			Log:    ctrl.Log.WithName("controllers").WithName("StackValues"),
			Scheme: mgr.GetScheme(),
		},
		&controllers.StackReleaseReconciler{
			Client:  mgr.GetClient(),
			Log:     ctrl.Log.WithName("controllers").WithName("StackReleases"),
			Scheme:  mgr.GetScheme(),
			Metrics: promv1.NewAPI(client),
		},
	}
	for _, r := range reconcilers {
		if err := r.SetupWithManager(mgr); err != nil {
			setupLog.Error(err, "unable to create controller", "controller", "StackApps")
			os.Exit(1)
		}
	}

	if os.Getenv("ENABLE_WEBHOOKS") != "false" {
		whTypes := []interface {
			SetupWebhookWithManager(ctrl.Manager) error
		}{
			&featuresv1alpha1.StackApp{},
			&featuresv1alpha1.AppRevision{},
			&featuresv1alpha1.VerificationKey{},
		}
		for _, t := range whTypes {
			if err = t.SetupWebhookWithManager(mgr); err != nil {
				setupLog.Error(err, "unable to create webhook", "webhook", fmt.Sprintf("%T", t))
				os.Exit(1)
			}
		}
	}

	setupLog.Info("Starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "Manager exited non-zero")
		os.Exit(1)
	}
}
