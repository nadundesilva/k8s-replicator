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

//+kubebuilder:rbac:groups="",resources=serviceaccounts,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="",resources=serviceaccounts/status,verbs=get;update;patch
//+kubebuilder:rbac:groups="",resources=serviceaccounts/finalizers,verbs=update

func newServiceAccountReplicator() *serviceAccountReplicator {
	return &serviceAccountReplicator{}
}

type serviceAccountReplicator struct{}

func (r *serviceAccountReplicator) GetKind() string {
	return "ServiceAccount"
}

func (r *serviceAccountReplicator) AddToScheme(scheme *runtime.Scheme) error {
	return corev1.AddToScheme(scheme)
}

func (r *serviceAccountReplicator) EmptyObject() client.Object {
	return &corev1.ServiceAccount{}
}

func (r *serviceAccountReplicator) EmptyObjectList() client.ObjectList {
	return &corev1.ServiceAccountList{}
}

func (r *serviceAccountReplicator) ObjectListToArray(list client.ObjectList) []client.Object {
	serviceAccounts := list.(*corev1.ServiceAccountList).Items
	array := make([]client.Object, len(serviceAccounts))
	for i := range serviceAccounts {
		array[i] = &serviceAccounts[i]
	}
	return array
}

func (r *serviceAccountReplicator) Replicate(sourceObject client.Object, targetObject client.Object) {
	sourceServiceAccount := sourceObject.(*corev1.ServiceAccount)
	targetServiceAccount := targetObject.(*corev1.ServiceAccount)

	// Copy ServiceAccount-specific fields
	targetServiceAccount.Secrets = make([]corev1.ObjectReference, len(sourceServiceAccount.Secrets))
	for i := range sourceServiceAccount.Secrets {
		sourceServiceAccount.Secrets[i].DeepCopyInto(&targetServiceAccount.Secrets[i])
	}
	targetServiceAccount.ImagePullSecrets = make([]corev1.LocalObjectReference, len(sourceServiceAccount.ImagePullSecrets))
	for i := range sourceServiceAccount.ImagePullSecrets {
		sourceServiceAccount.ImagePullSecrets[i].DeepCopyInto(&targetServiceAccount.ImagePullSecrets[i])
	}
	targetServiceAccount.AutomountServiceAccountToken = sourceServiceAccount.AutomountServiceAccountToken
}
