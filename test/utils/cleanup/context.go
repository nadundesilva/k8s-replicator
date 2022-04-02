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
package cleanup

import (
	"context"
	"testing"

	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/e2e-framework/klient/k8s"
)

const (
	testObjectsContextKey           = "__test_objects__"
	testControllerObjectsContextKey = "__test_controller_objects__"
)

type testObjects struct {
	namespaces          corev1.NamespaceList
	clusterRoles        rbacv1.ClusterRoleList
	clusterRoleBindings rbacv1.ClusterRoleBindingList
}

func AddTestObjectToContext(ctx context.Context, t *testing.T, object k8s.Object) context.Context {
	return addObjectToContext(ctx, t, object, testObjectsContextKey, "test")
}

func AddControllerObjectToContext(ctx context.Context, t *testing.T, object k8s.Object) context.Context {
	return addObjectToContext(ctx, t, object, testControllerObjectsContextKey, "controller")
}

func addObjectToContext(ctx context.Context, t *testing.T, object k8s.Object, contextKey, usage string) context.Context {
	ctxValue := ctx.Value(contextKey)
	var objects *testObjects
	if ctxValue == nil {
		objects = &testObjects{}
	} else {
		objects = ctxValue.(*testObjects)
	}

	var objectType string
	if namespace, ok := object.(*corev1.Namespace); ok {
		objects.namespaces.Items = append(objects.namespaces.Items, *namespace)
		objectType = "namespace"
	} else if clusterrole, ok := object.(*rbacv1.ClusterRole); ok {
		objects.clusterRoles.Items = append(objects.clusterRoles.Items, *clusterrole)
		objectType = "cluster role"
	} else if clusterrolebinding, ok := object.(*rbacv1.ClusterRoleBinding); ok {
		objects.clusterRoleBindings.Items = append(objects.clusterRoleBindings.Items, *clusterrolebinding)
		objectType = "cluster role binding"
	} else {
		t.Fatalf("cannot add unknown object type %s as %s object", object.GetObjectKind().GroupVersionKind().String(), usage)
	}
	t.Logf("added %s %s object %s to context", usage, objectType, object.GetName())
	return context.WithValue(ctx, contextKey, objects)
}

func RemoveTestObjectFromContext(ctx context.Context, t *testing.T, object k8s.Object) context.Context {
	return removeObjectFromContext(ctx, t, object, testObjectsContextKey, "test")
}

func RemoveControllerObjectFromContext(ctx context.Context, t *testing.T, object k8s.Object) context.Context {
	return removeObjectFromContext(ctx, t, object, testControllerObjectsContextKey, "controller")
}

func removeObjectFromContext(ctx context.Context, t *testing.T, object k8s.Object, contextKey, usage string) context.Context {
	ctxValue := ctx.Value(contextKey)
	var objects *testObjects
	if ctxValue == nil {
		objects = &testObjects{}
		t.Log("initialized new test objects struct in context")
	} else {
		objects = ctxValue.(*testObjects)
	}

	removeItemFromContext := func(testObjList []runtime.Object, obj runtime.Object) []runtime.Object {
		var index *int
		for i, testObject := range testObjList {
			if testObject.(metav1.Object).GetName() == obj.(metav1.Object).GetName() {
				index = &i
				break
			}
		}
		if index == nil {
			t.Fatalf("test object to be removed from context not present")
		} else {
			testObjList = append(testObjList[:*index], testObjList[*index+1:]...)
		}
		return testObjList
	}

	var objectType string
	if namespace, ok := object.(*corev1.Namespace); ok {
		objList, err := meta.ExtractList(&objects.namespaces)
		if err != nil {
			t.Fatalf("failed to extract namespaces list: %v", err)
		}
		objList = removeItemFromContext(objList, namespace)
		err = meta.SetList(&objects.namespaces, objList)
		if err != nil {
			t.Fatalf("failed to set the new reduced list: %v", err)
		}
		objectType = "namespace"
	} else if clusterrole, ok := object.(*rbacv1.ClusterRole); ok {
		objList, err := meta.ExtractList(&objects.clusterRoles)
		if err != nil {
			t.Fatalf("failed to extract cluster roles list: %v", err)
		}
		objList = removeItemFromContext(objList, clusterrole)
		err = meta.SetList(&objects.clusterRoles, objList)
		if err != nil {
			t.Fatalf("failed to set the new reduced list: %v", err)
		}
		objectType = "cluster role"
	} else if clusterrolebinding, ok := object.(*rbacv1.ClusterRoleBinding); ok {
		objList, err := meta.ExtractList(&objects.clusterRoleBindings)
		if err != nil {
			t.Fatalf("failed to extract cluster role bindings list: %v", err)
		}
		objList = removeItemFromContext(objList, clusterrolebinding)
		err = meta.SetList(&objects.clusterRoleBindings, objList)
		if err != nil {
			t.Fatalf("failed to set the new reduced list: %v", err)
		}
		objectType = "cluster role binding"
	} else {
		t.Fatalf("cannot remove unknown object type %s as %s object", object.GetObjectKind().GroupVersionKind().String(), usage)
	}
	t.Logf("removed %s %s object %s from context", usage, objectType, object.GetName())
	return context.WithValue(ctx, contextKey, objects)
}
