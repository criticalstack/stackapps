package controllers

import (
	"context"
	"math/rand"

	"os"
	"path/filepath"
	"testing"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	"sigs.k8s.io/controller-runtime/pkg/envtest/printer"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	featuresv1alpha1 "github.com/criticalstack/stackapps/api/v1alpha1"
	// +kubebuilder:scaffold:imports
)

// These tests use Ginkgo (BDD-style Go testing framework). Refer to
// http://onsi.github.io/ginkgo/ to learn more about Ginkgo.

var cfg *rest.Config
var k8sClient client.Client
var k8sManager ctrl.Manager
var testEnv *envtest.Environment
var finished = make(chan struct{})
var ns corev1.Namespace
var cleanup []*corev1.Namespace

func TestAPIs(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecsWithDefaultAndCustomReporters(t,
		"Controller Suite",
		[]Reporter{printer.NewlineReporter{}})
}

func randString(n int) string {
	const chars = "abcdefghijklmnopqrstuvwxyz"
	b := make([]byte, n)
	for i := 0; i < n; i++ {
		b[i] = chars[rand.Intn(len(chars))]
	}
	return string(b)
}

var _ = BeforeEach(func() {
	ns.SetName("test-ns-" + randString(8))
	cp := ns.DeepCopy()
	Expect(k8sClient.Create(context.Background(), cp)).Should(Succeed())
	cleanup = append(cleanup, cp)
})

var _ = AfterEach(func() {
	// Expect(k8sClient.Delete(context.Background(), ns.DeepCopy())).Should(Succeed())
})

var _ = BeforeSuite(func(done Done) {
	logf.SetLogger(zap.LoggerTo(GinkgoWriter, true))

	if os.Getenv("TEST_USE_EXISTING_CLUSTER") == "true" {
		t := true
		testEnv = &envtest.Environment{
			UseExistingCluster: &t,
		}
	} else {
		testEnv = &envtest.Environment{
			CRDDirectoryPaths: []string{filepath.Join("..", "chart", "crds")},
		}
	}

	var err error
	cfg, err = testEnv.Start()
	Expect(err).ToNot(HaveOccurred())
	Expect(cfg).ToNot(BeNil())

	err = featuresv1alpha1.AddToScheme(scheme.Scheme)
	Expect(err).NotTo(HaveOccurred())

	// +kubebuilder:scaffold:scheme

	k8sManager, err = ctrl.NewManager(cfg, ctrl.Options{
		Scheme: scheme.Scheme,
	})
	Expect(err).ToNot(HaveOccurred())

	recs := []interface {
		SetupWithManager(ctrl.Manager) error
	}{
		&StackAppReconciler{
			Client: k8sManager.GetClient(),
			Log:    ctrl.Log.WithName("controllers").WithName("StackApp"),
			Scheme: scheme.Scheme,
		},
		&StackReleaseReconciler{
			Client: k8sManager.GetClient(),
			Log:    ctrl.Log.WithName("controllers").WithName("StackRelease"),
			Scheme: scheme.Scheme,
		},
		&AppRevisionReconciler{
			Client: k8sManager.GetClient(),
			Log:    ctrl.Log.WithName("controllers").WithName("AppRevision"),
			Scheme: scheme.Scheme,
		},
		&StackValueReconciler{
			Client: k8sManager.GetClient(),
			Log:    ctrl.Log.WithName("controllers").WithName("StackValue"),
			Scheme: scheme.Scheme,
		},
	}

	for _, r := range recs {
		Expect(r.SetupWithManager(k8sManager)).To(Succeed())
	}

	go func() {
		<-ctrl.SetupSignalHandler()
		close(finished)
	}()

	go func() {
		err := k8sManager.Start(finished)
		Expect(err).ToNot(HaveOccurred())
		gexec.KillAndWait(5 * time.Second)
		err = testEnv.Stop()
		Expect(err).ToNot(HaveOccurred())
	}()

	k8sClient, err = client.New(cfg, client.Options{Scheme: scheme.Scheme})
	Expect(err).ToNot(HaveOccurred())
	Expect(k8sClient).ToNot(BeNil())

	close(done)
}, 60)

var _ = AfterSuite(func() {
	By("tearing down the test environment")
	for _, x := range cleanup {
		Expect(k8sClient.Delete(context.Background(), x)).Should(Succeed())
	}
	close(finished)
})
