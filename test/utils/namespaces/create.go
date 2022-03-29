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

	"github.com/nadundesilva/k8s-replicator/test/utils/cleanup"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
)

const (
	namespacePrefix = "replicator-e2e-ns"
)

func CreateRandom(ctx context.Context, t *testing.T, cfg *envconf.Config, options ...CreateOption) (*corev1.Namespace, context.Context) {
	opts := &CreateOptions{
		namespacePrefix: namespacePrefix,
		labels:          map[string]string{},
	}
	for _, option := range options {
		option(opts)
	}

	namespace := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name:   envconf.RandomName(opts.namespacePrefix, 32),
			Labels: opts.labels,
		},
	}
	err := cfg.Client().Resources().Create(ctx, namespace)
	if err != nil {
		t.Fatalf("failed to create namespace %s: %v", namespace.GetName(), err)
	}
	return namespace, cleanup.AddTestObjectToContext(ctx, t, namespace)
}
