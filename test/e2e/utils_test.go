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
	"path/filepath"
	"testing"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

		kind := groupVersionKind.Version + ":" + groupVersionKind.Kind
		if len(groupVersionKind.Group) > 0 {
			kind = groupVersionKind.Group + "/" + kind
		}

		var newObj k8s.Object
		switch kind {
		case "v1:Namespace":
			newObj = obj.(*corev1.Namespace)
		case "apps/v1:Deployment":
			deploymentObj := obj.(*appsv1.Deployment)
			deploymentObj.Spec.Template.Spec.Containers[0].Image = controllerDockerImage
			controllerDeployment = deploymentObj
			newObj = deploymentObj
		case "v1:ServiceAccount":
			newObj = obj.(*corev1.ServiceAccount)
		case "rbac.authorization.k8s.io/v1:ClusterRole":
			newObj = obj.(*rbacv1.ClusterRole)
		case "rbac.authorization.k8s.io/v1:ClusterRoleBinding":
			newObj = obj.(*rbacv1.ClusterRoleBinding)
		case "v1:ConfigMap":
			newObj = obj.(*corev1.ConfigMap)
		default:
			t.Fatalf("unknown kind: %s", kind)
		}
		err = cfg.Client().Resources().Create(ctx, newObj)
		if err != nil {
			t.Fatalf("failed to create controller resource of kind %s: %v", kind, err)
		}
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
		t.Errorf("failed to create source namespace %s: %v", sourceNamespace.GetName(), err)
	}
}

func createSourceObject(ctx context.Context, t *testing.T, cfg *envconf.Config, namespace string, obj k8s.Object) {
	obj.SetNamespace(namespace)
	err := cfg.Client().Resources(namespace).Create(ctx, obj)
	if err != nil {
		t.Errorf("failed to create source object: %v", err)
	}
}

func validateReplication(ctx context.Context, t *testing.T, cfg *envconf.Config, objList k8s.ObjectList) {
	nsList := &corev1.NamespaceList{}
	err := cfg.Client().Resources().List(ctx, nsList)
	if err != nil {
		t.Errorf("failed to list namespaces: %v", err)
	}

	err = cfg.Client().Resources().List(ctx, objList)
	if err != nil {
		t.Errorf("failed to list resource: %v", err)
	}
}
