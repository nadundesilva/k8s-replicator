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
package resources

import (
	"context"
	"testing"
	"time"

	"sigs.k8s.io/e2e-framework/klient/k8s"
	"sigs.k8s.io/e2e-framework/klient/wait"
	"sigs.k8s.io/e2e-framework/klient/wait/conditions"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
)

func DeleteObject(ctx context.Context, t *testing.T, cfg *envconf.Config, namespace string, obj k8s.Object) {
	clonedObj := obj.DeepCopyObject().(k8s.Object)
	clonedObj.SetNamespace(namespace)

	err := cfg.Client().Resources(namespace).Delete(ctx, clonedObj)
	if err != nil {
		t.Fatalf("failed to delete object: %v", err)
	}
	t.Logf("deleted object %s/%s", namespace, clonedObj.GetName())
}

func DeleteObjectWithWait(ctx context.Context, t *testing.T, cfg *envconf.Config, namespace string, obj k8s.Object) {
	clonedObj := obj.DeepCopyObject().(k8s.Object)
	clonedObj.SetNamespace(namespace)
	DeleteObject(ctx, t, cfg, namespace, clonedObj)

	t.Logf("waiting for object %s/%s to delete", namespace, clonedObj.GetName())
	err := wait.For(
		conditions.New(cfg.Client().Resources(namespace)).ResourceDeleted(clonedObj),
		wait.WithTimeout(time.Minute),
		wait.WithImmediate(),
		wait.WithInterval(time.Second),
	)
	if err != nil {
		t.Fatalf("failed to wait for object to delete: %v", err)
	}
	t.Logf("waiting for object %s/%s to delete complete", namespace, clonedObj.GetName())
}
