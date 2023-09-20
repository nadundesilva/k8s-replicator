/*
 * Copyright (c) 2023, Nadun De Silva. All Rights Reserved.
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

package replication

import (
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func newConfigMapReplicator() *configMapReplicator {
	return &configMapReplicator{}
}

type configMapReplicator struct{}

func (r *configMapReplicator) GetKind() string {
	return "ConfigMap"
}

func (r *configMapReplicator) EmptyObject() client.Object {
	return &corev1.ConfigMap{}
}

func (r *configMapReplicator) EmptyObjectList() client.ObjectList {
	return &corev1.ConfigMapList{}
}

func (r *configMapReplicator) ObjectListToArray(list client.ObjectList) []client.Object {
	array := []client.Object{}
	configMaps := list.(*corev1.ConfigMapList).Items
	for i := range configMaps {
		configMap := configMaps[i]
		array = append(array, &configMap)
	}
	return array
}

func (r *configMapReplicator) Replicate(sourceObject client.Object, targetObject client.Object) {
	sourceConfigMap := sourceObject.(*corev1.ConfigMap)
	targetConfigMap := targetObject.(*corev1.ConfigMap)

	targetConfigMap.Immutable = sourceConfigMap.Immutable
	targetConfigMap.Data = sourceConfigMap.Data
	targetConfigMap.BinaryData = sourceConfigMap.BinaryData
}
