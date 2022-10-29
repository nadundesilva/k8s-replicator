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
	"encoding/base64"
	"fmt"
	"strings"

	"reflect"
	"time"

	"context"

	"github.com/google/uuid"
	. "github.com/onsi/ginkgo"
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

var _ = Describe("Secret Replication", func() {
	var sourceNamespace *corev1.Namespace

	BeforeEach(func() {
		sourceNamespace = &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: "source-ns-" + uuid.New().String(),
			},
		}
		Expect(k8sClient.Create(context.Background(), sourceNamespace.DeepCopy())).Should(Succeed())
	})

	AfterEach(func() {
		Expect(k8sClient.Delete(context.Background(), sourceNamespace.DeepCopy())).Should(Succeed())
		sourceNamespace = nil
	})

	Context("When creating secret", func() {
		var sourceSecret *corev1.Secret
		BeforeEach(func() {
			sourceSecret = &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: sourceNamespace.GetName(),
					Name:      "test-secret-" + uuid.New().String(),
					Labels: map[string]string{
						groupFqn + "/unit-test-label-key": "test-label-value",
						ObjectTypeLabelKey:                ObjectTypeLabelValueSource,
					},
					Annotations: map[string]string{
						groupFqn + "/unit-test-annotation-key": "test-annotation-value",
					},
				},
				Type: corev1.SecretTypeOpaque,
				Data: map[string][]byte{
					"secret-data-item-one-key": []byte(base64.StdEncoding.EncodeToString([]byte("secret-data-item-one-value"))),
				},
			}
		})

		AfterEach(func() {
			Expect(k8sClient.Delete(context.Background(), sourceSecret.DeepCopy())).Should(Succeed())
			sourceSecret = nil
		})

		It("Should replicate to all normal namespaces", func() {
			ctx := context.Background()

			targetNamespaces := createNamespaces(ctx, "test-ns", 5, nil)
			Expect(k8sClient.Create(ctx, sourceSecret.DeepCopy())).Should(Succeed())

			validateReplication(ctx, sourceSecret, targetNamespaces...)
		})

		It("Should not change source secret", func() {
			ctx := context.Background()
			normalNamespaces := createNamespaces(ctx, "test-ns", 3, nil)
			Expect(k8sClient.Create(ctx, sourceSecret.DeepCopy())).Should(Succeed())

			validateReplication(ctx, sourceSecret, normalNamespaces...)
			Consistently(func() bool {
				finalSourceSecret := &corev1.Secret{}
				err := k8sClient.Get(ctx, client.ObjectKeyFromObject(sourceSecret), finalSourceSecret)
				if err != nil {
					return errors.IsNotFound(err)
				}
				return reflect.DeepEqual(sourceSecret.GetLabels(), finalSourceSecret.GetLabels()) &&
					reflect.DeepEqual(sourceSecret.GetAnnotations(), finalSourceSecret.GetAnnotations()) &&
					reflect.DeepEqual(sourceSecret.Type, finalSourceSecret.Type) &&
					reflect.DeepEqual(sourceSecret.Immutable, finalSourceSecret.Immutable) &&
					reflect.DeepEqual(sourceSecret.Data, finalSourceSecret.Data) &&
					reflect.DeepEqual(sourceSecret.StringData, finalSourceSecret.StringData)
			}, duration, interval).Should(BeTrue())
		})

		It("Should not replicate to kube prefixed namespaces", func() {
			ctx := context.Background()
			normalNamespaces := createNamespaces(ctx, "test-ns", 3, nil)

			kubeNamespaces := createNamespaces(ctx, "kube", 2, nil)
			Expect(k8sClient.Create(ctx, sourceSecret.DeepCopy())).Should(Succeed())

			validateReplication(ctx, sourceSecret, normalNamespaces...)
			validateNoReplication(ctx, sourceSecret, kubeNamespaces...)
		})

		It("Should replicate to kube prefixed namespaces with managed label", func() {
			ctx := context.Background()
			normalNamespaces := createNamespaces(ctx, "test-ns", 3, nil)

			labels := map[string]string{
				NamespaceTypeLabelKey: NamespaceTypeLabelValueManaged,
			}
			kubeNamespaces := createNamespaces(ctx, "kube", 2, labels)
			Expect(k8sClient.Create(ctx, sourceSecret.DeepCopy())).Should(Succeed())

			validateReplication(ctx, sourceSecret, normalNamespaces...)
			validateReplication(ctx, sourceSecret, kubeNamespaces...)
		})

		It("Should not replicate to namespaces with ignored label", func() {
			ctx := context.Background()
			normalNamespaces := createNamespaces(ctx, "test-ns", 3, nil)

			labels := map[string]string{
				NamespaceTypeLabelKey: NamespaceTypeLabelValueIgnored,
			}
			ignoredNamespaces := createNamespaces(ctx, "test-ns", 2, labels)
			Expect(k8sClient.Create(ctx, sourceSecret.DeepCopy())).Should(Succeed())

			validateReplication(ctx, sourceSecret, normalNamespaces...)
			validateNoReplication(ctx, sourceSecret, ignoredNamespaces...)
		})

		It("Should not replicate to operator namespace", func() {
			ctx := context.Background()
			normalNamespaces := createNamespaces(ctx, "test-ns", 3, nil)

			operatorNamespace = createNamespaces(ctx, "operator-ns", 1, nil)[0]
			Expect(k8sClient.Create(ctx, sourceSecret.DeepCopy())).Should(Succeed())

			validateReplication(ctx, sourceSecret, normalNamespaces...)
			validateNoReplication(ctx, sourceSecret, operatorNamespace)
			operatorNamespace = ""
		})

		It("Should replicate to operator namespace if it has managed label", func() {
			ctx := context.Background()
			normalNamespaces := createNamespaces(ctx, "test-ns", 3, nil)

			labels := map[string]string{
				NamespaceTypeLabelKey: NamespaceTypeLabelValueManaged,
			}
			operatorNamespace = createNamespaces(ctx, "operator-ns", 1, labels)[0]
			Expect(k8sClient.Create(ctx, sourceSecret.DeepCopy())).Should(Succeed())

			validateReplication(ctx, sourceSecret, normalNamespaces...)
			validateReplication(ctx, sourceSecret, operatorNamespace)
			operatorNamespace = ""
		})
	})
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

func validateNoReplication(ctx context.Context, sourceSecret *corev1.Secret, targetNamespaces ...string) {
	for _, ns := range targetNamespaces {
		lookupKey := client.ObjectKey{
			Namespace: ns,
			Name:      sourceSecret.GetName(),
		}
		Consistently(func() bool {
			replicatedSecret := &corev1.Secret{}
			err := k8sClient.Get(ctx, lookupKey, replicatedSecret)
			return err != nil && errors.IsNotFound(err)
		}, timeout, interval).Should(BeTrue())
	}
}

func validateReplication(ctx context.Context, sourceSecret *corev1.Secret, targetNamespaces ...string) {
	for _, ns := range targetNamespaces {
		lookupKey := client.ObjectKey{
			Namespace: ns,
			Name:      sourceSecret.GetName(),
		}
		Eventually(func() bool {
			replicatedSecret := &corev1.Secret{}
			err := k8sClient.Get(ctx, lookupKey, replicatedSecret)
			if err != nil {
				return false
			}

			objectType, objectTypeOk := replicatedSecret.GetLabels()[ObjectTypeLabelKey]
			Expect(objectTypeOk).Should(BeTrue())
			Expect(objectType).Should(Equal(ObjectTypeLabelValueReplica))

			sourceNamespace, sourceNamespaceOk := replicatedSecret.GetAnnotations()[SourceNamespaceAnnotationKey]
			Expect(sourceNamespaceOk).Should(BeTrue())
			Expect(sourceNamespace).Should(Equal(sourceSecret.GetNamespace()))

			return isMapsEqualWithoutReplicatorKeys(sourceSecret.GetLabels(), replicatedSecret.GetLabels()) &&
				isMapsEqualWithoutReplicatorKeys(sourceSecret.GetAnnotations(), replicatedSecret.GetAnnotations()) &&
				reflect.DeepEqual(sourceSecret.Type, replicatedSecret.Type) &&
				reflect.DeepEqual(sourceSecret.Immutable, replicatedSecret.Immutable) &&
				reflect.DeepEqual(sourceSecret.Data, replicatedSecret.Data) &&
				reflect.DeepEqual(sourceSecret.StringData, replicatedSecret.StringData)
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
