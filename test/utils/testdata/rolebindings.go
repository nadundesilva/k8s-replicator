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

func generateRoleBindingTestDatum() Resource {
	roleRef := rbacv1.RoleRef{
		APIGroup: "rbac.authorization.k8s.io",
		Kind:     "Role",
		Name:     "test-role",
	}
	return process(resourceData{
		Name: "RoleBinding",
		SourceObject: &rbacv1.RoleBinding{
			ObjectMeta: metav1.ObjectMeta{
				Name: fmt.Sprintf("test-rolebinding-%s", uuid.New().String()),
				Labels: map[string]string{
					"e2e-tests.replicator.io/test-label-key": "test-label-value",
				},
				Annotations: map[string]string{
					"e2e-tests.replicator.io/test-annotation-key": "test-annotation-value",
				},
			},
			RoleRef: roleRef,
			Subjects: []rbacv1.Subject{
				{
					Kind:      "ServiceAccount",
					APIGroup:  "",
					Name:      "test-service-account",
					Namespace: "test-namespace",
				},
				{
					Kind:      "User",
					APIGroup:  "rbac.authorization.k8s.io",
					Name:      "test-user",
					Namespace: "",
				},
				{
					Kind:      "Group",
					APIGroup:  "rbac.authorization.k8s.io",
					Name:      "test-group",
					Namespace: "",
				},
			},
		},
		SourceObjectUpdate: &rbacv1.RoleBinding{
			ObjectMeta: metav1.ObjectMeta{
				Labels: map[string]string{
					"e2e-tests.replicator.io/test-label-key": "test-label-value",
				},
				Annotations: map[string]string{
					"e2e-tests.replicator.io/test-annotation-key": "test-annotation-value",
				},
			},
			RoleRef: roleRef,
			Subjects: []rbacv1.Subject{
				{
					Kind:      "ServiceAccount",
					APIGroup:  "",
					Name:      "test-service-account-updated",
					Namespace: "test-namespace-updated",
				},
				{
					Kind:      "User",
					APIGroup:  "rbac.authorization.k8s.io",
					Name:      "test-user-updated",
					Namespace: "",
				},
			},
		},
		EmptyObject:     &rbacv1.RoleBinding{},
		EmptyObjectList: &rbacv1.RoleBindingList{},
		IsEqual: func(sourceObject client.Object, replicaObject client.Object) bool {
			sourceRoleBinding := sourceObject.(*rbacv1.RoleBinding)
			replicaRoleBinding := replicaObject.(*rbacv1.RoleBinding)
			return reflect.DeepEqual(sourceRoleBinding.RoleRef, replicaRoleBinding.RoleRef) &&
				reflect.DeepEqual(sourceRoleBinding.Subjects, replicaRoleBinding.Subjects)
		},
	})
}
