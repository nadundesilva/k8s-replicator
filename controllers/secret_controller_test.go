/*
 * Copyright (c) 2022, Nadun De Silva. All Rights Reserved.
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *   http://www.apache.org/licenses/LICENSE-2.0
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */
package controllers

import (
	"fmt"
	"strings"

	"reflect"
	"time"

	"context"

	"github.com/google/uuid"
	"github.com/nadundesilva/k8s-replicator/test/utils/testdata"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	timeout  = time.Second * 10
	duration = time.Second * 10
	interval = time.Millisecond * 250
)

var _ = Describe("Object Replication", func() {
	for _, resource := range testdata.GenerateResourceTestData() {
		Describe(fmt.Sprintf("Resource %s", resource.Name), func() {
			var sourceNamespace *corev1.Namespace
			var sourceObject client.Object

			BeforeEach(func() {
				sourceNamespace = &corev1.Namespace{
					ObjectMeta: metav1.ObjectMeta{
						Name: "source-ns-" + uuid.New().String(),
					},
				}
				Expect(k8sClient.Create(context.Background(), sourceNamespace)).Should(Succeed())

				sourceObject = resource.SourceObject()
				sourceObject.SetNamespace(sourceNamespace.GetName())
				sourceObject.GetLabels()[ObjectTypeLabelKey] = ObjectTypeLabelValueReplicated
			})

			AfterEach(func() {
				Expect(k8sClient.Delete(context.Background(), sourceNamespace)).Should(Succeed())
				sourceNamespace = nil

				Expect(k8sClient.Delete(context.Background(), sourceObject)).Should(Succeed())
				sourceObject = nil
			})

			Context("When creating object", func() {
				It("Should replicate to all normal namespaces", func() {
					ctx := context.Background()

					targetNamespaces := createNamespaces(ctx, "test-ns", 5, nil)
					Expect(k8sClient.Create(ctx, sourceObject)).Should(Succeed())

					validateReplication(ctx, sourceObject, resource, targetNamespaces...)
				})

				It("Should not change source object", func() {
					ctx := context.Background()
					normalNamespaces := createNamespaces(ctx, "test-ns", 3, nil)
					Expect(k8sClient.Create(ctx, sourceObject)).Should(Succeed())

					validateReplication(ctx, sourceObject, resource, normalNamespaces...)
					Consistently(func() bool {
						finalSourceObject := resource.EmptyObject()
						err := k8sClient.Get(ctx, client.ObjectKeyFromObject(sourceObject), finalSourceObject)
						if err != nil {
							return errors.IsNotFound(err)
						}
						return reflect.DeepEqual(sourceObject.GetLabels(), finalSourceObject.GetLabels()) &&
							reflect.DeepEqual(sourceObject.GetAnnotations(), finalSourceObject.GetAnnotations()) &&
							resource.IsEqual(sourceObject, finalSourceObject)
					}, duration, interval).Should(BeTrue())
				})

				It("Should not replicate to kube prefixed namespaces", func() {
					ctx := context.Background()
					normalNamespaces := createNamespaces(ctx, "test-ns", 3, nil)

					kubeNamespaces := createNamespaces(ctx, "kube", 2, nil)
					Expect(k8sClient.Create(ctx, sourceObject)).Should(Succeed())

					validateReplication(ctx, sourceObject, resource, normalNamespaces...)
					validateNoReplication(ctx, sourceObject, resource, kubeNamespaces...)
				})

				It("Should replicate to kube prefixed namespaces with managed label", func() {
					ctx := context.Background()
					normalNamespaces := createNamespaces(ctx, "test-ns", 3, nil)

					labels := map[string]string{
						NamespaceTypeLabelKey: NamespaceTypeLabelValueManaged,
					}
					kubeNamespaces := createNamespaces(ctx, "kube", 2, labels)
					Expect(k8sClient.Create(ctx, sourceObject)).Should(Succeed())

					validateReplication(ctx, sourceObject, resource, normalNamespaces...)
					validateReplication(ctx, sourceObject, resource, kubeNamespaces...)
				})

				It("Should not replicate to namespaces with ignored label", func() {
					ctx := context.Background()
					normalNamespaces := createNamespaces(ctx, "test-ns", 3, nil)

					labels := map[string]string{
						NamespaceTypeLabelKey: NamespaceTypeLabelValueIgnored,
					}
					ignoredNamespaces := createNamespaces(ctx, "test-ns", 2, labels)
					Expect(k8sClient.Create(ctx, sourceObject)).Should(Succeed())

					validateReplication(ctx, sourceObject, resource, normalNamespaces...)
					validateNoReplication(ctx, sourceObject, resource, ignoredNamespaces...)
				})

				It("Should not replicate to operator namespace", func() {
					ctx := context.Background()
					normalNamespaces := createNamespaces(ctx, "test-ns", 3, nil)

					operatorNamespace = createNamespaces(ctx, "operator-ns", 1, nil)[0]
					Expect(k8sClient.Create(ctx, sourceObject)).Should(Succeed())

					validateReplication(ctx, sourceObject, resource, normalNamespaces...)
					validateNoReplication(ctx, sourceObject, resource, operatorNamespace)
					operatorNamespace = ""
				})

				It("Should replicate to operator namespace if it has managed label", func() {
					ctx := context.Background()
					normalNamespaces := createNamespaces(ctx, "test-ns", 3, nil)

					labels := map[string]string{
						NamespaceTypeLabelKey: NamespaceTypeLabelValueManaged,
					}
					operatorNamespace = createNamespaces(ctx, "operator-ns", 1, labels)[0]
					Expect(k8sClient.Create(ctx, sourceObject)).Should(Succeed())

					validateReplication(ctx, sourceObject, resource, normalNamespaces...)
					validateReplication(ctx, sourceObject, resource, operatorNamespace)
					operatorNamespace = ""
				})
			})
		})
	}
})

func createNamespaces(ctx context.Context, prefix string, count int, labels map[string]string) []string {
	namespaces := []string{}
	for i := 0; i < count; i++ {
		namespaceName := fmt.Sprintf("%s-%s", prefix, uuid.New().String())
		ns := corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name:   namespaceName,
				Labels: labels,
			},
		}
		Expect(k8sClient.Create(ctx, &ns)).Should(Succeed())
		namespaces = append(namespaces, namespaceName)
	}
	return namespaces
}

func validateNoReplication(ctx context.Context, sourceObject client.Object, resource testdata.Resource, targetNamespaces ...string) {
	for _, ns := range targetNamespaces {
		lookupKey := client.ObjectKey{
			Namespace: ns,
			Name:      sourceObject.GetName(),
		}
		Consistently(func() bool {
			err := k8sClient.Get(ctx, lookupKey, resource.EmptyObject())
			return err != nil && errors.IsNotFound(err)
		}, timeout, interval).Should(BeTrue())
	}
}

func validateReplication(ctx context.Context, sourceObject client.Object, resource testdata.Resource, targetNamespaces ...string) {
	for _, ns := range targetNamespaces {
		lookupKey := client.ObjectKey{
			Namespace: ns,
			Name:      sourceObject.GetName(),
		}
		Eventually(func() bool {
			replicatedObject := resource.EmptyObject()
			err := k8sClient.Get(ctx, lookupKey, replicatedObject)
			if err != nil {
				return false
			}

			objectType, objectTypeOk := replicatedObject.GetLabels()[ObjectTypeLabelKey]
			Expect(objectTypeOk).Should(BeTrue())
			Expect(objectType).Should(Equal(ObjectTypeLabelValueReplica))

			sourceNamespace, sourceNamespaceOk := replicatedObject.GetAnnotations()[SourceNamespaceAnnotationKey]
			Expect(sourceNamespaceOk).Should(BeTrue())
			Expect(sourceNamespace).Should(Equal(sourceObject.GetNamespace()))

			return isMapsEqualWithoutReplicatorKeys(sourceObject.GetLabels(), replicatedObject.GetLabels()) &&
				isMapsEqualWithoutReplicatorKeys(sourceObject.GetAnnotations(), replicatedObject.GetAnnotations()) &&
				resource.IsEqual(sourceObject, replicatedObject)
		}, timeout, interval).Should(BeTrue())
	}
}

func isMapsEqualWithoutReplicatorKeys(mapA map[string]string, mapB map[string]string) bool {
	removeReplicatorKeys := func(inputMap map[string]string) map[string]string {
		finalMap := map[string]string{}
		for k, v := range inputMap {
			if !strings.HasPrefix(k, groupFqn) {
				finalMap[k] = v
			}
		}
		return finalMap
	}
	sanitizedMapA := removeReplicatorKeys(mapA)
	sanitizedMapB := removeReplicatorKeys(mapB)
	return reflect.DeepEqual(sanitizedMapA, sanitizedMapB)
}
