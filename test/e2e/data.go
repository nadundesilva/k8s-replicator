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
package e2e

import (
	"encoding/base64"
	"reflect"
	"testing"

	"github.com/nadundesilva/k8s-replicator/pkg/replicator"
	"github.com/nadundesilva/k8s-replicator/test/utils/validation"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/e2e-framework/klient/k8s"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
)

type resourceTestDatum struct {
	name               string
	objectList         k8s.ObjectList
	sourceObject       k8s.Object
	sourceObjectUpdate k8s.Object
	matcher            validation.ObjectMatcher
}

func generateResourcesCreationTestData(t *testing.T) []resourceTestDatum {
	resources := []resourceTestDatum{
		{
			name:       "Secret",
			objectList: &corev1.SecretList{},
			sourceObject: &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name: envconf.RandomName("test-secret", 32),
					Labels: map[string]string{
						"e2e-tests.replicator.io/test-label-key": "test-label-value",
					},
					Annotations: map[string]string{
						"e2e-tests.replicator.io/test-annotation-key": "test-annotation-value",
					},
				},
				Data: map[string][]byte{
					"secret-data-item-one-key": []byte(base64.StdEncoding.EncodeToString([]byte("secret-data-item-one-value"))),
				},
			},
			sourceObjectUpdate: &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"e2e-tests.replicator.io/test-label-key": "test-label-value",
					},
					Annotations: map[string]string{
						"e2e-tests.replicator.io/test-annotation-key": "test-annotation-value",
					},
				},
				Data: map[string][]byte{
					"secret-data-item-two-key": []byte(base64.StdEncoding.EncodeToString([]byte("secret-data-item-two-value"))),
				},
			},
			matcher: func(sourceObject k8s.Object, replicaObject k8s.Object) bool {
				sourceSecret := sourceObject.(*corev1.Secret)
				replicaSecret := replicaObject.(*corev1.Secret)
				if !reflect.DeepEqual(sourceSecret.Data, replicaSecret.Data) {
					t.Errorf("secret data not equal; want %s, got %s",
						sourceSecret.Data, replicaSecret.Data)
				}
				return true
			},
		},
		{
			name:       "ConfigMap",
			objectList: &corev1.ConfigMapList{},
			sourceObject: &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name: envconf.RandomName("test-config-map", 32),
					Labels: map[string]string{
						"e2e-tests.replicator.io/test-label-key": "test-label-value",
					},
					Annotations: map[string]string{
						"e2e-tests.replicator.io/test-annotation-key": "test-annotation-value",
					},
				},
				BinaryData: map[string][]byte{
					"config-map-data-item-one-key": []byte(base64.StdEncoding.EncodeToString([]byte("config-map-data-item-one-value"))),
				},
			},
			sourceObjectUpdate: &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"e2e-tests.replicator.io/test-label-key": "test-label-value",
					},
					Annotations: map[string]string{
						"e2e-tests.replicator.io/test-annotation-key": "test-annotation-value",
					},
				},
				BinaryData: map[string][]byte{
					"config-map-data-item-two-key": []byte(base64.StdEncoding.EncodeToString([]byte("config-map-data-item-two-value"))),
				},
			},
			matcher: func(sourceObject k8s.Object, replicaObject k8s.Object) bool {
				sourceConfigMap := sourceObject.(*corev1.ConfigMap)
				replicaConfigMap := replicaObject.(*corev1.ConfigMap)
				if !reflect.DeepEqual(sourceConfigMap.BinaryData, replicaConfigMap.BinaryData) {
					t.Errorf("config map data not equal; want %s, got %s",
						sourceConfigMap.BinaryData, replicaConfigMap.BinaryData)
				}
				return true
			},
		},
	}

	for _, resource := range resources {
		resource.sourceObjectUpdate.SetName(resource.sourceObject.GetName())

		updateSourceObjectLabels := func(sourceObject k8s.Object) {
			labels := sourceObject.GetLabels()
			labels[replicator.ObjectTypeLabelKey] = replicator.ObjectTypeLabelValueSource
		}
		updateSourceObjectLabels(resource.sourceObject)
		updateSourceObjectLabels(resource.sourceObjectUpdate)
	}
	return resources
}
