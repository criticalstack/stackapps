package controllers

import (
	"context"
	"time"

	featuresv1alpha1 "github.com/criticalstack/stackapps/api/v1alpha1"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("StackAppController", func() {
	ctx := context.Background()
	timeout := 3 * time.Second
	interval := 1 * time.Second

	Context("when there is no StackAppConfig", func() {
		It("should set an error status", func() {
			var sa featuresv1alpha1.StackApp
			sa.SetName("my-stackapp")
			Expect(k8sClient.Create(ctx, &sa)).Should(Succeed())
			defer func() {
				Expect(k8sClient.Delete(context.Background(), &sa)).Should(Succeed())
			}()
			Eventually(func() bool {
				found := &featuresv1alpha1.StackApp{}
				err := k8sClient.Get(ctx, client.ObjectKey{Name: sa.Name}, found)
				Expect(err).NotTo(HaveOccurred(), "failed to get StackApp")
				return found.Status.State == featuresv1alpha1.StackAppStateError &&
					found.Status.Reason == "StackAppConfig"
			}, timeout, interval).Should(BeTrue())
		})
	})

	Context("with a normal StackAppConfig", func() {
		It("should create a StackRelease", func() {
			var config featuresv1alpha1.StackAppConfig
			config.SetName("my-stackapp")
			config.Spec.AppNamespace = ns.Name
			config.Spec.Releases.Enabled = true

			Expect(k8sClient.Create(ctx, &config)).Should(Succeed())
			defer func() {
				Expect(k8sClient.Delete(context.Background(), &config)).Should(Succeed())
			}()

			var sa featuresv1alpha1.StackApp
			sa.SetName("my-stackapp")
			sa.Spec.AppRevision = featuresv1alpha1.AppRevisionSpec{
				Revision:  1,
				Manifests: "blah",
			}
			Expect(k8sClient.Create(ctx, &sa)).Should(Succeed())
			defer func() {
				Expect(k8sClient.Delete(context.Background(), &sa)).Should(Succeed())
			}()

			By("becoming ready")
			found := &featuresv1alpha1.StackApp{}
			Eventually(func() bool {
				err := k8sClient.Get(ctx, client.ObjectKey{Name: sa.Name}, found)
				Expect(err).NotTo(HaveOccurred(), "failed to get StackApp")
				return found.Status.State == featuresv1alpha1.StackAppStateReady
			}, timeout, interval).Should(BeTrue())
			Expect(found.Status.CurrentRelease.Name).Should(Equal(sa.Name))

			By("creating a StackRelease")
			srKey := client.ObjectKey{
				Name: found.Status.CurrentRelease.Name,
			}
			sr := &featuresv1alpha1.StackRelease{}
			Expect(k8sClient.Get(ctx, srKey, sr)).Should(Succeed())
			Expect(sr.Namespace).To(BeZero(), "StackRelease is not cluster-scoped")
			Expect(sr.Spec.Config.ProxyNamespace).ToNot(BeEmpty())

			By("owning the created StackRelease")
			Expect(metav1.IsControlledBy(sr, &sa.ObjectMeta)).To(BeTrue())
		})

		It("should update an existing StackRelease when the revision changes", func() {
			var config featuresv1alpha1.StackAppConfig
			config.SetName("my-stackapp")
			config.Spec.AppNamespace = ns.Name
			config.Spec.Releases.Enabled = true

			Expect(k8sClient.Create(ctx, &config)).Should(Succeed())
			defer func() {
				Expect(k8sClient.Delete(context.Background(), &config)).Should(Succeed())
			}()

			var sa featuresv1alpha1.StackApp
			sa.SetName("my-stackapp")
			sa.Spec.AppRevision = featuresv1alpha1.AppRevisionSpec{
				Revision:  1,
				Manifests: "blah",
			}
			Expect(k8sClient.Create(ctx, &sa)).Should(Succeed())
			defer func() {
				Expect(k8sClient.Delete(context.Background(), &sa)).Should(Succeed())
			}()

			Eventually(func() bool {
				err := k8sClient.Get(ctx, client.ObjectKey{Name: sa.Name}, &sa)
				Expect(err).NotTo(HaveOccurred(), "failed to get StackApp")
				return sa.Status.State == featuresv1alpha1.StackAppStateReady
			}, timeout, interval).Should(BeTrue())

			sr := &featuresv1alpha1.StackRelease{}
			Expect(k8sClient.Get(ctx, client.ObjectKey{Name: sa.Status.CurrentRelease.Name}, sr)).Should(Succeed())
			Expect(sr.Spec.AppRevision.Revision).To(Equal(uint64(1)))

			Eventually(func() error {
				err := k8sClient.Get(ctx, client.ObjectKey{Name: sa.Name}, &sa)
				Expect(err).NotTo(HaveOccurred(), "failed to get StackApp")
				sa.Spec.AppRevision = featuresv1alpha1.AppRevisionSpec{
					Revision:  2,
					Manifests: "blah-2",
				}
				return k8sClient.Update(ctx, &sa)
			}, timeout, interval).Should(Succeed())

			Eventually(func() uint64 {
				Expect(k8sClient.Get(ctx, client.ObjectKey{Name: sa.Status.CurrentRelease.Name}, sr)).Should(Succeed())
				return sr.Spec.AppRevision.Revision
			}, timeout, time.Second).Should(Equal(uint64(2)))
		})
	})
})
