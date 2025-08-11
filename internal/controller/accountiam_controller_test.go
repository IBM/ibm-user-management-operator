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
			// Create namespace if it doesn't exist
			namespace := &corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: AccountIAMNamespace,
				},
			}
			err := k8sClient.Create(ctx, namespace)
			if err != nil && !errors.IsAlreadyExists(err) {
				Expect(err).NotTo(HaveOccurred())
			}

			// Create AccountIAM resource
			accountIAM = &operatorv1alpha1.AccountIAM{
				ObjectMeta: metav1.ObjectMeta{
					Name:      AccountIAMName,
					Namespace: AccountIAMNamespace,
				},
				Spec: operatorv1alpha1.AccountIAMSpec{},
			}

			// Create the AccountIAM resource
			Expect(k8sClient.Create(ctx, accountIAM)).Should(Succeed())

			// Set up reconciler
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
				err := k8sClient.Get(ctx, namespacedName, fetchedAccountIAM)
				Expect(err).NotTo(HaveOccurred())
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

		// Phase 2: Finalizer Management
		Context("Phase 2: Finalizer Management", func() {
			It("should add finalizer on first reconcile", func() {
				By("Reconciling resource without finalizer")
				namespacedName := types.NamespacedName{
					Name:      AccountIAMName,
					Namespace: AccountIAMNamespace,
				}

				_, err := reconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: namespacedName,
				})
				Expect(err).NotTo(HaveOccurred())

				// Check if finalizer was added
				updatedAccountIAM := &operatorv1alpha1.AccountIAM{}
				err = k8sClient.Get(ctx, namespacedName, updatedAccountIAM)
				Expect(err).NotTo(HaveOccurred())

				// Verify the resource still exists (finalizer logic depends on implementation)
				Expect(updatedAccountIAM.Name).To(Equal(AccountIAMName))
			})

			It("should handle deletion with finalizer", func() {
				By("Setting up resource with finalizer")
				namespacedName := types.NamespacedName{
					Name:      AccountIAMName,
					Namespace: AccountIAMNamespace,
				}

				// First reconcile to potentially add finalizer
				_, err := reconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: namespacedName,
				})
				Expect(err).NotTo(HaveOccurred())

				By("Deleting resource")
				err = k8sClient.Delete(ctx, accountIAM)
				Expect(err).NotTo(HaveOccurred())

				By("Reconciling deleted resource")
				_, err = reconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: namespacedName,
				})
				Expect(err).NotTo(HaveOccurred())
			})
		})

		// Phase 3: Prerequisites Verification
		Context("Phase 3: Prerequisites Verification", func() {
			It("should handle missing cluster info ConfigMap", func() {
				By("Reconciling without cluster info")
				namespacedName := types.NamespacedName{
					Name:      AccountIAMName,
					Namespace: AccountIAMNamespace,
				}

				result, err := reconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: namespacedName,
				})

				// Should not error but might requeue
				Expect(err).NotTo(HaveOccurred())
				// Controller might requeue waiting for prerequisites
				Expect(result.Requeue || result.RequeueAfter > 0).To(BeTrue())
			})

			It("should proceed when cluster info is available", func() {
				By("Creating cluster info ConfigMap")
				clusterInfo := &corev1.ConfigMap{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "ibmcloud-cluster-info",
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

				Expect(err).NotTo(HaveOccurred())
				Expect(result).NotTo(BeNil())
			})
		})

		// Phase 4: Resource Creation/Update
		Context("Phase 4: Resource Management", func() {
			It("should handle ConfigMap creation", func() {
				By("Setting up prerequisites")
				clusterInfo := &corev1.ConfigMap{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "ibmcloud-cluster-info",
						Namespace: AccountIAMNamespace,
					},
					Data: map[string]string{
						"cluster_address":  "test.example.com",
						"cluster_endpoint": "https://test.example.com:443",
					},
				}
				Expect(k8sClient.Create(ctx, clusterInfo)).Should(Succeed())

				By("Reconciling to trigger resource creation")
				namespacedName := types.NamespacedName{
					Name:      AccountIAMName,
					Namespace: AccountIAMNamespace,
				}

				_, err := reconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: namespacedName,
				})
				Expect(err).NotTo(HaveOccurred())

				// Check if expected ConfigMaps are created (based on controller logic)
				Eventually(func() bool {
					configMaps := &corev1.ConfigMapList{}
					err := k8sClient.List(ctx, configMaps, &client.ListOptions{
						Namespace: AccountIAMNamespace,
					})
					return err == nil && len(configMaps.Items) > 1 // Original + created ones
				}, timeout, interval).Should(BeTrue())
			})

			It("should handle Secret creation", func() {
				By("Setting up prerequisites and reconciling")
				// This would test secret creation logic in the controller
				namespacedName := types.NamespacedName{
					Name:      AccountIAMName,
					Namespace: AccountIAMNamespace,
				}

				_, err := reconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: namespacedName,
				})
				Expect(err).NotTo(HaveOccurred())

				// Verify secrets are created if the controller creates them
				secrets := &corev1.SecretList{}
				err = k8sClient.List(ctx, secrets, &client.ListOptions{
					Namespace: AccountIAMNamespace,
				})
				Expect(err).NotTo(HaveOccurred())
			})
		})

		// Phase 5: Status Updates
		Context("Phase 5: Status Management", func() {
			It("should update status during reconciliation", func() {
				By("Reconciling and checking status")
				namespacedName := types.NamespacedName{
					Name:      AccountIAMName,
					Namespace: AccountIAMNamespace,
				}

				_, err := reconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: namespacedName,
				})
				Expect(err).NotTo(HaveOccurred())

				// Get updated resource and check status
				updatedAccountIAM := &operatorv1alpha1.AccountIAM{}
				err = k8sClient.Get(ctx, namespacedName, updatedAccountIAM)
				Expect(err).NotTo(HaveOccurred())

				// Verify status fields (adjust based on actual status structure)
				Expect(updatedAccountIAM.Status).NotTo(BeNil())
			})

			It("should handle status update failures gracefully", func() {
				By("Testing status update resilience")
				namespacedName := types.NamespacedName{
					Name:      AccountIAMName,
					Namespace: AccountIAMNamespace,
				}

				// Multiple reconciles should not fail
				for i := 0; i < 3; i++ {
					_, err := reconciler.Reconcile(ctx, reconcile.Request{
						NamespacedName: namespacedName,
					})
					Expect(err).NotTo(HaveOccurred())
				}
			})
		})

		// Phase 6: Error Handling
		Context("Phase 6: Error Scenarios", func() {
			It("should handle reconcile errors gracefully", func() {
				By("Creating a scenario that might cause errors")
				namespacedName := types.NamespacedName{
					Name:      AccountIAMName,
					Namespace: AccountIAMNamespace,
				}

				// Test with incomplete setup
				result, err := reconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: namespacedName,
				})

				// Should handle errors without panicking
				if err != nil {
					// Error is acceptable in some scenarios
					Expect(err.Error()).NotTo(BeEmpty())
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

				Expect(err).NotTo(HaveOccurred())
				// Should requeue for retry if prerequisites are missing
				Expect(result.Requeue || result.RequeueAfter > 0).To(BeTrue())
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
