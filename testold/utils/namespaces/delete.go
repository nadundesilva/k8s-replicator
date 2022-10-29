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
package namespaces

import (
	"context"
	"testing"

	"github.com/nadundesilva/k8s-replicator/testold/utils/cleanup"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/e2e-framework/klient/k8s"
	"sigs.k8s.io/e2e-framework/klient/wait"
	"sigs.k8s.io/e2e-framework/klient/wait/conditions"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
)

func Delete(ctx context.Context, t *testing.T, cfg *envconf.Config, namespace *corev1.Namespace) context.Context {
	clonedNamespace := namespace.DeepCopyObject().(k8s.Object)

	err := cfg.Client().Resources().Delete(ctx, clonedNamespace)
	if err != nil {
		t.Fatalf("failed to delete namespace: %v", err)
	}
	t.Logf("deleted namespace %s", clonedNamespace.GetName())
	return cleanup.RemoveTestObjectFromContext(ctx, t, clonedNamespace)
}

func DeleteWithWait(ctx context.Context, t *testing.T, cfg *envconf.Config, namespace *corev1.Namespace) context.Context {
	clonedNamespace := namespace.DeepCopyObject().(k8s.Object)
	ctx = Delete(ctx, t, cfg, namespace)

	t.Logf("waiting for namespace %s to delete", clonedNamespace.GetName())
	err := wait.For(conditions.New(cfg.Client().Resources()).ResourceDeleted(clonedNamespace))
	if err != nil {
		t.Fatalf("failed to wait for namespace to delete: %v", err)
	}
	t.Logf("waiting for namespace %s to delete complete", clonedNamespace.GetName())
	return ctx
}
