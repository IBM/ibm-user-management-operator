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
	"encoding/base64"
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

	Context("String Utility Functions", func() {
		It("should concatenate strings correctly", func() {
			result := Concat("hello", "-", "world", "!")
			Expect(result).To(Equal("hello-world!"))
		})

		It("should concatenate empty strings", func() {
			result := Concat("", "", "")
			Expect(result).To(Equal(""))
		})

		It("should concatenate single string", func() {
			result := Concat("single")
			Expect(result).To(Equal("single"))
		})
	})

	Context("Random String Functions", func() {
		It("should generate random strings of specified lengths", func() {
			lengths := []int{8, 16, 32}
			results, err := RandStrings(lengths...)

			Expect(err).NotTo(HaveOccurred())
			Expect(results).To(HaveLen(3))

			for _, result := range results {
				Expect(result).NotTo(BeEmpty())
				// Verify it's base64 encoded (double encoded in this case)
				_, err := base64.StdEncoding.DecodeString(string(result))
				Expect(err).NotTo(HaveOccurred())
			}
		})

		It("should handle zero lengths", func() {
			results, err := RandStrings(0, 0)
			Expect(err).NotTo(HaveOccurred())
			Expect(results).To(HaveLen(2))
		})

		It("should handle empty input", func() {
			results, err := RandStrings()
			Expect(err).NotTo(HaveOccurred())
			Expect(results).To(HaveLen(0))
		})
	})

	Context("Data Combination Functions", func() {
		type TestStruct1 struct {
			Field1 string
			Field2 int
		}

		type TestStruct2 struct {
			Field3 bool
			Field4 float64
		}

		It("should combine data from multiple structs", func() {
			s1 := TestStruct1{Field1: "test", Field2: 42}
			s2 := TestStruct2{Field3: true, Field4: 3.14}

			result := CombineData(s1, s2)

			Expect(result).To(HaveKey("Field1"))
			Expect(result).To(HaveKey("Field2"))
			Expect(result).To(HaveKey("Field3"))
			Expect(result).To(HaveKey("Field4"))
			Expect(result["Field1"]).To(Equal("test"))
			Expect(result["Field2"]).To(Equal(42))
			Expect(result["Field3"]).To(Equal(true))
			Expect(result["Field4"]).To(Equal(3.14))
		})

		It("should handle pointer structs", func() {
			s1 := &TestStruct1{Field1: "test", Field2: 42}
			result := CombineData(s1)

			Expect(result).To(HaveKey("Field1"))
			Expect(result["Field1"]).To(Equal("test"))
		})

		It("should handle non-struct values gracefully", func() {
			result := CombineData("not a struct", 123, []string{"slice"})
			Expect(result).To(BeEmpty())
		})
	})

	Context("Certificate Indentation Functions", func() {
		It("should indent certificate correctly", func() {
			cert := "-----BEGIN CERTIFICATE-----\nMIIC...\n-----END CERTIFICATE-----"
			indented := IndentCert(cert, 4)

			lines := []string{
				"    -----BEGIN CERTIFICATE-----",
				"    MIIC...",
				"    -----END CERTIFICATE-----",
			}
			expected := lines[0] + "\n" + lines[1] + "\n" + lines[2]
			Expect(indented).To(Equal(expected))
		})

		It("should handle zero indentation", func() {
			cert := "test\ncert"
			result := IndentCert(cert, 0)
			Expect(result).To(Equal("test\ncert"))
		})

		It("should handle single line", func() {
			cert := "single line"
			result := IndentCert(cert, 2)
			Expect(result).To(Equal("  single line"))
		})
	})
})
