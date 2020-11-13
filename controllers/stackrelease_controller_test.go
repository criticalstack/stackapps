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

var _ = Describe("StackReleaseController", func() {
	ctx := context.Background()
	timeout := 3 * time.Second
	interval := 1 * time.Second

	Context("with default backend type", func() {
		It("should create an AppRevision", func() {
			var sr featuresv1alpha1.StackRelease
			sr.SetName("my-app")
			sr.Spec.AppName = "the-app-name"
			sr.Spec.AppRevision = featuresv1alpha1.AppRevisionSpec{
				Revision:  1,
				Manifests: "test",
				Config: featuresv1alpha1.AppRevisionConfig{
					AppNamespace: ns.Name,
				},
			}

			Expect(k8sClient.Create(ctx, &sr)).Should(Succeed())
			defer func() {
				Expect(k8sClient.Delete(context.Background(), &sr)).Should(Succeed())
			}()

			Eventually(func() bool {
				err := k8sClient.Get(ctx, client.ObjectKey{Name: sr.Name}, &sr)
				Expect(err).NotTo(HaveOccurred(), "failed to get StackRelease")
				return sr.Status.State == featuresv1alpha1.StackReleaseStateReady
			}, timeout, interval).Should(BeTrue(), "status did not become ready")
			Expect(sr.Status.CurrentRevision.Name).Should(Equal(sr.Name))
			Expect(sr.Status.CurrentRevision.Namespace).Should(Equal(ns.Name))

			arKey := client.ObjectKey{
				Name:      sr.Status.CurrentRevision.Name,
				Namespace: sr.Status.CurrentRevision.Namespace,
			}
			var ar featuresv1alpha1.AppRevision
			Expect(k8sClient.Get(ctx, arKey, &ar)).Should(Succeed())
			defer func() {
				Expect(k8sClient.Delete(context.Background(), &ar)).Should(Succeed())
			}()
			Expect(ar.Spec).Should(Equal(sr.Spec.AppRevision))
			Expect(metav1.IsControlledBy(&ar, &sr.ObjectMeta)).To(BeTrue())
		})
	})
})
