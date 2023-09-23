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
package testdata

import (
	"encoding/base64"
	"reflect"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
)

func generateConfigMapTestDatum() Resource {
	return process(resourceData{
		Name: "ConfigMap",
		SourceObject: &corev1.ConfigMap{
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
		SourceObjectUpdate: &corev1.ConfigMap{
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
		EmptyObject:     &corev1.ConfigMap{},
		EmptyObjectList: &corev1.ConfigMapList{},
		IsEqual: func(sourceObject client.Object, replicaObject client.Object) bool {
			sourceConfigMap := sourceObject.(*corev1.ConfigMap)
			replicaConfigMap := replicaObject.(*corev1.ConfigMap)
			return reflect.DeepEqual(sourceConfigMap.BinaryData, replicaConfigMap.BinaryData)
		},
	})
}
