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

package utils

import (
	"context"
	"os"
	"testing"

	odlm "github.com/IBM/operand-deployment-lifecycle-manager/v4/api/v1alpha1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	routev1 "github.com/openshift/api/route/v1"
	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
)

var (
	testEnv   *envtest.Environment
	cfg       *rest.Config
	k8sClient client.Client
	ctx       = context.Background()
)

func TestUtils(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Utils Suite")
}

var _ = BeforeSuite(func() {
	By("bootstrapping test environment")
	testEnv = &envtest.Environment{}

	var err error
	cfg, err = testEnv.Start()
	Expect(err).NotTo(HaveOccurred())
	Expect(cfg).NotTo(BeNil())

	// Add schemes
	err = corev1.AddToScheme(scheme.Scheme)
	Expect(err).NotTo(HaveOccurred())
	err = appsv1.AddToScheme(scheme.Scheme)
	Expect(err).NotTo(HaveOccurred())
	err = batchv1.AddToScheme(scheme.Scheme)
	Expect(err).NotTo(HaveOccurred())
	err = routev1.AddToScheme(scheme.Scheme)
	Expect(err).NotTo(HaveOccurred())
	err = odlm.AddToScheme(scheme.Scheme)
	Expect(err).NotTo(HaveOccurred())

	k8sClient, err = client.New(cfg, client.Options{Scheme: scheme.Scheme})
	Expect(err).NotTo(HaveOccurred())
	Expect(k8sClient).NotTo(BeNil())
})

var _ = AfterSuite(func() {
	By("tearing down the test environment")
	err := testEnv.Stop()
	Expect(err).NotTo(HaveOccurred())
})

var _ = Describe("Utils Functions", func() {

	Context("Environment Functions", func() {
		It("should get operator namespace from environment", func() {
			os.Setenv("OPERATOR_NAMESPACE", "test-operator-ns")
			defer os.Unsetenv("OPERATOR_NAMESPACE")

			ns := GetOperatorNamespace()
			Expect(ns).To(Equal("test-operator-ns"))
		})

		It("should return empty string when OPERATOR_NAMESPACE is not set", func() {
			os.Unsetenv("OPERATOR_NAMESPACE")
			ns := GetOperatorNamespace()
			Expect(ns).To(Equal(""))
		})

		It("should get watch namespace from environment", func() {
			// Test with WATCH_NAMESPACE set
			os.Setenv("WATCH_NAMESPACE", "test-watch-ns")
			defer os.Unsetenv("WATCH_NAMESPACE")

			ns := GetWatchNamespace()
			Expect(ns).To(Equal("test-watch-ns"))
		})

		It("should fallback to operator namespace when WATCH_NAMESPACE is not set", func() {
			os.Unsetenv("WATCH_NAMESPACE")
			os.Setenv("OPERATOR_NAMESPACE", "test-operator-ns")
			defer os.Unsetenv("OPERATOR_NAMESPACE")

			ns := GetWatchNamespace()
			Expect(ns).To(Equal("test-operator-ns"))
		})
	})

	Context("Unstructured Functions", func() {
		It("should create unstructured object with correct GVK", func() {
			u := NewUnstructured("test.group", "TestKind", "v1")
			Expect(u).NotTo(BeNil())

			gvk := u.GetObjectKind().GroupVersionKind()
			Expect(gvk.Group).To(Equal("test.group"))
			Expect(gvk.Kind).To(Equal("TestKind"))
			Expect(gvk.Version).To(Equal("v1"))
		})
	})
})
