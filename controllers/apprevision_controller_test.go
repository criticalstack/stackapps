package controllers

import (
	"bytes"
	"context"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"time"

	featuresv1alpha1 "github.com/criticalstack/stackapps/api/v1alpha1"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/types"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ktypes "k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("AppRevisionController", func() {
	ctx := context.Background()
	timeout := 10 * time.Second
	interval := 1 * time.Second

	reportSuccess := ContainElements(
		SatisfyAll(
			haveConditionType(featuresv1alpha1.AppRevisionDeploymentFailed),
			haveConditionStatus(corev1.ConditionFalse),
		),
		SatisfyAll(
			haveConditionType(featuresv1alpha1.AppRevisionReady),
			haveConditionStatus(corev1.ConditionTrue),
		),
	)

	failDeploymentWithReason := func(r string) types.GomegaMatcher {
		return ContainElement(SatisfyAll(
			haveConditionType(featuresv1alpha1.AppRevisionDeploymentFailed),
			haveConditionStatus(corev1.ConditionTrue),
			haveReason(r),
		))
	}

	var cm corev1.ConfigMap
	cm.SetName("app-r1")
	manifests := `apiVersion: apps/v1
kind: Deployment
metadata:
  name: my-app
spec:
  selector:
    matchLabels:
      app: my-app
  template:
    metadata:
      labels:
        app: my-app
    spec:
      containers:
      - name: hi
        image: busybox
`
	cm.Data = map[string]string{
		"manifests": manifests,
	}

	var ar *featuresv1alpha1.AppRevision

	config := featuresv1alpha1.AppRevisionConfig{}

	shouldFailWithReason := func(r string) (string, func()) {
		return "fails to deploy", func() {
			Eventually(getAppRevisionConditions(ctx, ar), timeout, interval).Should(failDeploymentWithReason(r))
		}
	}
	shouldDeploySuccessfully := func() (string, func()) {
		return "deploys successfully", func() {
			Eventually(getAppRevisionConditions(ctx, ar), timeout, interval).Should(reportSuccess)
			var d appsv1.Deployment
			Expect(k8sClient.Get(ctx, client.ObjectKey{Name: "my-app", Namespace: ar.Namespace}, &d)).Should(Succeed())

			Expect(d.Spec.Template.Spec.Containers).To(ConsistOf(corev1.Container{
				Name:                     "hi",
				Image:                    "busybox",
				TerminationMessagePath:   "/dev/termination-log",
				TerminationMessagePolicy: "File",
				ImagePullPolicy:          "Always",
			}))
			Expect(metav1.IsControlledBy(&d, &ar.ObjectMeta)).To(BeTrue(), "AppRevision does not control deployment")
		}
	}
	var vk *featuresv1alpha1.VerificationKey
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		panic(err)
	}
	// Data: map[string][]byte{"privatekey":
	b := &pem.Block{
		Type:  "RSA PUBLIC KEY",
		Bytes: x509.MarshalPKCS1PublicKey(&key.PublicKey),
	}
	buf := new(bytes.Buffer)
	if err := pem.Encode(buf, b); err != nil {
		panic(err)
	}

	Context("apprevision is deployed", func() {
		BeforeEach(func() {
			ar = &featuresv1alpha1.AppRevision{}
			ar.SetName("my-app")
			ar.SetNamespace(ns.Name)
			ar.Spec = featuresv1alpha1.AppRevisionSpec{
				Revision:  1,
				Manifests: cm.Name,
			}
			config = featuresv1alpha1.AppRevisionConfig{}
			config.AppNamespace = ns.Name

			vk = &featuresv1alpha1.VerificationKey{}
			vk.SetName("test-key-name")
			vk.SetNamespace(ns.Name)
			vk.Data = string(buf.Bytes())
		})

		JustBeforeEach(func() {
			ar.Spec.Config = config
			Expect(k8sClient.Create(ctx, ar.DeepCopy())).Should(Succeed())
		})

		AfterEach(func() {
			Expect(k8sClient.Delete(context.Background(), ar)).Should(Succeed())

			var d appsv1.Deployment
			d.SetName("my-app")
			d.SetNamespace(ns.Name)
			k8sClient.Delete(ctx, &d)
		})

		Context("without manifests existing", func() {
			It(shouldFailWithReason("NotFound"))
		})

		Context("with existing manifests", func() {
			BeforeEach(func() {
				cm.SetNamespace(ns.Name)
				Expect(k8sClient.Create(ctx, cm.DeepCopy())).Should(Succeed())

				hash := sha256.Sum256([]byte(manifests))
				b, err := rsa.SignPKCS1v15(rand.Reader, key, crypto.SHA256, hash[:])
				Expect(err).ShouldNot(HaveOccurred())
				ar.Spec.Signatures = map[string][]byte{
					vk.Name: b,
				}
				Expect(k8sClient.Create(ctx, vk)).Should(Succeed())
				ar.Spec.Signatures = map[string][]byte{
					vk.Name: b,
				}
			})
			AfterEach(func() {
				Expect(k8sClient.Delete(ctx, &cm)).Should(Succeed())
				Expect(k8sClient.Delete(ctx, vk)).Should(Succeed())
			})

			Context("unsigned apprevision", func() {
				BeforeEach(func() {
					ar.Spec.Signatures = nil
				})
				Context("signing is required", func() {
					BeforeEach(func() {
						config.Signing.Optional = false
					})
					It(shouldFailWithReason("InvalidSigning"))
				})
				Context("signing is optional", func() {
					BeforeEach(func() {
						config.Signing.Optional = true
					})
					It(shouldDeploySuccessfully())
				})
			})
			Context("invalid signature in apprevision", func() {
				BeforeEach(func() {
					ar.Spec.Signatures = map[string][]byte{
						vk.Name: []byte("notgonnawork"),
					}
				})

				Context("signing is required", func() {
					BeforeEach(func() {
						config.Signing.Optional = false
					})
					It(shouldFailWithReason("InvalidSigning"))
				})
				Context("signing is optional", func() {
					BeforeEach(func() {
						config.Signing.Optional = true
					})
					It(shouldFailWithReason("InvalidSigning"))
				})
				Context("invalid verification key exists", func() {
					BeforeEach(func() {
						vk.Data = "somekeydata"
						Expect(k8sClient.Update(ctx, vk)).Should(Succeed())
					})
					It(shouldFailWithReason("InvalidSigning"))
				})
				Context("signature verification is disabled", func() {
					BeforeEach(func() {
						config.Signing.Optional = false
						config.Signing.InsecureSkipVerification = true
					})
					It(shouldDeploySuccessfully())
				})
			})

			It(shouldDeploySuccessfully())

			Context("signature is invalid", func() {
				BeforeEach(func() {
					ar.Spec.Signatures = map[string][]byte{
						vk.Name: []byte("wrong"),
					}
				})
				It(shouldFailWithReason("InvalidSigning"))
			})

			It("recreates deleted resources", func() {
				Eventually(getAppRevisionConditions(ctx, ar), timeout, interval).Should(reportSuccess)
				var d appsv1.Deployment
				Expect(k8sClient.Get(ctx, client.ObjectKey{Name: "my-app", Namespace: ar.Namespace}, &d)).Should(Succeed())
				oldUID := d.GetUID()
				havingUID := func(u ktypes.UID) types.GomegaMatcher {
					return WithTransform(func(r featuresv1alpha1.AppRevisionResource) ktypes.UID {
						return r.GetUID()
					}, Equal(u))
				}

				Eventually(getResources(ctx, ar), timeout, interval).Should(ContainElement(havingUID(oldUID)), "AppRevision resources did not contain the deployment")

				Expect(k8sClient.Delete(ctx, &d)).Should(Succeed())

				Eventually(getResources(ctx, ar), timeout, interval).ShouldNot(ContainElement(havingUID(oldUID)), "AppRevision resources still had the old deployment in it")

				Eventually(func() error {
					return k8sClient.Get(ctx, client.ObjectKey{Name: "my-app", Namespace: ar.Namespace}, &d)
				}, timeout, interval).Should(Succeed())
				Expect(d.GetUID()).NotTo(Equal(oldUID))
				Expect(getResources(ctx, ar)()).To(ContainElement(havingUID(d.GetUID())), "AppRevision resources did not contain the replacement deployment")

				// re-check resources
				Expect(k8sClient.Get(ctx, client.ObjectKey{Name: "my-app", Namespace: ar.Namespace}, &d)).Should(Succeed())

				Expect(ar.Status.OriginalResources).To(HaveLen(1))
				Expect(ar.Status.OriginalResources).Should(ContainElement(havingUID(d.GetUID())), "original resources was not updated - recreated resource not added")
				Expect(ar.Status.OriginalResources).ShouldNot(ContainElement(havingUID(oldUID)), "original resources was not updated - deleted resource not removed")
			})
			Context("dev mode is on", func() {
				BeforeEach(func() {
					config.DevMode = true
				})
				It("does not recreate deleted resources", func() {
					Eventually(getAppRevisionConditions(ctx, ar), timeout, interval).Should(reportSuccess)
					var d appsv1.Deployment
					Expect(k8sClient.Get(ctx, client.ObjectKey{Name: "my-app", Namespace: ar.Namespace}, &d)).Should(Succeed())
					oldUID := d.GetUID()
					havingUID := func(u ktypes.UID) types.GomegaMatcher {
						return WithTransform(func(r featuresv1alpha1.AppRevisionResource) ktypes.UID {
							return r.GetUID()
						}, Equal(u))
					}

					Eventually(getResources(ctx, ar), timeout, interval).Should(ContainElement(havingUID(oldUID)), "AppRevision resources did not contain the deployment")

					Expect(k8sClient.Delete(ctx, &d)).Should(Succeed())

					// re-check resources
					Eventually(func() error {
						return k8sClient.Get(ctx, client.ObjectKey{Name: "my-app", Namespace: ar.Namespace}, &d)
					}, timeout, interval).ShouldNot(Succeed())
					Eventually(getResources(ctx, ar), timeout, interval).ShouldNot(ContainElement(havingUID(oldUID)), "AppRevision resources still had the old deployment in it")

					Expect(ar.Status.OriginalResources).To(HaveLen(1))
					Expect(ar.Status.OriginalResources).Should(ContainElement(havingUID(oldUID)), "original resources changed but it should not have")
				})
			})
		})
	})
})

func getResources(ctx context.Context, ar *featuresv1alpha1.AppRevision) func() interface{} {
	return func() interface{} {
		err := k8sClient.Get(ctx, client.ObjectKey{Name: ar.Name, Namespace: ar.Namespace}, ar)
		Expect(err).NotTo(HaveOccurred(), "failed to get AppRevision")
		return ar.Status.Resources
	}
}

func getAppRevisionConditions(ctx context.Context, ar *featuresv1alpha1.AppRevision) func() interface{} {
	return func() interface{} {
		err := k8sClient.Get(ctx, client.ObjectKey{Name: ar.Name, Namespace: ar.Namespace}, ar)
		Expect(err).NotTo(HaveOccurred(), "failed to get AppRevision")
		return ar.Status.Conditions
	}
}

func haveConditionType(t featuresv1alpha1.AppRevisionConditionType) types.GomegaMatcher {
	return WithTransform(func(c featuresv1alpha1.AppRevisionCondition) featuresv1alpha1.AppRevisionConditionType {
		return c.Type
	}, Equal(t))
}

func haveConditionStatus(s corev1.ConditionStatus) types.GomegaMatcher {
	return WithTransform(func(c featuresv1alpha1.AppRevisionCondition) corev1.ConditionStatus {
		return c.Status
	}, Equal(s))
}

func haveReason(s string) types.GomegaMatcher {
	return WithTransform(func(c featuresv1alpha1.AppRevisionCondition) string {
		return c.Reason
	}, Equal(s))
}
