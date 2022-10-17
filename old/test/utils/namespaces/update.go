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

	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/e2e-framework/klient/k8s"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
)

func Update(ctx context.Context, t *testing.T, cfg *envconf.Config, namespace *corev1.Namespace) {
	clonedNs := namespace.DeepCopyObject().(k8s.Object)
	err := cfg.Client().Resources().Update(ctx, clonedNs)
	if err != nil {
		t.Fatalf("failed to update namespace: %v", err)
	}
	t.Logf("updated namespace %s with labels %s", namespace.GetName(), namespace.GetLabels())
}
