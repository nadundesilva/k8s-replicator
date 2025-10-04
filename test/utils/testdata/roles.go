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
package testdata

import (
	"fmt"
	"reflect"

	"github.com/google/uuid"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func generateRoleTestDatum() Resource {
	return process(resourceData{
		Name: "Role",
		SourceObject: &rbacv1.Role{
			ObjectMeta: metav1.ObjectMeta{
				Name: fmt.Sprintf("test-role-%s", uuid.New().String()),
				Labels: map[string]string{
					"e2e-tests.replicator.io/test-label-key": "test-label-value",
				},
				Annotations: map[string]string{
					"e2e-tests.replicator.io/test-annotation-key": "test-annotation-value",
				},
			},
			Rules: []rbacv1.PolicyRule{
				{
					APIGroups:     []string{""},
					Resources:     []string{"pods"},
					ResourceNames: []string{"test-pod-1", "test-pod-2"},
					Verbs:         []string{"get", "list", "watch"},
				},
				{
					APIGroups: []string{"apps"},
					Resources: []string{"deployments"},
					Verbs:     []string{"get", "list", "watch", "create", "update", "patch", "delete"},
				},
			},
		},
		SourceObjectUpdate: &rbacv1.Role{
			ObjectMeta: metav1.ObjectMeta{
				Labels: map[string]string{
					"e2e-tests.replicator.io/test-label-key": "test-label-value",
				},
				Annotations: map[string]string{
					"e2e-tests.replicator.io/test-annotation-key": "test-annotation-value",
				},
			},
			Rules: []rbacv1.PolicyRule{
				{
					APIGroups: []string{"apps"},
					Resources: []string{"deployments", "replicasets"},
					Verbs:     []string{"get", "list", "watch", "create", "update", "patch", "delete"},
				},
				{
					APIGroups:     []string{""},
					Resources:     []string{"pods", "services"},
					ResourceNames: []string{"test-pod-1"},
					Verbs:         []string{"get", "list", "watch"},
				},
			},
		},
		EmptyObject:     &rbacv1.Role{},
		EmptyObjectList: &rbacv1.RoleList{},
		IsEqual: func(sourceObject client.Object, replicaObject client.Object) bool {
			sourceRole := sourceObject.(*rbacv1.Role)
			replicaRole := replicaObject.(*rbacv1.Role)
			return reflect.DeepEqual(sourceRole.Rules, replicaRole.Rules)
		},
	})
}
