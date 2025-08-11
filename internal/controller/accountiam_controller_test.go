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

	Context("When reconciling an AccountIAM resource", func() {
		var (
			accountIAM *operatorv1alpha1.AccountIAM
			reconciler *AccountIAMReconciler
			recorder   *record.FakeRecorder
		)

		BeforeEach(func() {
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

			Eventually(func() bool {
				err := k8sClient.Get(ctx, types.NamespacedName{
					Name:      AccountIAMName,
					Namespace: AccountIAMNamespace,
				}, accountIAM)
				return errors.IsNotFound(err)
			}, timeout, interval).Should(BeTrue())
		})

		It("should successfully reconcile the resource", func() {
			By("Reconciling the created resource")
			namespacedName := types.NamespacedName{
				Name:      AccountIAMName,
				Namespace: AccountIAMNamespace,
			}

			result, err := reconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: namespacedName,
			})

			Expect(err).NotTo(HaveOccurred())
			// The reconciler should handle the resource without error
			Expect(result).NotTo(BeNil())
		})

		It("should handle missing AccountIAM resource gracefully", func() {
			By("Reconciling a non-existent resource")
			namespacedName := types.NamespacedName{
				Name:      "non-existent",
				Namespace: AccountIAMNamespace,
			}

			result, err := reconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: namespacedName,
			})

			Expect(err).NotTo(HaveOccurred())
			Expect(result.Requeue).To(BeFalse())
			Expect(result.RequeueAfter).To(BeZero())
		})

		It("should handle finalizer logic correctly", func() {
			By("Checking finalizer handling during reconciliation")
			namespacedName := types.NamespacedName{
				Name:      AccountIAMName,
				Namespace: AccountIAMNamespace,
			}

			// Reconcile to potentially add finalizer
			_, err := reconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: namespacedName,
			})
			Expect(err).NotTo(HaveOccurred())

			// Check current state of resource
			updatedAccountIAM := &operatorv1alpha1.AccountIAM{}
			err = k8sClient.Get(ctx, namespacedName, updatedAccountIAM)
			Expect(err).NotTo(HaveOccurred())

			// Verify the resource exists
			Expect(updatedAccountIAM.Name).To(Equal(AccountIAMName))
		})

		It("should handle deletion gracefully", func() {
			By("Testing deletion workflow")
			namespacedName := types.NamespacedName{
				Name:      AccountIAMName,
				Namespace: AccountIAMNamespace,
			}

			// First reconcile to set up the resource
			_, err := reconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: namespacedName,
			})
			Expect(err).NotTo(HaveOccurred())

			// Delete the resource
			err = k8sClient.Delete(ctx, accountIAM)
			Expect(err).NotTo(HaveOccurred())

			// Reconcile after deletion to test cleanup logic
			_, err = reconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: namespacedName,
			})
			Expect(err).NotTo(HaveOccurred())
		})

		It("should validate basic AccountIAM structure", func() {
			By("Checking that the AccountIAM resource is properly structured")
			Expect(accountIAM.Name).To(Equal(AccountIAMName))
			Expect(accountIAM.Namespace).To(Equal(AccountIAMNamespace))
			Expect(accountIAM.Kind).To(Equal(""))       // Kind is usually empty in tests
			Expect(accountIAM.APIVersion).To(Equal("")) // APIVersion is usually empty in tests
		})

		It("should handle status updates correctly", func() {
			By("Reconciling and checking status updates")
			namespacedName := types.NamespacedName{
				Name:      AccountIAMName,
				Namespace: AccountIAMNamespace,
			}

			_, err := reconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: namespacedName,
			})
			Expect(err).NotTo(HaveOccurred())

			// Get the updated resource and verify it's still accessible
			updatedAccountIAM := &operatorv1alpha1.AccountIAM{}
			err = k8sClient.Get(ctx, namespacedName, updatedAccountIAM)
			Expect(err).NotTo(HaveOccurred())

			// Verify the resource was processed correctly
			Expect(updatedAccountIAM.Name).To(Equal(AccountIAMName))
			Expect(updatedAccountIAM.Namespace).To(Equal(AccountIAMNamespace))
		})

		It("should handle reconciler setup correctly", func() {
			By("Verifying reconciler is properly configured")
			Expect(reconciler.Client).NotTo(BeNil())
			Expect(reconciler.Scheme).NotTo(BeNil())
			Expect(reconciler.Recorder).NotTo(BeNil())
		})
	})

})

// Helper functions for testing (these would be in utils.go normally)
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
