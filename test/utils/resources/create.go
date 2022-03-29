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

	"github.com/nadundesilva/k8s-replicator/pkg/replicator"
	"sigs.k8s.io/e2e-framework/klient/k8s"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
)

func CreateSourceObject(ctx context.Context, t *testing.T, cfg *envconf.Config, namespace string, obj k8s.Object) {
	obj.SetNamespace(namespace)
	clonedObj := obj.DeepCopyObject().(k8s.Object)
	labels := clonedObj.GetLabels()
	labels[replicator.ReplicationObjectTypeLabelKey] = replicator.ReplicationObjectTypeLabelValueSource

	err := cfg.Client().Resources(namespace).Create(ctx, clonedObj)
	if err != nil {
		t.Fatalf("failed to create source object: %v", err)
	}
}
