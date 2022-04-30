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
package benchmark

import (
	"context"
	"testing"
	"time"

	"github.com/nadundesilva/k8s-replicator/test/utils/cleanup"
	"github.com/nadundesilva/k8s-replicator/test/utils/controller"
	"github.com/nadundesilva/k8s-replicator/test/utils/namespaces"
	"github.com/nadundesilva/k8s-replicator/test/utils/resources"
	"github.com/nadundesilva/k8s-replicator/test/utils/testdata"
	"github.com/nadundesilva/k8s-replicator/test/utils/validation"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/features"
)

func TestNamespaceCreation(t *testing.T) {
	testFeatures := []features.Feature{}

	resource := testdata.GenerateSecretTestDatum()
	initialNamespaceCounts := []int{1, 10, 100, 1000}
	testNamespaceCounts := []int{1, 10, 100, 1000}

	for _, initialNamespaceCount := range initialNamespaceCounts {
		for _, testNamespaceCount := range testNamespaceCounts {
			testFeatures = append(testFeatures, features.New("namespace count increases from 1000 to 2000").
				Setup(func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
					ctx = controller.SetupReplicator(ctx, t, cfg)
					ctx = namespaces.CreateSource(ctx, t, cfg)
					resources.CreateObject(ctx, t, cfg, namespaces.GetSource(ctx).GetName(), resource.SourceObject)

					for i := 0; i < initialNamespaceCount; i++ {
						_, ctx = namespaces.CreateRandom(ctx, t, cfg)
					}
					validation.ValidateReplication(ctx, t, cfg, resource.SourceObject, resource.ObjectList,
						validation.WithReplicationTimeout(time.Minute*10))
					return ctx
				}).
				Teardown(cleanup.CleanTestObjects).
				Assess("ignored object", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
					startTime := time.Now()

					for i := 0; i < testNamespaceCount; i++ {
						_, ctx = namespaces.CreateRandom(ctx, t, cfg)
					}
					validation.ValidateReplication(ctx, t, cfg, resource.SourceObject, resource.ObjectList,
						validation.WithReplicationTimeout(time.Minute*10))

					duration := time.Since(startTime)
					report = append(report, reportItem{
						Target:       Namespace,
						InitialCount: initialNamespaceCount,
						TestCount:    testNamespaceCount,
						Duration:     duration,
					})
					return ctx
				}).
				Feature())
		}
	}

	testenv.Test(t, testFeatures...)
}
