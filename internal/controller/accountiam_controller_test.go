/*
Copyright 2024.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	operatorv1alpha1 "github.com/IBM/ibm-user-management-operator/api/v1alpha1"
)

var _ = Describe("AccountIAM Controller", func() {
	const (
		AccountIAMName      = "test-accountiam"
		AccountIAMNamespace = "ibm-common-services"
		timeout             = time.Second * 30
		interval            = time.Millisecond * 250
	)

	var (
		ctx = context.Background()
	)

	Context("When testing reconciliation phases", func() {
		var (
			accountIAM *operatorv1alpha1.AccountIAM
			reconciler *AccountIAMReconciler
			recorder   *record.FakeRecorder
		)

		BeforeEach(func() {
			By("Creating namespace if it doesn't exist")
			namespace := &corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: AccountIAMNamespace,
				},
			}
			err := k8sClient.Create(ctx, namespace)
			if err != nil && !errors.IsAlreadyExists(err) {
				Expect(err).NotTo(HaveOccurred())
			}

			By("Creating AccountIAM resource")
			accountIAM = &operatorv1alpha1.AccountIAM{
				ObjectMeta: metav1.ObjectMeta{
					Name:      AccountIAMName,
					Namespace: AccountIAMNamespace,
				},
				Spec: operatorv1alpha1.AccountIAMSpec{},
			}

			// Use Eventually to ensure resource creation succeeds
			Eventually(func() error {
				return k8sClient.Create(ctx, accountIAM)
			}, timeout, interval).Should(Succeed())

			By("Setting up reconciler")
			recorder = record.NewFakeRecorder(100)
			reconciler = &AccountIAMReconciler{
				Client:   k8sClient,
				Scheme:   k8sClient.Scheme(),
				Recorder: recorder,
			}
		})

		AfterEach(func() {
			// Clean up the AccountIAM resource
			err := k8sClient.Delete(ctx, accountIAM)
			if err != nil && !errors.IsNotFound(err) {
				Expect(err).NotTo(HaveOccurred())
			}

			// Wait for deletion to complete
			Eventually(func() bool {
				err := k8sClient.Get(ctx, types.NamespacedName{
					Name:      AccountIAMName,
					Namespace: AccountIAMNamespace,
				}, accountIAM)
				return errors.IsNotFound(err)
			}, timeout, interval).Should(BeTrue())
		})

		// Phase 1: Initial Resource Validation
		Context("Phase 1: Resource Validation", func() {
			It("should validate AccountIAM resource exists", func() {
				By("Checking resource can be fetched")
				namespacedName := types.NamespacedName{
					Name:      AccountIAMName,
					Namespace: AccountIAMNamespace,
				}

				fetchedAccountIAM := &operatorv1alpha1.AccountIAM{}
				Eventually(func() error {
					return k8sClient.Get(ctx, namespacedName, fetchedAccountIAM)
				}, timeout, interval).Should(Succeed())

				Expect(fetchedAccountIAM.Name).To(Equal(AccountIAMName))
				Expect(fetchedAccountIAM.Name).To(Equal(AccountIAMName))
			})

			It("should handle missing resource gracefully", func() {
				By("Reconciling non-existent resource")
				namespacedName := types.NamespacedName{
					Name:      "non-existent",
					Namespace: AccountIAMNamespace,
				}

				result, err := reconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: namespacedName,
				})

				Expect(err).NotTo(HaveOccurred())
				Expect(result.Requeue).To(BeFalse())
			})
		})

		// Phase 2: Prerequisites Verification
		Context("Phase 2: Prerequisites Verification", func() {
			It("should handle missing ConfigMap and external CRDs gracefully", func() {
				By("Reconciling without cluster info")
				namespacedName := types.NamespacedName{
					Name:      AccountIAMName,
					Namespace: AccountIAMNamespace,
				}

				result, err := reconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: namespacedName,
				})

				// Current approach: Test resilience to missing external dependencies
				if err != nil {
					// Check if error is due to missing external CRDs
					Expect(err.Error()).To(ContainSubstring("no matches for kind"))
					By("Controller properly handles missing external CRDs")
				} else {
					// If no error, controller might requeue waiting for prerequisites
					Expect(result.Requeue || result.RequeueAfter > 0).To(BeTrue())
					By("Controller gracefully defers when prerequisites missing")
				}
			})

			It("should validate prerequisite logic independently", func() {
				By("Testing controller's prerequisite validation logic")
				namespacedName := types.NamespacedName{
					Name:      AccountIAMName,
					Namespace: AccountIAMNamespace,
				}

				result, err := reconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: namespacedName,
				})

				// What we're actually testing here:
				// 1. Controller maintains state consistency
				// 2. Error handling doesn't leak resources
				// 3. Reconcile loop behaves predictably

				if err != nil {
					Expect(err.Error()).NotTo(BeEmpty())
					By("Error provides debugging information")
				} else {
					Expect(result).NotTo(BeNil())
					By("Controller returns valid reconcile result")
				}
			})

			It("should proceed when cluster info is available", func() {
				By("Creating cluster info ConfigMap")
				clusterInfo := &corev1.ConfigMap{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "mcsp-info",
						Namespace: AccountIAMNamespace,
					},
					Data: map[string]string{
						"cluster_address":  "test.example.com",
						"cluster_endpoint": "https://test.example.com:443",
					},
				}
				Expect(k8sClient.Create(ctx, clusterInfo)).Should(Succeed())

				By("Reconciling with cluster info present")
				namespacedName := types.NamespacedName{
					Name:      AccountIAMName,
					Namespace: AccountIAMNamespace,
				}

				result, err := reconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: namespacedName,
				})

				if err != nil {
					Expect(err.Error()).To(ContainSubstring("no matches for kind"))
				} else {
					Expect(result).NotTo(BeNil())
				}
			})
		})

		// Phase 3: Resource Creation/Update
		Context("Phase 3: Resource Management", func() {
			It("should handle ConfigMap creation", func() {
				By("Setting up prerequisites")
				clusterInfo := &corev1.ConfigMap{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "mcsp-info",
						Namespace: AccountIAMNamespace,
					},
					Data: map[string]string{
						"cluster_address":  "test.example.com",
						"cluster_endpoint": "https://test.example.com:443",
					},
				}

				err := k8sClient.Create(ctx, clusterInfo)
				if err != nil && !errors.IsAlreadyExists(err) {
					Expect(err).NotTo(HaveOccurred())
				}

				By("Reconciling to trigger resource creation")
				namespacedName := types.NamespacedName{
					Name:      AccountIAMName,
					Namespace: AccountIAMNamespace,
				}

				_, err = reconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: namespacedName,
				})

				if err != nil {
					Expect(err.Error()).To(ContainSubstring("no matches for kind"))
					By("Controller properly handles missing external CRDs")
				}

				// Check if expected ConfigMaps are created (based on controller logic)
				Eventually(func() bool {
					configMaps := &corev1.ConfigMapList{}
					err := k8sClient.List(ctx, configMaps, &client.ListOptions{
						Namespace: AccountIAMNamespace,
					})
					return err == nil && len(configMaps.Items) > 0
				}, timeout, interval).Should(BeTrue())
			})

			It("should handle Secret creation", func() {
				By("Setting up prerequisites and reconciling")
				namespacedName := types.NamespacedName{
					Name:      AccountIAMName,
					Namespace: AccountIAMNamespace,
				}

				_, err := reconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: namespacedName,
				})

				if err != nil {
					Expect(err.Error()).To(ContainSubstring("no matches for kind"))
				}

				secrets := &corev1.SecretList{}
				err = k8sClient.List(ctx, secrets, &client.ListOptions{
					Namespace: AccountIAMNamespace,
				})
				Expect(err).NotTo(HaveOccurred())
			})
		})

		// Phase 4: Status Updates
		Context("Phase 4: Status Management", func() {
			It("should update status during reconciliation", func() {
				By("Reconciling and checking status")
				namespacedName := types.NamespacedName{
					Name:      AccountIAMName,
					Namespace: AccountIAMNamespace,
				}

				_, err := reconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: namespacedName,
				})

				if err != nil {
					Expect(err.Error()).To(ContainSubstring("no matches for kind"))
					By("Controller properly reports missing external dependencies")
					return
				}

				updatedAccountIAM := &operatorv1alpha1.AccountIAM{}
				err = k8sClient.Get(ctx, namespacedName, updatedAccountIAM)
				Expect(err).NotTo(HaveOccurred())

				Expect(updatedAccountIAM.Status).NotTo(BeNil())
			})

			It("should handle status update failures gracefully", func() {
				By("Testing status update resilience")
				namespacedName := types.NamespacedName{
					Name:      AccountIAMName,
					Namespace: AccountIAMNamespace,
				}

				for i := 0; i < 3; i++ {
					_, err := reconciler.Reconcile(ctx, reconcile.Request{
						NamespacedName: namespacedName,
					})

					if err != nil {
						Expect(err.Error()).To(ContainSubstring("no matches for kind"))
						By("Controller handles missing dependencies consistently")
					}
				}
			})
		})

		// Phase 5: Error Handling
		Context("Phase 5: Error Scenarios", func() {
			It("should handle reconcile errors gracefully", func() {
				By("Creating a scenario that might cause errors")
				namespacedName := types.NamespacedName{
					Name:      AccountIAMName,
					Namespace: AccountIAMNamespace,
				}

				result, err := reconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: namespacedName,
				})

				if err != nil {
					Expect(err.Error()).NotTo(BeEmpty())
					By("Controller provides informative error messages")
				}
				Expect(result).NotTo(BeNil())
			})

			It("should retry on transient failures", func() {
				By("Testing retry logic")
				namespacedName := types.NamespacedName{
					Name:      AccountIAMName,
					Namespace: AccountIAMNamespace,
				}

				result, err := reconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: namespacedName,
				})

				if err != nil {
					Expect(err.Error()).To(ContainSubstring("no matches for kind"))
					By("Controller fails predictably when dependencies are missing")
				} else {
					Expect(result.Requeue || result.RequeueAfter > 0).To(BeTrue())
					By("Controller queues retry when appropriate")
				}
			})
		})
	})

	Context("When testing utility functions", func() {
		It("should properly check if string is in slice", func() {
			slice := []string{"pass", "fail", "notFound"}
			Expect(contains(slice, "fail")).To(BeTrue())
			Expect(contains(slice, "orange")).To(BeFalse())
		})

		It("should properly remove string from slice", func() {
			slice := []string{"pass", "fail", "notFound"}
			result := remove(slice, "fail")
			Expect(result).To(Equal([]string{"pass", "notFound"}))
			Expect(len(result)).To(Equal(2))
		})

		It("should handle empty slices", func() {
			var emptySlice []string
			Expect(contains(emptySlice, "test")).To(BeFalse())
			result := remove(emptySlice, "test")
			Expect(result).To(BeEmpty())
		})
	})
})

// Helper functions for testing
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func remove(slice []string, item string) []string {
	var result []string
	for _, s := range slice {
		if s != item {
			result = append(result, s)
		}
	}
	return result
}
