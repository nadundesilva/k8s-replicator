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
package e2e

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/nadundesilva/k8s-replicator/pkg/replicator"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/e2e-framework/klient/k8s"
	"sigs.k8s.io/e2e-framework/klient/k8s/resources"
	"sigs.k8s.io/e2e-framework/klient/wait"
	"sigs.k8s.io/e2e-framework/klient/wait/conditions"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/kustomize/api/krusty"
	"sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/kustomize/kyaml/filesys"
)

const (
	kustomizeDirName      = "kustomize"
	namespacePrefix       = "replicator-e2e-"
	testObjectsContextKey = "__test_objects__"
)

type testObjects struct {
	namespaces          corev1.NamespaceList
	clusterRoles        rbacv1.ClusterRoleList
	clusterRoleBindings rbacv1.ClusterRoleBindingList
}

func setupReplicatorController(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
	kustomizeDir, err := filepath.Abs(filepath.Join("..", "..", kustomizeDirName))
	if err != nil {
		t.Fatalf("failed to resolve kustomize dir %s: %v", kustomizeDirName, err)
	}
	t.Logf("creating controller from kustomize dir: %s", kustomizeDir)

	fileSys := filesys.MakeFsOnDisk()
	if !fileSys.Exists(kustomizeDir) {
		t.Fatalf("kustomization dir %s does not exist on file system", kustomizeDir)
	}

	k := krusty.MakeKustomizer(&krusty.Options{
		AddManagedbyLabel: true,
		PluginConfig: &types.PluginConfig{
			FnpLoadingOptions: types.FnPluginLoadingOptions{},
		},
	})
	m, err := k.Run(fileSys, kustomizeDir)
	if err != nil {
		t.Fatalf("failed build kustomization: %v", err)
	}

	var controllerDeployment *appsv1.Deployment
	for _, resource := range m.Resources() {
		yaml, err := resource.AsYAML()
		if err != nil {
			t.Fatalf("failed get kustomization output yaml: %v", err)
		}
		obj, groupVersionKind, err := scheme.Codecs.UniversalDeserializer().Decode(yaml, nil, nil)
		if err != nil {
			t.Fatalf("failed parse kustomization output yaml: %v", err)
		}

		kind := groupVersionKind.String()
		if deployment, ok := obj.(*appsv1.Deployment); ok {
			deployment.Spec.Template.Spec.Containers[0].Image = controllerImage
			controllerDeployment = deployment
		} else if namespace, ok := obj.(*corev1.Namespace); ok {
			ctx = addTestObjectToContext(ctx, t, namespace)
		} else if clusterrole, ok := obj.(*rbacv1.ClusterRole); ok {
			ctx = addTestObjectToContext(ctx, t, clusterrole)
		} else if clusterrolebinding, ok := obj.(*rbacv1.ClusterRoleBinding); ok {
			ctx = addTestObjectToContext(ctx, t, clusterrolebinding)
		}
		err = cfg.Client().Resources().Create(ctx, obj.(k8s.Object))
		if err != nil {
			t.Fatalf("failed to create controller resource of kind %s: %v", kind, err)
		}
	}
	if controllerDeployment == nil {
		t.Fatalf("controller deployment not found in controller kustomize files")
	}

	err = wait.For(conditions.New(cfg.Client().Resources()).ResourceMatch(controllerDeployment, func(object k8s.Object) bool {
		d := object.(*appsv1.Deployment)
		return d.Status.AvailableReplicas > 0 && d.Status.ReadyReplicas > 0
	}), wait.WithTimeout(time.Minute))
	if err != nil {
		t.Fatalf("failed to wait for controller deployment to be ready: %v", err)
	}
	return ctx
}

func createRandomNamespace(ctx context.Context, t *testing.T, cfg *envconf.Config) (*corev1.Namespace, context.Context) {
	namespace := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: envconf.RandomName(namespacePrefix+"ns", 32),
		},
	}
	err := cfg.Client().Resources().Create(ctx, namespace)
	if err != nil {
		t.Fatalf("failed to create namespace %s: %v", namespace.GetName(), err)
	}
	return namespace, addTestObjectToContext(ctx, t, namespace)
}

func createSourceObject(ctx context.Context, t *testing.T, cfg *envconf.Config, namespace string, obj k8s.Object) {
	obj.SetNamespace(namespace)
	clonedObj := obj.DeepCopyObject().(k8s.Object)
	labels := clonedObj.GetLabels()
	labels[replicator.ReplicationObjectTypeLabelKey] = replicator.ReplicationObjectTypeLabelValueSource

	err := cfg.Client().Resources(namespace).Create(ctx, clonedObj)
	if err != nil {
		t.Fatalf("failed to create source object: %v", err)
	}
}

func deleteObject(ctx context.Context, t *testing.T, cfg *envconf.Config, namespace string, obj k8s.Object) {
	clonedObj := obj.DeepCopyObject().(k8s.Object)
	clonedObj.SetNamespace(namespace)

	err := cfg.Client().Resources(namespace).Delete(ctx, clonedObj)
	if err != nil {
		t.Fatalf("failed to delete object: %v", err)
	}
}

func deleteObjectWithWait(ctx context.Context, t *testing.T, cfg *envconf.Config, namespace string, obj k8s.Object) {
	clonedObj := obj.DeepCopyObject().(k8s.Object)
	clonedObj.SetNamespace(namespace)
	deleteObject(ctx, t, cfg, namespace, clonedObj)

	err := wait.For(conditions.New(cfg.Client().Resources(namespace)).ResourceDeleted(clonedObj))
	if err != nil {
		t.Fatalf("failed to wait for object to delete: %v", err)
	}
}

func deleteNamespace(ctx context.Context, t *testing.T, cfg *envconf.Config, namespace *corev1.Namespace) context.Context {
	clonedNamespace := namespace.DeepCopyObject().(k8s.Object)

	err := cfg.Client().Resources().Delete(ctx, clonedNamespace)
	if err != nil {
		t.Fatalf("failed to delete namespace: %v", err)
	}
	return removeTestObjectFromContext(ctx, t, clonedNamespace)
}

func deleteNamespaceWithWait(ctx context.Context, t *testing.T, cfg *envconf.Config, namespace *corev1.Namespace) context.Context {
	clonedNamespace := namespace.DeepCopyObject().(k8s.Object)
	ctx = deleteNamespace(ctx, t, cfg, namespace)

	err := wait.For(conditions.New(cfg.Client().Resources()).ResourceDeleted(clonedNamespace))
	if err != nil {
		t.Fatalf("failed to wait for namespace to delete: %v", err)
	}
	return ctx
}

type objectMatcher func(sourceObject k8s.Object, targetObject k8s.Object) bool

func validateReplication(ctx context.Context, t *testing.T, cfg *envconf.Config,
	sourceObject k8s.Object, objectList k8s.ObjectList, objMatcher objectMatcher) {
	nsList := &corev1.NamespaceList{}
	err := cfg.Client().Resources().List(ctx, nsList)
	if err != nil {
		t.Fatalf("failed to list namespaces: %v", err)
	}

	listItems := []runtime.Object{}
	for _, ns := range nsList.Items {
		val, ok := ns.GetLabels()[replicator.ReplicationTargetNamespaceTypeLabelKey]
		if (ok && val == replicator.ReplicationTargetNamespaceTypeLabelValueIgnored) ||
			((!ok || val != replicator.ReplicationTargetNamespaceTypeLabelValueReplicated) &&
			(strings.HasPrefix(ns.GetName(), "kube-") || ns.GetName() == "k8s-replicator")) {
			continue
		}

		clonedObj := sourceObject.DeepCopyObject().(k8s.Object)
		clonedObj.SetNamespace(ns.GetName())
		listItems = append(listItems, clonedObj)
	}
	err = meta.SetList(objectList, listItems)
	if err != nil {
		t.Fatalf("failed to create list of objects to wait for: %v", err)
	}

	err = wait.For(conditions.New(cfg.Client().Resources()).ResourcesMatch(
		objectList.DeepCopyObject().(k8s.ObjectList),
		func(object k8s.Object) bool {
			matchMap := func(sourceMap map[string]string, targetMap map[string]string) error {
				for k, v := range sourceMap {
					if k == replicator.ReplicationObjectTypeLabelKey {
						continue
					}
					if value, ok := targetMap[k]; ok {
						if value != v {
							return fmt.Errorf("source object %s/%s value %s for key %s does not exist in cloned object",
								sourceObject.GetNamespace(), sourceObject.GetName(), v, k)
						}
					} else {
						return fmt.Errorf("source object %s/%s key %s does not exist in cloned object",
							sourceObject.GetNamespace(), sourceObject.GetName(), k)
					}
				}
				return nil
			}
			err := matchMap(sourceObject.GetLabels(), object.GetLabels())
			if err != nil {
				t.Errorf("object %s/%s labels are not matching %v", object.GetNamespace(), object.GetName(), err)
			}
			err = matchMap(sourceObject.GetAnnotations(), object.GetAnnotations())
			if err != nil {
				t.Errorf("object %s/%s annotations are not matching %v", object.GetNamespace(), object.GetName(), err)
			}

			objType, objTypeOk := object.GetLabels()[replicator.ReplicationObjectTypeLabelKey]
			if !objTypeOk {
				t.Errorf("object %s/%s does not contain label key %s", object.GetNamespace(), object.GetName(),
					replicator.ReplicationObjectTypeLabelKey)
			}
			if sourceObject.GetNamespace() == object.GetNamespace() {
				if objTypeOk && objType != replicator.ReplicationObjectTypeLabelValueSource {
					t.Errorf("object %s/%s label %s does not contain the expected value; want %s, got %s",
						object.GetNamespace(), object.GetName(), replicator.ReplicationObjectTypeLabelKey,
						replicator.ReplicationObjectTypeLabelValueSource, objType)
				}
			} else {
				if objTypeOk && objType != replicator.ReplicationObjectTypeLabelValueClone {
					t.Errorf("object %s/%s label %s does not contain the expected value; want %s, got %s",
						object.GetNamespace(), object.GetName(), replicator.ReplicationObjectTypeLabelKey,
						replicator.ReplicationObjectTypeLabelValueClone, objType)
				}

				sourceNs, sourceNsOk := object.GetLabels()[replicator.ReplicationSourceNamespaceLabelKey]
				if sourceNsOk {
					if sourceNs != sourceObject.GetNamespace() {
						t.Errorf("object %s/%s label %s does not contain the source namespace; want %s, got %s",
							object.GetNamespace(), object.GetName(), replicator.ReplicationSourceNamespaceLabelKey,
							sourceObject.GetNamespace(), sourceNs)
					}
				} else {
					t.Errorf("object %s/%s does not contain label key %s", object.GetNamespace(), object.GetName(),
						replicator.ReplicationSourceNamespaceLabelKey)
				}
			}
			objMatcher(sourceObject, object)
			return true
		}),
		wait.WithTimeout(time.Minute),
	)
	if err != nil {
		t.Fatalf("failed to wait for replicated objects: %v", err)
	}
}

func validateResourceDeletion(ctx context.Context, t *testing.T, cfg *envconf.Config, sourceObject k8s.Object) {
	nsList := &corev1.NamespaceList{}
	err := cfg.Client().Resources().List(ctx, nsList)
	if err != nil {
		t.Fatalf("failed to list namespaces: %v", err)
	}
	for _, namespace := range nsList.Items {
		clonedObj := sourceObject.DeepCopyObject().(k8s.Object)
		clonedObj.SetNamespace(namespace.GetName())
		err := wait.For(conditions.New(cfg.Client().Resources(namespace.GetName())).ResourceDeleted(clonedObj))
		if err != nil {
			t.Fatalf("failed to wait for replicated objects: %v", err)
		}
	}
}

func addTestObjectToContext(ctx context.Context, t *testing.T, object k8s.Object) context.Context {
	ctxValue := ctx.Value(testObjectsContextKey)
	var objects *testObjects
	if ctxValue == nil {
		objects = &testObjects{}
	} else {
		objects = ctxValue.(*testObjects)
	}

	if namespace, ok := object.(*corev1.Namespace); ok {
		objects.namespaces.Items = append(objects.namespaces.Items, *namespace)
	} else if clusterrole, ok := object.(*rbacv1.ClusterRole); ok {
		objects.clusterRoles.Items = append(objects.clusterRoles.Items, *clusterrole)
	} else if clusterrolebinding, ok := object.(*rbacv1.ClusterRoleBinding); ok {
		objects.clusterRoleBindings.Items = append(objects.clusterRoleBindings.Items, *clusterrolebinding)
	} else {
		t.Fatalf("cannot add unknown object type %s as test object", object.GetObjectKind().GroupVersionKind().String())
	}
	return context.WithValue(ctx, testObjectsContextKey, objects)
}

func removeTestObjectFromContext(ctx context.Context, t *testing.T, object k8s.Object) context.Context {
	ctxValue := ctx.Value(testObjectsContextKey)
	var objects *testObjects
	if ctxValue == nil {
		objects = &testObjects{}
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
	} else {
		t.Fatalf("cannot remove unknown object type %s as test object", object.GetObjectKind().GroupVersionKind().String())
	}
	return context.WithValue(ctx, testObjectsContextKey, objects)
}

func cleanupTestObjects(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
	ctxValue := ctx.Value(testObjectsContextKey)
	if ctxValue != nil {
		deleteObjs := func(object k8s.Object) {
			err := cfg.Client().Resources().Delete(ctx, object.DeepCopyObject().(k8s.Object),
				resources.WithDeletePropagation("Background"))
			if err != nil {
				t.Errorf("failed to delete test object %s: %v", object.GetName(), err)
			}
		}
		waitForDeleteObjs := func(objList k8s.ObjectList) {
			clonedObjList := objList.DeepCopyObject().(k8s.ObjectList)
			err := wait.For(conditions.New(cfg.Client().Resources()).ResourcesDeleted(clonedObjList))
			if err != nil {
				t.Fatalf("failed to wait for objects to delete: %v", err)
			}
		}

		objects := ctxValue.(*testObjects)
		for _, obj := range objects.namespaces.Items {
			deleteObjs(&obj)
		}
		for _, obj := range objects.clusterRoles.Items {
			deleteObjs(&obj)
		}
		for _, obj := range objects.clusterRoleBindings.Items {
			deleteObjs(&obj)
		}
		waitForDeleteObjs(&objects.namespaces)
		waitForDeleteObjs(&objects.clusterRoles)
		waitForDeleteObjs(&objects.clusterRoleBindings)
		ctx = context.WithValue(ctx, testObjectsContextKey, nil)
	}
	return ctx
}
