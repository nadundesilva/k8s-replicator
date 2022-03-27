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
	"testing"
	"time"

	"github.com/nadundesilva/k8s-replicator/pkg/replicator"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/e2e-framework/klient/k8s"
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
	controllerDockerImage = "ghcr.io/nadundesilva/k8s-replicator:test"
)

func setupReplicatorController(ctx context.Context, t *testing.T, cfg *envconf.Config) {
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
		if deploymentObj, ok := obj.(*appsv1.Deployment); ok {
			deploymentObj.Spec.Template.Spec.Containers[0].Image = controllerDockerImage
			controllerDeployment = deploymentObj
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
}

func createSourceNamespace(ctx context.Context, t *testing.T, cfg *envconf.Config, name string) {
	sourceNamespace := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	}
	err := cfg.Client().Resources().Create(ctx, sourceNamespace)
	if err != nil {
		t.Fatalf("failed to create source namespace %s: %v", sourceNamespace.GetName(), err)
	}
}

func createSourceObject(ctx context.Context, t *testing.T, cfg *envconf.Config, namespace string, obj k8s.Object) {
	obj.SetNamespace(namespace)
	labels := obj.GetLabels()
	labels[replicator.ReplicationObjectTypeLabelKey] = replicator.ReplicationObjectTypeLabelValueSource

	err := cfg.Client().Resources(namespace).Create(ctx, obj)
	if err != nil {
		t.Fatalf("failed to create source object: %v", err)
	}
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
		obj := sourceObject.DeepCopyObject()
		metaObj := obj.(metav1.Object)
		metaObj.SetNamespace(ns.GetName())
		listItems = append(listItems, obj)
	}
	meta.SetList(objectList, listItems)

	err = wait.For(conditions.New(cfg.Client().Resources()).ResourcesMatch(
		objectList,
		func(object k8s.Object) bool {
			matchMap := func(sourceMap map[string]string, targetMap map[string]string) error {
				for k, v := range sourceMap {
					if k == replicator.ReplicationObjectTypeLabelKey {
						continue
					}
					if value, ok := targetMap[k]; ok {
						if value != v {
							return fmt.Errorf("source object value %s for key %s does not exist in cloned object in namespace %s",
								v, k, object.GetNamespace())
						}
					} else {
						return fmt.Errorf("source object key %s does not exist in cloned object in namespace %s",
							k, object.GetNamespace())
					}
				}
				return nil
			}
			err := matchMap(sourceObject.GetLabels(), object.GetLabels())
			if err != nil {
				t.Errorf("labels are not matching %v", err)
			}
			err = matchMap(sourceObject.GetAnnotations(), object.GetAnnotations())
			if err != nil {
				t.Errorf("annotations are not matching %v", err)
			}

			objType, objTypeOk := object.GetLabels()[replicator.ReplicationObjectTypeLabelKey]
			if !objTypeOk {
				t.Errorf("object does not contain label key %s", replicator.ReplicationObjectTypeLabelKey)
			}
			if sourceObject.GetNamespace() == object.GetNamespace() {
				if objTypeOk && objType != replicator.ReplicationObjectTypeLabelValueSource {
					t.Errorf("object label %s does not contain the expected value; want %s, got %s",
						replicator.ReplicationObjectTypeLabelKey, replicator.ReplicationObjectTypeLabelValueSource,
						objType)
				}
			} else {
				if objTypeOk && objType != replicator.ReplicationObjectTypeLabelValueClone {
					t.Errorf("object label %s does not contain the expected value; want %s, got %s",
						replicator.ReplicationObjectTypeLabelKey, replicator.ReplicationObjectTypeLabelValueClone,
						objType)
				}

				sourceNs, sourceNsOk := object.GetLabels()[replicator.ReplicationSourceNamespaceLabelKey]
				if sourceNsOk {
					if sourceNs != sourceObject.GetNamespace() {
						t.Errorf("object label %s does not contain the clone source namespace; want %s, got %s",
							replicator.ReplicationSourceNamespaceLabelKey, sourceObject.GetNamespace(),
							sourceNs)
					}
				} else {
					t.Errorf("object does not contain label key %s", replicator.ReplicationSourceNamespaceLabelKey)
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
