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
package controller

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/google/uuid"
	. "github.com/nadundesilva/k8s-replicator/test/utils/gomega"
	"github.com/nadundesilva/k8s-replicator/test/utils/testdata"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	testTimeout           = SpecTimeout(time.Minute)
	assertionTimeout      = time.Second * 10
	assertionPollInterval = time.Second
)

var _ = Describe("Object Replication", func() {
	for _, resource := range testdata.GenerateResourceTestData() {
		Describe(fmt.Sprintf("Resource %s", resource.Name), func() {
			var sourceNamespace *corev1.Namespace
			var sourceObject client.Object
			nc := namespaceCreator{}

			BeforeEach(func(ctx SpecContext) {
				sourceNamespace = &corev1.Namespace{
					ObjectMeta: metav1.ObjectMeta{
						Name: "source-ns-" + uuid.New().String(),
					},
				}
				Expect(k8sClient.Create(ctx, sourceNamespace)).To(Succeed())

				sourceObject = resource.SourceObject()
				sourceObject.SetNamespace(sourceNamespace.GetName())
				sourceObject.GetLabels()[objectTypeLabelKey] = objectTypeLabelValueReplicated
			})

			AfterEach(func(ctx SpecContext) {
				Expect(k8sClient.Delete(ctx, sourceObject)).To(Succeed())
				sourceObject = nil

				deleteNamespace(ctx, sourceNamespace)
				sourceNamespace = nil

				nc.Cleanup(ctx)
			})

			Context("When creating object", func() {
				It("Should replicate to all normal namespaces", func(ctx SpecContext) {
					targetNamespaces := nc.CreateNamespaces(ctx, "test-ns", 5, nil)
					Expect(k8sClient.Create(ctx, sourceObject)).To(Succeed())

					validateReplication(ctx, sourceObject, resource, targetNamespaces...)
				}, testTimeout)

				It("Should not change source object", func(ctx SpecContext) {
					normalNamespaces := nc.CreateNamespaces(ctx, "test-ns", 3, nil)
					Expect(k8sClient.Create(ctx, sourceObject)).To(Succeed())
					validateReplication(ctx, sourceObject, resource, normalNamespaces...)

					Consistently(func() map[string]string {
						finalSourceObject := resource.EmptyObject()
						err := k8sClient.Get(ctx, client.ObjectKeyFromObject(sourceObject), finalSourceObject)
						if err != nil {
							return sourceObject.GetLabels()
						}
						return finalSourceObject.GetLabels()
					}, assertionTimeout, assertionPollInterval, ctx).Should(BeEquivalentTo(sourceObject.GetLabels()))

					Consistently(func() map[string]string {
						finalSourceObject := resource.EmptyObject()
						err := k8sClient.Get(ctx, client.ObjectKeyFromObject(sourceObject), finalSourceObject)
						if err != nil {
							return sourceObject.GetAnnotations()
						}
						return finalSourceObject.GetAnnotations()
					}, assertionTimeout, assertionPollInterval, ctx).Should(BeEquivalentTo(sourceObject.GetAnnotations()))

					Consistently(func() client.Object {
						finalSourceObject := resource.EmptyObject()
						err := k8sClient.Get(ctx, client.ObjectKeyFromObject(sourceObject), finalSourceObject)
						if err != nil {
							return sourceObject
						}
						return finalSourceObject
					}, assertionTimeout, assertionPollInterval, ctx).Should(BeEquivalentToResource(sourceObject, resource.IsEqual))
				}, testTimeout)

				It("Should not replicate to kube prefixed namespaces", func(ctx SpecContext) {
					normalNamespaces := nc.CreateNamespaces(ctx, "test-ns", 3, nil)

					kubeNamespaces := nc.CreateNamespaces(ctx, "kube", 2, nil)
					Expect(k8sClient.Create(ctx, sourceObject)).To(Succeed())

					validateReplication(ctx, sourceObject, resource, normalNamespaces...)
					validateNoReplication(ctx, sourceObject, resource, kubeNamespaces...)
				}, testTimeout)

				It("Should not replicate to operator namespace", func(ctx SpecContext) {
					normalNamespaces := nc.CreateNamespaces(ctx, "test-ns", 3, nil)

					operatorNs := nc.CreateNamespaces(ctx, "operator-ns", 1, nil)[0]
					operatorNamespace = operatorNs.GetName()
					Expect(k8sClient.Create(ctx, sourceObject)).To(Succeed())

					validateReplication(ctx, sourceObject, resource, normalNamespaces...)
					validateNoReplication(ctx, sourceObject, resource, operatorNs)
					operatorNamespace = ""
				}, testTimeout)

				It("Should not replicate to namespaces with ignored label", func(ctx SpecContext) {
					normalNamespaces := nc.CreateNamespaces(ctx, "test-ns", 3, nil)

					labels := map[string]string{
						namespaceTypeLabelKey: namespaceTypeLabelValueIgnored,
					}
					ignoredNamespaces := nc.CreateNamespaces(ctx, "test-ns", 2, labels)
					Expect(k8sClient.Create(ctx, sourceObject)).To(Succeed())

					validateReplication(ctx, sourceObject, resource, normalNamespaces...)
					validateNoReplication(ctx, sourceObject, resource, ignoredNamespaces...)
				}, testTimeout)

				It("Should replicate to kube prefixed namespaces with managed label", func(ctx SpecContext) {
					normalNamespaces := nc.CreateNamespaces(ctx, "test-ns", 3, nil)

					labels := map[string]string{
						namespaceTypeLabelKey: namespaceTypeLabelValueManaged,
					}
					kubeNamespaces := nc.CreateNamespaces(ctx, "kube", 2, labels)
					Expect(k8sClient.Create(ctx, sourceObject)).To(Succeed())

					validateReplication(ctx, sourceObject, resource, normalNamespaces...)
					validateReplication(ctx, sourceObject, resource, kubeNamespaces...)
				}, testTimeout)

				It("Should replicate to operator namespace if it has managed label", func(ctx SpecContext) {
					normalNamespaces := nc.CreateNamespaces(ctx, "test-ns", 3, nil)

					labels := map[string]string{
						namespaceTypeLabelKey: namespaceTypeLabelValueManaged,
					}
					operatorNs := nc.CreateNamespaces(ctx, "operator-ns", 1, labels)[0]
					operatorNamespace = operatorNs.GetName()
					Expect(k8sClient.Create(ctx, sourceObject)).To(Succeed())

					validateReplication(ctx, sourceObject, resource, normalNamespaces...)
					validateReplication(ctx, sourceObject, resource, operatorNs)
					operatorNamespace = ""
				}, testTimeout)
			})
		})
	}
})

type namespaceCreator struct {
	testNamespaces []*corev1.Namespace
}

func (nc *namespaceCreator) CreateNamespaces(ctx context.Context, prefix string, count int, labels map[string]string) []*corev1.Namespace {
	namespaces := []*corev1.Namespace{}
	for i := 0; i < count; i++ {
		namespaceName := fmt.Sprintf("%s-%s", prefix, uuid.New().String())
		ns := &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name:   namespaceName,
				Labels: labels,
			},
		}
		Expect(k8sClient.Create(ctx, ns)).To(Succeed())
		namespaces = append(namespaces, ns)
	}
	nc.testNamespaces = append(nc.testNamespaces, namespaces...)
	return namespaces
}

func (nc *namespaceCreator) Cleanup(ctx context.Context) {
	for _, ns := range nc.testNamespaces {
		deleteNamespace(ctx, ns)
	}
}

func deleteNamespace(ctx context.Context, ns *corev1.Namespace) {
	Expect(k8sClient.Delete(ctx, ns)).To(Succeed())
}

func validateNoReplication(ctx context.Context, sourceObject client.Object, resource testdata.Resource, targetNamespaces ...*corev1.Namespace) {
	for _, ns := range targetNamespaces {
		lookupKey := client.ObjectKey{
			Namespace: ns.GetName(),
			Name:      sourceObject.GetName(),
		}
		Consistently(func() bool {
			err := k8sClient.Get(ctx, lookupKey, resource.EmptyObject())
			return err != nil && errors.IsNotFound(err)
		}, assertionTimeout, assertionPollInterval, ctx).Should(BeTrue())
	}
}

func validateReplication(ctx context.Context, sourceObject client.Object, resource testdata.Resource, targetNamespaces ...*corev1.Namespace) {
	for _, ns := range targetNamespaces {
		lookupKey := client.ObjectKey{
			Namespace: ns.GetName(),
			Name:      sourceObject.GetName(),
		}
		Eventually(func() bool {
			replicatedObject := resource.EmptyObject()
			err := k8sClient.Get(ctx, lookupKey, replicatedObject)
			if err != nil {
				return false
			}

			objectType, objectTypeOk := replicatedObject.GetLabels()[objectTypeLabelKey]
			Expect(objectTypeOk).To(BeTrue())
			Expect(objectType).To(Equal(objectTypeLabelValueReplica))

			sourceNamespace, sourceNamespaceOk := replicatedObject.GetAnnotations()[sourceNamespaceAnnotationKey]
			Expect(sourceNamespaceOk).To(BeTrue())
			Expect(sourceNamespace).To(Equal(sourceObject.GetNamespace()))

			return isMapsEqualWithoutReplicatorKeys(sourceObject.GetLabels(), replicatedObject.GetLabels()) &&
				isMapsEqualWithoutReplicatorKeys(sourceObject.GetAnnotations(), replicatedObject.GetAnnotations()) &&
				resource.IsEqual(sourceObject, replicatedObject)
		}, assertionTimeout, assertionPollInterval, ctx).Should(BeTrue())
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
