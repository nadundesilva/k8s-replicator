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
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

//+kubebuilder:rbac:groups="networking.k8s.io",resources=networkpolicies,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="networking.k8s.io",resources=networkpolicies/status,verbs=get;update;patch
//+kubebuilder:rbac:groups="networking.k8s.io",resources=networkpolicies/finalizers,verbs=update

func newNetworkPolicyReplicator() *networkPolicyReplicator {
	return &networkPolicyReplicator{}
}

type networkPolicyReplicator struct{}

func (r *networkPolicyReplicator) GetKind() string {
	return "NetworkPolicy"
}

func (r *networkPolicyReplicator) AddToScheme(scheme *runtime.Scheme) error {
	return networkingv1.AddToScheme(scheme)
}

func (r *networkPolicyReplicator) EmptyObject() client.Object {
	return &networkingv1.NetworkPolicy{}
}

func (r *networkPolicyReplicator) EmptyObjectList() client.ObjectList {
	return &networkingv1.NetworkPolicyList{}
}

func (r *networkPolicyReplicator) ObjectListToArray(list client.ObjectList) []client.Object {
	networkPolicies := list.(*networkingv1.NetworkPolicyList).Items
	array := make([]client.Object, len(networkPolicies))
	for i := range networkPolicies {
		array[i] = &networkPolicies[i]
	}
	return array
}

func (r *networkPolicyReplicator) Replicate(sourceObject client.Object, targetObject client.Object) {
	sourceNetworkPolicy := sourceObject.(*networkingv1.NetworkPolicy)
	targetNetworkPolicy := targetObject.(*networkingv1.NetworkPolicy)

	// Copy NetworkPolicy-specific fields
	sourceNetworkPolicy.Spec.DeepCopyInto(&targetNetworkPolicy.Spec)
}
