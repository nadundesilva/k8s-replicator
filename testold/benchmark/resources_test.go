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
	"fmt"
	"testing"
	"time"

	"github.com/nadundesilva/k8s-replicator/testold/utils/cleanup"
	"github.com/nadundesilva/k8s-replicator/testold/utils/controller"
	"github.com/nadundesilva/k8s-replicator/testold/utils/namespaces"
	"github.com/nadundesilva/k8s-replicator/testold/utils/resources"
	"github.com/nadundesilva/k8s-replicator/testold/utils/testdata"
	"github.com/nadundesilva/k8s-replicator/testold/utils/validation"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/features"
)

func TestResourceCreation(t *testing.T) {
	testFeatures := []features.Feature{}

	resource := testdata.GenerateSecretTestDatum()
	namespaceCounts := []int{1, 10, 100, 1000}

	measureReplication := func(ctx context.Context, t *testing.T, cfg *envconf.Config, namespaceCount int) context.Context {
		startTime := time.Now()
		resources.CreateObject(ctx, t, cfg, namespaces.GetSource(ctx).GetName(), resource.SourceObject)
		validation.ValidateReplication(ctx, t, cfg, resource.SourceObject, resource.ObjectList,
			validation.WithReplicationTimeout(time.Minute*10))

		duration := time.Since(startTime)
		report.resource = append(report.resource, ReportItem{
			InitialNamespaceCount: namespaceCount,
			NewNamespaceCount:     0,
			Duration:              fmt.Sprint(duration),
		})
		return ctx
	}

	for _, namespaceCount := range namespaceCounts {
		initialNamespaceCount := namespaceCount
		featureName := fmt.Sprintf("resource creation with %d namespaces", initialNamespaceCount)
		testFeatures = append(testFeatures, features.New(featureName).
			Setup(func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
				ctx = controller.SetupReplicator(ctx, t, cfg)
				ctx = namespaces.CreateSource(ctx, t, cfg)
				for i := 0; i < initialNamespaceCount; i++ {
					_, ctx = namespaces.CreateRandom(ctx, t, cfg)
				}
				return ctx
			}).
			Teardown(cleanup.CleanTestObjects).
			Assess("replication time", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
				// Testing creating initial namespaces set
				time.Sleep(time.Second * 30)
				ctx = measureReplication(ctx, t, cfg, initialNamespaceCount)
				return ctx
			}).
			Feature())
	}

	testenv.Test(t, testFeatures...)
}
