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

	"github.com/nadundesilva/k8s-replicator/test/utils/cleanup"
	"github.com/nadundesilva/k8s-replicator/test/utils/common"
	"github.com/nadundesilva/k8s-replicator/test/utils/controller"
	"github.com/nadundesilva/k8s-replicator/test/utils/namespaces"
	"github.com/nadundesilva/k8s-replicator/test/utils/resources"
	"github.com/nadundesilva/k8s-replicator/test/utils/testdata"
	"github.com/nadundesilva/k8s-replicator/test/utils/validation"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/features"
)

func TestResourcesCreation(t *testing.T) {
	testResources := testdata.GenerateResourceTestData()

	testFeatures := []features.Feature{}
	for _, resource := range testResources {
		setupInitialNamspaces := func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			for i := 1; i < 10; i++ {
				_, ctx = namespaces.CreateRandom(ctx, t, cfg)
			}
			return ctx
		}
		assessResourcesReplication := func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			validation.ValidateReplication(ctx, t, cfg, resource.SourceObject(), resource.EmptyObjectList(),
				validation.WithObjectMatcher(resource.IsEqual))
			return ctx
		}

		testFeatures = append(testFeatures, newFeatureBuilder("controller starts before initial namespaces creation", resource).
			WithLabel("operation", "create").
			Setup(func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
				ctx = setupInitialNamspaces(ctx, t, cfg)
				ctx = controller.SetupReplicator(ctx, t, cfg)
				ctx = namespaces.CreateSource(ctx, t, cfg)
				resources.CreateObject(ctx, t, cfg, common.GetSourceObjectNamespace(ctx).GetName(), resource.SourceObject())
				_, ctx = namespaces.CreateRandom(ctx, t, cfg)
				_, ctx = namespaces.CreateRandom(ctx, t, cfg)
				return ctx
			}).
			Teardown(cleanup.CleanTestObjects).
			Assess("replicated objects", assessResourcesReplication).
			Feature())

		testFeatures = append(testFeatures, newFeatureBuilder("controller starts between initial namespaces creation and source namespace creation", resource).
			WithLabel("operation", "create").
			Setup(func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
				ctx = setupInitialNamspaces(ctx, t, cfg)
				ctx = controller.SetupReplicator(ctx, t, cfg)
				ctx = namespaces.CreateSource(ctx, t, cfg)
				resources.CreateObject(ctx, t, cfg, common.GetSourceObjectNamespace(ctx).GetName(), resource.SourceObject())
				_, ctx = namespaces.CreateRandom(ctx, t, cfg)
				_, ctx = namespaces.CreateRandom(ctx, t, cfg)
				return ctx
			}).
			Teardown(cleanup.CleanTestObjects).
			Assess("replicated objects", assessResourcesReplication).
			Feature())

		testFeatures = append(testFeatures, newFeatureBuilder("controller starts between source namespace creation and source object creation", resource).
			WithLabel("operation", "create").
			Setup(func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
				ctx = setupInitialNamspaces(ctx, t, cfg)
				ctx = namespaces.CreateSource(ctx, t, cfg)
				ctx = controller.SetupReplicator(ctx, t, cfg)
				resources.CreateObject(ctx, t, cfg, common.GetSourceObjectNamespace(ctx).GetName(), resource.SourceObject())
				_, ctx = namespaces.CreateRandom(ctx, t, cfg)
				_, ctx = namespaces.CreateRandom(ctx, t, cfg)
				return ctx
			}).
			Teardown(cleanup.CleanTestObjects).
			Assess("replicated objects", assessResourcesReplication).
			Feature())

		testFeatures = append(testFeatures, newFeatureBuilder("controller starts between source object creation and first new namespace creation", resource).
			WithLabel("operation", "create").
			Setup(func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
				ctx = setupInitialNamspaces(ctx, t, cfg)
				ctx = namespaces.CreateSource(ctx, t, cfg)
				resources.CreateObject(ctx, t, cfg, common.GetSourceObjectNamespace(ctx).GetName(), resource.SourceObject())
				ctx = controller.SetupReplicator(ctx, t, cfg)
				_, ctx = namespaces.CreateRandom(ctx, t, cfg)
				_, ctx = namespaces.CreateRandom(ctx, t, cfg)
				return ctx
			}).
			Teardown(cleanup.CleanTestObjects).
			Assess("replicated objects", assessResourcesReplication).
			Feature())

		testFeatures = append(testFeatures, newFeatureBuilder("controller starts between first new namespace creation and second new namespace creation", resource).
			WithLabel("operation", "create").
			Setup(func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
				ctx = setupInitialNamspaces(ctx, t, cfg)
				ctx = namespaces.CreateSource(ctx, t, cfg)
				resources.CreateObject(ctx, t, cfg, common.GetSourceObjectNamespace(ctx).GetName(), resource.SourceObject())
				_, ctx = namespaces.CreateRandom(ctx, t, cfg)
				ctx = controller.SetupReplicator(ctx, t, cfg)
				_, ctx = namespaces.CreateRandom(ctx, t, cfg)
				return ctx
			}).
			Teardown(cleanup.CleanTestObjects).
			Assess("replicated objects", assessResourcesReplication).
			Feature())

		testFeatures = append(testFeatures, newFeatureBuilder("controller starts after second new namespace creation", resource).
			WithLabel("operation", "create").
			Setup(func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
				ctx = setupInitialNamspaces(ctx, t, cfg)
				ctx = namespaces.CreateSource(ctx, t, cfg)
				resources.CreateObject(ctx, t, cfg, common.GetSourceObjectNamespace(ctx).GetName(), resource.SourceObject())
				_, ctx = namespaces.CreateRandom(ctx, t, cfg)
				_, ctx = namespaces.CreateRandom(ctx, t, cfg)
				ctx = controller.SetupReplicator(ctx, t, cfg)
				return ctx
			}).
			Teardown(cleanup.CleanTestObjects).
			Assess("replicated objects", assessResourcesReplication).
			Feature())
	}

	testenv.Test(t, testFeatures...)
}

func TestResourcesUpdation(t *testing.T) {
	testResources := testdata.GenerateResourceTestData()

	testFeatures := []features.Feature{}
	for _, resource := range testResources {
		setupInitialNamspaces := func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			for i := 1; i < 10; i++ {
				_, ctx = namespaces.CreateRandom(ctx, t, cfg)
			}
			return ctx
		}
		assessResourcesReplication := func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			validation.ValidateReplication(ctx, t, cfg, resource.SourceObjectUpdate(), resource.EmptyObjectList(),
				validation.WithObjectMatcher(resource.IsEqual))
			return ctx
		}

		testFeatures = append(testFeatures, newFeatureBuilder("controller updates replicas when source object is updated", resource).
			WithLabel("operation", "update").
			Setup(func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
				ctx = setupInitialNamspaces(ctx, t, cfg)
				ctx = namespaces.CreateSource(ctx, t, cfg)
				resources.CreateObject(ctx, t, cfg, common.GetSourceObjectNamespace(ctx).GetName(), resource.SourceObject())
				_, ctx = namespaces.CreateRandom(ctx, t, cfg)
				ctx = controller.SetupReplicator(ctx, t, cfg)
				_, ctx = namespaces.CreateRandom(ctx, t, cfg)
				resources.UpdateObject(ctx, t, cfg, common.GetSourceObjectNamespace(ctx).GetName(), resource.SourceObjectUpdate())
				return ctx
			}).
			Teardown(cleanup.CleanTestObjects).
			Assess("updated replicas", assessResourcesReplication).
			Feature())

		testFeatures = append(testFeatures, newFeatureBuilder("controller removes replicas when source object source type label is removed", resource).
			WithLabel("operation", "update").
			Setup(func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
				ctx = setupInitialNamspaces(ctx, t, cfg)
				ctx = namespaces.CreateSource(ctx, t, cfg)
				resources.CreateObject(ctx, t, cfg, common.GetSourceObjectNamespace(ctx).GetName(), resource.SourceObject())
				_, ctx = namespaces.CreateRandom(ctx, t, cfg)
				ctx = controller.SetupReplicator(ctx, t, cfg)
				_, ctx = namespaces.CreateRandom(ctx, t, cfg)
				validation.ValidateReplication(ctx, t, cfg, resource.SourceObject(), resource.EmptyObjectList(),
					validation.WithObjectMatcher(resource.IsEqual))

				updatedObject := resource.SourceObject()
				sourceObjectLabels := updatedObject.GetLabels()
				delete(sourceObjectLabels, common.ObjectTypeLabelKey)
				resources.UpdateObject(ctx, t, cfg, common.GetSourceObjectNamespace(ctx).GetName(), updatedObject)
				return ctx
			}).
			Teardown(cleanup.CleanTestObjects).
			Assess("deleted replicas", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
				validation.ValidateResourceDeletion(ctx, t, cfg, resource.SourceObject(),
					validation.WithDeletionIgnoredNamespaces(common.GetSourceObjectNamespace(ctx).GetName()))
				return ctx
			}).
			Feature())

		testFeatures = append(testFeatures, newFeatureBuilder("controller removes replicas when source object type is set to different value", resource).
			WithLabel("operation", "update").
			Setup(func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
				ctx = setupInitialNamspaces(ctx, t, cfg)
				ctx = namespaces.CreateSource(ctx, t, cfg)
				resources.CreateObject(ctx, t, cfg, common.GetSourceObjectNamespace(ctx).GetName(), resource.SourceObject())
				_, ctx = namespaces.CreateRandom(ctx, t, cfg)
				ctx = controller.SetupReplicator(ctx, t, cfg)
				_, ctx = namespaces.CreateRandom(ctx, t, cfg)
				validation.ValidateReplication(ctx, t, cfg, resource.SourceObject(), resource.EmptyObjectList(),
					validation.WithObjectMatcher(resource.IsEqual))

				updatedObject := resource.SourceObject()
				sourceObjectLabels := updatedObject.GetLabels()
				sourceObjectLabels[common.ObjectTypeLabelKey] = "ignored"
				resources.UpdateObject(ctx, t, cfg, common.GetSourceObjectNamespace(ctx).GetName(), updatedObject)
				return ctx
			}).
			Teardown(cleanup.CleanTestObjects).
			Assess("deleted replicas", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
				validation.ValidateResourceDeletion(ctx, t, cfg, resource.SourceObject(),
					validation.WithDeletionIgnoredNamespaces(common.GetSourceObjectNamespace(ctx).GetName()))
				return ctx
			}).
			Feature())
	}

	testenv.Test(t, testFeatures...)
}

func TestResourcesDeletion(t *testing.T) {
	testResources := testdata.GenerateResourceTestData()

	testFeatures := []features.Feature{}
	for _, resource := range testResources {
		setupInitialNamspaces := func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			for i := 1; i < 10; i++ {
				_, ctx = namespaces.CreateRandom(ctx, t, cfg)
			}
			return ctx
		}

		testFeatures = append(testFeatures, newFeatureBuilder("controller deletes all replicas when source object is deleted", resource).
			WithLabel("operation", "delete").
			Setup(func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
				ctx = setupInitialNamspaces(ctx, t, cfg)
				ctx = namespaces.CreateSource(ctx, t, cfg)
				resources.CreateObject(ctx, t, cfg, common.GetSourceObjectNamespace(ctx).GetName(), resource.SourceObject())
				ctx = controller.SetupReplicator(ctx, t, cfg)
				validation.ValidateReplication(ctx, t, cfg, resource.SourceObject(), resource.EmptyObjectList(),
					validation.WithObjectMatcher(resource.IsEqual))

				resources.DeleteObjectWithWait(ctx, t, cfg, common.GetSourceObjectNamespace(ctx).GetName(), resource.SourceObject())
				return ctx
			}).
			Teardown(cleanup.CleanTestObjects).
			Assess("deleted replicas", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
				validation.ValidateResourceDeletion(ctx, t, cfg, resource.SourceObject())
				return ctx
			}).
			Feature())

		testFeatures = append(testFeatures, newFeatureBuilder("controller deletes all replicas when namespace with source object is deleted", resource).
			WithLabel("operation", "delete").
			Setup(func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
				ctx = setupInitialNamspaces(ctx, t, cfg)
				ctx = namespaces.CreateSource(ctx, t, cfg)
				resources.CreateObject(ctx, t, cfg, common.GetSourceObjectNamespace(ctx).GetName(), resource.SourceObject())
				ctx = controller.SetupReplicator(ctx, t, cfg)
				validation.ValidateReplication(ctx, t, cfg, resource.SourceObject(), resource.EmptyObjectList(),
					validation.WithObjectMatcher(resource.IsEqual))

				ctx = namespaces.DeleteWithWait(ctx, t, cfg, common.GetSourceObjectNamespace(ctx))
				return ctx
			}).
			Teardown(cleanup.CleanTestObjects).
			Assess("deleted replicas", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
				validation.ValidateResourceDeletion(ctx, t, cfg, resource.SourceObject())
				return ctx
			}).
			Feature())

		testFeatures = append(testFeatures, newFeatureBuilder("controller recreates deleted replicas", resource).
			WithLabel("operation", "delete").
			Setup(func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
				ctx = setupInitialNamspaces(ctx, t, cfg)
				ctx = namespaces.CreateSource(ctx, t, cfg)
				resources.CreateObject(ctx, t, cfg, common.GetSourceObjectNamespace(ctx).GetName(), resource.SourceObject())
				ctx = controller.SetupReplicator(ctx, t, cfg)
				cloneNamespace, ctx := namespaces.CreateRandom(ctx, t, cfg)
				validation.ValidateReplication(ctx, t, cfg, resource.SourceObject(), resource.EmptyObjectList(),
					validation.WithObjectMatcher(resource.IsEqual))

				resources.DeleteObject(ctx, t, cfg, cloneNamespace.GetName(), resource.SourceObject())
				return ctx
			}).
			Teardown(cleanup.CleanTestObjects).
			Assess("recreated replicas", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
				validation.ValidateReplication(ctx, t, cfg, resource.SourceObject(), resource.EmptyObjectList(),
					validation.WithObjectMatcher(resource.IsEqual))
				return ctx
			}).
			Feature())

		testFeatures = append(testFeatures, newFeatureBuilder("controller allows namespace with replica to be deleted", resource).
			WithLabel("operation", "delete").
			Setup(func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
				ctx = setupInitialNamspaces(ctx, t, cfg)
				ctx = namespaces.CreateSource(ctx, t, cfg)
				resources.CreateObject(ctx, t, cfg, common.GetSourceObjectNamespace(ctx).GetName(), resource.SourceObject())
				ctx = controller.SetupReplicator(ctx, t, cfg)
				cloneNamespace, ctx := namespaces.CreateRandom(ctx, t, cfg)
				validation.ValidateReplication(ctx, t, cfg, resource.SourceObject(), resource.EmptyObjectList(),
					validation.WithObjectMatcher(resource.IsEqual))

				ctx = namespaces.DeleteWithWait(ctx, t, cfg, cloneNamespace)
				return ctx
			}).
			Teardown(cleanup.CleanTestObjects).
			Assess("remaining replicas", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
				validation.ValidateReplication(ctx, t, cfg, resource.SourceObject(), resource.EmptyObjectList(),
					validation.WithObjectMatcher(resource.IsEqual))
				return ctx
			}).
			Feature())
	}

	testenv.Test(t, testFeatures...)
}
