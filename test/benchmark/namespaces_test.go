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

	"github.com/nadundesilva/k8s-replicator/test/utils/cleanup"
	"github.com/nadundesilva/k8s-replicator/test/utils/common"
	"github.com/nadundesilva/k8s-replicator/test/utils/controller"
	"github.com/nadundesilva/k8s-replicator/test/utils/namespaces"
	"github.com/nadundesilva/k8s-replicator/test/utils/resources"
	"github.com/nadundesilva/k8s-replicator/test/utils/validation"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/features"
)

func TestNamespaceCreation(t *testing.T) {
	testFeatures := []features.Feature{}

	resource := getBenchmarkTestData(t)
	initialNamespaceCounts := []int{1, 10, 100, 1000}
	newNamespaceCount := 100

	measureReplication := func(ctx context.Context, t *testing.T, cfg *envconf.Config, initialNamespaceCount, newNamespaceCount int) context.Context {
		startTime := time.Now()
		for i := 0; i < newNamespaceCount; i++ {
			_, ctx = namespaces.CreateRandom(ctx, t, cfg)
		}
		validation.ValidateReplication(ctx, t, cfg, resource.SourceObject(), resource.EmptyObjectList(),
			validation.WithReplicationTimeout(time.Minute*10))

		duration := time.Since(startTime)
		report.namespace = append(report.namespace, ReportItem{
			InitialNamespaceCount: initialNamespaceCount,
			NewNamespaceCount:     newNamespaceCount,
			Duration:              fmt.Sprint(duration),
		})
		return ctx
	}

	for _, initialNamespaceCount := range initialNamespaceCounts {
		startingNamespaceCount := initialNamespaceCount
		featureName := fmt.Sprintf("namespace count increases by %d from %d", newNamespaceCount, startingNamespaceCount)
		testFeatures = append(testFeatures, features.New(featureName).
			Setup(func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
				ctx = controller.SetupReplicator(ctx, t, cfg, controller.WithVerbosityLevel(controllerLogVerbosity))
				ctx = namespaces.CreateSource(ctx, t, cfg)
				resources.CreateObject(ctx, t, cfg, common.GetSourceObjectNamespace(ctx).GetName(), resource.SourceObject())
				return ctx
			}).
			Teardown(func(ctx context.Context, t *testing.T, c *envconf.Config) context.Context {
				return cleanup.CleanTestObjectsWithOptions(ctx, t, c, cleanup.WithTimeout(time.Minute*5))
			}).
			Assess("replication time", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
				// Testing creating initial namespaces set
				time.Sleep(time.Second * 30)
				ctx = measureReplication(ctx, t, cfg, 0, startingNamespaceCount)

				// Testing creating namespaces with an initial set of namespaces already in cluster
				time.Sleep(time.Second * 30)
				ctx = measureReplication(ctx, t, cfg, startingNamespaceCount, newNamespaceCount)
				return ctx
			}).
			Feature())
	}

	testenv.Test(t, testFeatures...)
}
