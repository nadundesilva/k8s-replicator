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

	"sigs.k8s.io/e2e-framework/klient/k8s"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
)

func UpdateObject(ctx context.Context, t *testing.T, cfg *envconf.Config, namespace string, obj k8s.Object) {
	clonedObj := obj.DeepCopyObject().(k8s.Object)
	clonedObj.SetNamespace(namespace)

	err := cfg.Client().Resources(namespace).Update(ctx, clonedObj)
	if err != nil {
		t.Fatalf("failed to update object: %v", err)
	}
	t.Logf("updated object %s/%s with labels %s", namespace, clonedObj.GetName(), clonedObj.GetLabels())
}

func GetObject(ctx context.Context, t *testing.T, cfg *envconf.Config, namespace, name string, obj k8s.Object) {
	err := cfg.Client().Resources(namespace).Get(ctx, name, namespace, obj)
	if err != nil {
		t.Fatalf("failed to update object: %v", err)
	}
}
