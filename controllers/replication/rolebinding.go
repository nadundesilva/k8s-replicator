/*
 * Copyright (c) 2025, Nadun De Silva. All Rights Reserved.
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
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

//+kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=rolebindings,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=rolebindings/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=rolebindings/finalizers,verbs=update

func newRoleBindingReplicator() *roleBindingReplicator {
	return &roleBindingReplicator{}
}

type roleBindingReplicator struct{}

func (r *roleBindingReplicator) GetKind() string {
	return "RoleBinding"
}

func (r *roleBindingReplicator) AddToScheme(scheme *runtime.Scheme) error {
	return rbacv1.AddToScheme(scheme)
}

func (r *roleBindingReplicator) EmptyObject() client.Object {
	return &rbacv1.RoleBinding{}
}

func (r *roleBindingReplicator) EmptyObjectList() client.ObjectList {
	return &rbacv1.RoleBindingList{}
}

func (r *roleBindingReplicator) ObjectListToArray(list client.ObjectList) []client.Object {
	roleBindings := list.(*rbacv1.RoleBindingList).Items
	array := make([]client.Object, len(roleBindings))
	for i := range roleBindings {
		array[i] = &roleBindings[i]
	}
	return array
}

func (r *roleBindingReplicator) Replicate(sourceObject client.Object, targetObject client.Object) {
	sourceRoleBinding := sourceObject.(*rbacv1.RoleBinding)
	targetRoleBinding := targetObject.(*rbacv1.RoleBinding)

	// Copy RoleBinding-specific fields
	sourceRoleBinding.RoleRef.DeepCopyInto(&targetRoleBinding.RoleRef)
	targetRoleBinding.Subjects = make([]rbacv1.Subject, len(sourceRoleBinding.Subjects))
	for i := range sourceRoleBinding.Subjects {
		sourceRoleBinding.Subjects[i].DeepCopyInto(&targetRoleBinding.Subjects[i])
	}
}
