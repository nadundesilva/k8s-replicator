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
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

//+kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="",resources=secrets/status,verbs=get;update;patch
//+kubebuilder:rbac:groups="",resources=secrets/finalizers,verbs=update

func newSecretReplicator() *secretReplicator {
	return &secretReplicator{}
}

type secretReplicator struct{}

func (r *secretReplicator) GetKind() string {
	return "Secret"
}

func (r *secretReplicator) AddToScheme(scheme *runtime.Scheme) error {
	return corev1.AddToScheme(scheme)
}

func (r *secretReplicator) EmptyObject() client.Object {
	return &corev1.Secret{}
}

func (r *secretReplicator) EmptyObjectList() client.ObjectList {
	return &corev1.SecretList{}
}

func (r *secretReplicator) ObjectListToArray(list client.ObjectList) []client.Object {
	secrets := list.(*corev1.SecretList).Items
	array := make([]client.Object, len(secrets))
	for i := range secrets {
		array[i] = &secrets[i]
	}
	return array
}

func (r *secretReplicator) Replicate(sourceObject client.Object, targetObject client.Object) {
	sourceSecret := sourceObject.(*corev1.Secret)
	targetSecret := targetObject.(*corev1.Secret)

	// Copy Secret-specific fields
	targetSecret.Immutable = sourceSecret.Immutable
	targetSecret.Data = sourceSecret.Data
	targetSecret.StringData = sourceSecret.StringData
	targetSecret.Type = sourceSecret.Type
}
