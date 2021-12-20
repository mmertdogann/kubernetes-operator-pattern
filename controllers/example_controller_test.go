package controllers

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	examplev1 "mmertdogann/example-operator/api/v1"
)

var _ = Describe("Example controller", func() {

	// Define utility constants for object names and testing timeouts/durations and intervals.
	const (
		ExampleName      = "test-example"
		ExampleNamespace = "default"
		timeout          = time.Second * 10
		interval         = time.Millisecond * 250
	)

	Context("When creating Example", func() {
		It("Should check CR specs", func() {
			By("By creating a new Example")
			ctx := context.Background()
			example := &examplev1.Example{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "example.example.com/v1",
					Kind:       "Example",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      ExampleName,
					Namespace: ExampleNamespace,
				},
				Spec: examplev1.ExampleSpec{
					Name: "test-nginx",
					Size: int32(2),
				},
			}
			Expect(k8sClient.Create(ctx, example)).Should(Succeed())

			exampleLookupKey := types.NamespacedName{Name: ExampleName, Namespace: ExampleNamespace}
			createdExample := &examplev1.Example{}

			// We'll need to retry getting this newly created Example, given that creation may not immediately happen.
			Eventually(func() bool {
				err := k8sClient.Get(ctx, exampleLookupKey, createdExample)
				if err != nil {
					return false
				}
				return true
			}, timeout, interval).Should(BeTrue())
			// Let's make sure our Schedule string value was properly converted/handled.
			Expect(createdExample.Spec.Name).Should(Equal("test-nginx"))
			Expect(createdExample.Spec.Size).Should(Equal(int32(2)))

		})
	})
})
