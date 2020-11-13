package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	featuresv1alpha1 "github.com/criticalstack/stackapps/api/v1alpha1"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	gomegatypes "github.com/onsi/gomega/types"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

var _ = Describe("StackValueController", func() {
	const timeout = time.Second * 5
	const interval = time.Millisecond * 10

	ctx := context.Background()
	providerAddr := fmt.Sprintf("localhost:%d", 8089)
	providerToken := "test-token"

	var namespace string
	var vaultServer serverWithCancel
	var stackVal featuresv1alpha1.StackValue
	var config featuresv1alpha1.StackAppConfig

	BeforeEach(func() {
		namespace = ns.Name

		stackVal = featuresv1alpha1.StackValue{
			TypeMeta: metav1.TypeMeta{
				Kind:       "StackValue",
				APIVersion: "features.criticalstack.com/v1alpha1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      "vault-stackvalue",
				Namespace: namespace,
			},
			Spec: featuresv1alpha1.StackValueSpec{
				AppName:    "sa-config",
				ObjectType: "Secret",
				Path:       "v1/secret/mydata",
				SourceType: featuresv1alpha1.StackValueSourceVault,
			},
		}

		vaultData := map[string]interface{}{
			"mydata": "admin",
		}
		vaultServer = newMockVaultServer(providerAddr, providerToken, vaultData)
		go vaultServer.Run()

		config = featuresv1alpha1.StackAppConfig{
			TypeMeta: metav1.TypeMeta{
				Kind:       "StackAppConfig",
				APIVersion: "features.criticalstack.com/v1alpha1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name: "sa-config",
			},
			Spec: featuresv1alpha1.StackAppConfigSpec{
				AppNamespace: namespace,
				StackValues: featuresv1alpha1.StackValuesConfig{
					Secret: &corev1.SecretReference{
						Name:      "vault-token",
						Namespace: namespace,
					},
					Sources: []*featuresv1alpha1.StackValueSource{
						{
							Name:  "vault-source",
							Type:  featuresv1alpha1.StackValueSourceVault,
							Route: "http://" + providerAddr,
							Token: []byte(providerToken),
						},
						{
							Name:  "artifactory-source",
							Type:  featuresv1alpha1.StackValueSourceArtifactory,
							Route: "whatever",
							Token: []byte(providerToken),
						},
					},
				},
			},
		}

	})

	AfterEach(func() {
		Expect(vaultServer.Cancel(3 * time.Second)).Should(Succeed())
	})

	failedConditionsWithReason := func(r string) gomegatypes.GomegaMatcher {
		return ContainElement(SatisfyAll(
			haveStackValueConditionType(featuresv1alpha1.StackValueFailed),
			haveStackValueConditionStatus(corev1.ConditionTrue),
			haveStackValueReason(r),
		))
	}

	succeedConditions := ContainElements(
		SatisfyAll(
			haveStackValueConditionType(featuresv1alpha1.StackValueFailed),
			haveStackValueConditionStatus(corev1.ConditionFalse),
		),
		SatisfyAll(
			haveStackValueConditionType(featuresv1alpha1.StackValueReady),
			haveStackValueConditionStatus(corev1.ConditionTrue),
		),
	)

	shouldFailWithReason := func(r string) (string, func()) {
		return "Should set an error status", func() {
			Eventually(getStackValueConditions(ctx, &stackVal), timeout, interval).Should(failedConditionsWithReason(r))
		}
	}

	Context("When StackValue is deployed", func() {
		JustBeforeEach(func() {
			Expect(k8sClient.Create(ctx, &stackVal)).Should(Succeed())
		})

		JustAfterEach(func() {
			Expect(k8sClient.Delete(ctx, &stackVal)).Should(Succeed())
		})

		Context("Without StackAppConfig existing", func() {
			It(shouldFailWithReason("GetStackAppConfig"))
		})

		Context("With existing StackAppConfig", func() {
			Context("When StackValues is not enabled", func() {
				BeforeEach(func() {
					config.Spec.StackValues.Enabled = false
				})

				JustBeforeEach(func() {
					Expect(k8sClient.Create(ctx, &config)).Should(Succeed())
				})

				JustAfterEach(func() {
					Expect(k8sClient.Delete(ctx, &config)).Should(Succeed())
				})

				It(shouldFailWithReason("StackValues not enabled"))
			})

			Context("When StackValues is enabled", func() {
				BeforeEach(func() {
					config.Spec.StackValues.Enabled = true
					Expect(k8sClient.Create(ctx, &config)).Should(Succeed())
				})

				JustAfterEach(func() {
					Expect(k8sClient.Delete(ctx, &config)).Should(Succeed())
				})

				Context("When the access token does not exist", func() {
					It(shouldFailWithReason("GetCredentials"))
				})

				Context("When the access token exists", func() {
					var accessToken corev1.Secret

					BeforeEach(func() {
						accessToken = corev1.Secret{
							TypeMeta: metav1.TypeMeta{
								Kind:       "Secret",
								APIVersion: "v1",
							},
							ObjectMeta: metav1.ObjectMeta{
								Name:      "vault-token",
								Namespace: namespace,
							},
							Type: "Opaque",
							Data: map[string][]byte{
								"vault-source":       []byte(providerToken),
								"artifactory-source": []byte(providerToken),
							},
						}
						Expect(k8sClient.Create(ctx, &accessToken)).Should(Succeed())
					})

					AfterEach(func() {
						Expect(k8sClient.Delete(ctx, &accessToken)).Should(Succeed())
					})

					Context("When the access token is wrong", func() {
						BeforeEach(func() {
							accessToken.Data["vault-source"] = []byte("wrongToken")
							Expect(k8sClient.Update(ctx, &accessToken)).Should(Succeed())
						})

						It(shouldFailWithReason("FetchValues"))
					})

					Context("When Object already exists", func() {
						// need to update when we figure out the desired behavior if an object already exists - possibly update
						var s corev1.Secret

						BeforeEach(func() {
							s = corev1.Secret{
								TypeMeta: metav1.TypeMeta{
									Kind:       "Secret",
									APIVersion: "v1",
								},
								ObjectMeta: metav1.ObjectMeta{
									Name:      "vault-stackvalue",
									Namespace: namespace,
								},
								Type: "Opaque",
							}
							Expect(k8sClient.Create(ctx, &s)).Should(Succeed())
						})

						AfterEach(func() {
							Expect(k8sClient.Delete(ctx, &s)).Should(Succeed())
						})

						It("Will not deploy new object from StackValue and will NOT fail", func() {
							Eventually(getStackValueConditions(ctx, &stackVal), timeout, interval).Should(succeedConditions)
							obj := &corev1.Secret{}
							Eventually(func() bool {
								err := k8sClient.Get(ctx, types.NamespacedName{Name: "vault-stackvalue", Namespace: namespace}, obj)
								return err == nil
							}, timeout, interval).Should(BeTrue())
							Expect(string(obj.Data["value"])).Should(BeEmpty())

						})
					})

					Context("When Object doesn't already exist", func() {
						It("Should create new object successfully from StackValue", func() {
							Eventually(getStackValueConditions(ctx, &stackVal), timeout, interval).Should(succeedConditions)
							obj := &corev1.Secret{}
							Eventually(func() bool {
								err := k8sClient.Get(ctx, types.NamespacedName{Name: "vault-stackvalue", Namespace: namespace}, obj)
								return err == nil
							}, timeout, interval).Should(BeTrue())
							Expect(string(obj.Data["value"])).To(Equal("admin"))
						})

					})
				})
			})
		})
	})
})

func getStackValueConditions(ctx context.Context, sv *featuresv1alpha1.StackValue) func() []featuresv1alpha1.StackValueCondition {
	return func() []featuresv1alpha1.StackValueCondition {
		found := featuresv1alpha1.StackValue{}
		err := k8sClient.Get(ctx, types.NamespacedName{Name: sv.Name, Namespace: sv.Namespace}, &found)
		Expect(err).NotTo(HaveOccurred(), "failed to get StackValue")

		return found.Status.Conditions
	}
}

func haveStackValueConditionType(t featuresv1alpha1.StackValueConditionType) gomegatypes.GomegaMatcher {
	return WithTransform(func(c featuresv1alpha1.StackValueCondition) featuresv1alpha1.StackValueConditionType {
		return c.Type
	}, Equal(t))
}

func haveStackValueConditionStatus(s corev1.ConditionStatus) gomegatypes.GomegaMatcher {
	return WithTransform(func(c featuresv1alpha1.StackValueCondition) corev1.ConditionStatus {
		return c.Status
	}, Equal(s))
}

func haveStackValueReason(s string) gomegatypes.GomegaMatcher {
	return WithTransform(func(c featuresv1alpha1.StackValueCondition) string {
		return c.Reason
	}, Equal(s))
}

type serverWithCancel struct {
	server *http.Server
	done   chan (error)
}

func newServerWithCancel(h http.Handler, addr string) serverWithCancel {
	return serverWithCancel{
		done: make(chan error),
		server: &http.Server{
			Addr:    addr,
			Handler: h,
		},
	}
}

func (s serverWithCancel) Run() {
	defer close(s.done)
	if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		fmt.Println(err)
		s.done <- err
	}
}

func (s serverWithCancel) Cancel(timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	if err := s.server.Shutdown(ctx); err != nil {
		return err
	}
	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-s.done:
		return err
	}
}

func newMockVaultServer(addr, token string, data map[string]interface{}) serverWithCancel {
	mux := http.NewServeMux()
	mux.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Header.Get("X-Vault-Token") != token {
			http.Error(w, "{\"errors\":[\"permission denied\"]}", http.StatusForbidden)
			return
		}
		if !strings.HasPrefix(r.URL.Path, "/v1/secret/") {
			http.Error(w, "{\"errors\":[\"not found\"]}", http.StatusNotFound)
			return
		}
		key := strings.TrimPrefix(r.URL.Path, "/v1/secret/")
		v, ok := data[key]
		if !ok {
			http.Error(w, "{\"errors\":[\"not found\"]}", http.StatusNotFound)
			return
		}
		b, err := json.Marshal(map[string]interface{}{
			"Data": map[string]interface{}{
				"data": map[string]interface{}{
					"value": v,
				},
			},
		})
		if err != nil {
			panic(err)
		}
		w.Write(b)
	}))
	return newServerWithCancel(mux, addr)
}
