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

func newRoleReplicator() *roleReplicator {
	return &roleReplicator{}
}

type roleReplicator struct{}

func (r *roleReplicator) GetKind() string {
	return "Role"
}

func (r *roleReplicator) AddToScheme(scheme *runtime.Scheme) error {
	return rbacv1.AddToScheme(scheme)
}

func (r *roleReplicator) EmptyObject() client.Object {
	return &rbacv1.Role{}
}

func (r *roleReplicator) EmptyObjectList() client.ObjectList {
	return &rbacv1.RoleList{}
}

func (r *roleReplicator) ObjectListToArray(list client.ObjectList) []client.Object {
	array := []client.Object{}
	roles := list.(*rbacv1.RoleList).Items
	for i := range roles {
		array = append(array, &roles[i])
	}
	return array
}

func (r *roleReplicator) Replicate(sourceObject client.Object, targetObject client.Object) {
	sourceRole := sourceObject.(*rbacv1.Role)
	targetRole := targetObject.(*rbacv1.Role)

	// Copy Role-specific fields
	targetRole.Rules = []rbacv1.PolicyRule{}
	for _, rule := range sourceRole.Rules {
		targetRole.Rules = append(targetRole.Rules, *rule.DeepCopy())
	}
}
