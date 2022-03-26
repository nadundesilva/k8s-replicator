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
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/e2e-framework/klient/k8s"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
)

func setupReplicatorController(ctx context.Context, t *testing.T, cfg *envconf.Config) {
	// Start controller on the cluster
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
