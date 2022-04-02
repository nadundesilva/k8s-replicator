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
	"fmt"
	"testing"

	"github.com/nadundesilva/k8s-replicator/pkg/replicator"
	"github.com/nadundesilva/k8s-replicator/test/utils/cleanup"
	"github.com/nadundesilva/k8s-replicator/test/utils/controller"
	"github.com/nadundesilva/k8s-replicator/test/utils/namespaces"
	"github.com/nadundesilva/k8s-replicator/test/utils/resources"
	"github.com/nadundesilva/k8s-replicator/test/utils/validation"
	"sigs.k8s.io/e2e-framework/klient/k8s"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/features"
)

func TestResourcesCreation(t *testing.T) {
	testResources := generateResourcesCreationTestData(t)

	testFeatures := []features.Feature{}
	for _, resource := range testResources {
		setupInitialNamspaces := func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			for i := 1; i < 10; i++ {
				_, ctx = namespaces.CreateRandom(ctx, t, cfg)
			}
			return ctx
		}
		assessResourcesReplication := func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			validation.ValidateReplication(ctx, t, cfg, resource.sourceObject, resource.objectList,
				validation.WithObjectMatcher(resource.matcher))
			return ctx
		}

		testFeatures = append(testFeatures, features.New("controller starts before initial namespaces creation").
			WithLabel("resource", resource.name).
			WithLabel("operation", "create").
			Setup(func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
				ctx = setupInitialNamspaces(ctx, t, cfg)
				ctx = controller.SetupReplicator(ctx, t, cfg)
				ctx = namespaces.CreateSource(ctx, t, cfg)
				resources.CreateObject(ctx, t, cfg, namespaces.GetSource(ctx).GetName(), resource.sourceObject)
				_, ctx = namespaces.CreateRandom(ctx, t, cfg)
				_, ctx = namespaces.CreateRandom(ctx, t, cfg)
				return ctx
			}).
			Teardown(cleanup.CleanTestObjects).
			Assess("replicated objects", assessResourcesReplication).
			Feature())

		testFeatures = append(testFeatures, features.New(fmt.Sprintf("controller starts netween initial namespaces creation and source namespace %s creation", resource.name)).
			WithLabel("resource", resource.name).
			WithLabel("operation", "create").
			Setup(func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
				ctx = setupInitialNamspaces(ctx, t, cfg)
				ctx = controller.SetupReplicator(ctx, t, cfg)
				ctx = namespaces.CreateSource(ctx, t, cfg)
				resources.CreateObject(ctx, t, cfg, namespaces.GetSource(ctx).GetName(), resource.sourceObject)
				_, ctx = namespaces.CreateRandom(ctx, t, cfg)
				_, ctx = namespaces.CreateRandom(ctx, t, cfg)
				return ctx
			}).
			Teardown(cleanup.CleanTestObjects).
			Assess("replicated objects", assessResourcesReplication).
			Feature())

		testFeatures = append(testFeatures, features.New(fmt.Sprintf("controller starts between source namespace creation and source object %s creation", resource.name)).
			WithLabel("resource", resource.name).
			WithLabel("operation", "create").
			Setup(func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
				ctx = setupInitialNamspaces(ctx, t, cfg)
				ctx = namespaces.CreateSource(ctx, t, cfg)
				ctx = controller.SetupReplicator(ctx, t, cfg)
				resources.CreateObject(ctx, t, cfg, namespaces.GetSource(ctx).GetName(), resource.sourceObject)
				_, ctx = namespaces.CreateRandom(ctx, t, cfg)
				_, ctx = namespaces.CreateRandom(ctx, t, cfg)
				return ctx
			}).
			Teardown(cleanup.CleanTestObjects).
			Assess("replicated objects", assessResourcesReplication).
			Feature())

		testFeatures = append(testFeatures, features.New(fmt.Sprintf("controller starts between source object %s creation and first new namespace creation", resource.name)).
			WithLabel("resource", resource.name).
			WithLabel("operation", "create").
			Setup(func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
				ctx = setupInitialNamspaces(ctx, t, cfg)
				ctx = namespaces.CreateSource(ctx, t, cfg)
				resources.CreateObject(ctx, t, cfg, namespaces.GetSource(ctx).GetName(), resource.sourceObject)
				ctx = controller.SetupReplicator(ctx, t, cfg)
				_, ctx = namespaces.CreateRandom(ctx, t, cfg)
				_, ctx = namespaces.CreateRandom(ctx, t, cfg)
				return ctx
			}).
			Teardown(cleanup.CleanTestObjects).
			Assess("replicated objects", assessResourcesReplication).
			Feature())

		testFeatures = append(testFeatures, features.New("controller starts between first new namespace creation and second new namespace creation").
			WithLabel("resource", resource.name).
			WithLabel("operation", "create").
			Setup(func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
				ctx = setupInitialNamspaces(ctx, t, cfg)
				ctx = namespaces.CreateSource(ctx, t, cfg)
				resources.CreateObject(ctx, t, cfg, namespaces.GetSource(ctx).GetName(), resource.sourceObject)
				_, ctx = namespaces.CreateRandom(ctx, t, cfg)
				ctx = controller.SetupReplicator(ctx, t, cfg)
				_, ctx = namespaces.CreateRandom(ctx, t, cfg)
				return ctx
			}).
			Teardown(cleanup.CleanTestObjects).
			Assess("replicated objects", assessResourcesReplication).
			Feature())

		testFeatures = append(testFeatures, features.New("controller starts after second new namespace creation").
			WithLabel("resource", resource.name).
			WithLabel("operation", "create").
			Setup(func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
				ctx = setupInitialNamspaces(ctx, t, cfg)
				ctx = namespaces.CreateSource(ctx, t, cfg)
				resources.CreateObject(ctx, t, cfg, namespaces.GetSource(ctx).GetName(), resource.sourceObject)
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
	testResources := generateResourcesCreationTestData(t)

	testFeatures := []features.Feature{}
	for _, resource := range testResources {
		setupInitialNamspaces := func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			for i := 1; i < 10; i++ {
				_, ctx = namespaces.CreateRandom(ctx, t, cfg)
			}
			return ctx
		}
		assessResourcesReplication := func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			validation.ValidateReplication(ctx, t, cfg, resource.sourceObjectUpdate, resource.objectList,
				validation.WithObjectMatcher(resource.matcher))
			return ctx
		}

		testFeatures = append(testFeatures, features.New("controller updates replicas when source object is updated").
			WithLabel("resource", resource.name).
			WithLabel("operation", "update").
			Setup(func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
				ctx = setupInitialNamspaces(ctx, t, cfg)
				ctx = namespaces.CreateSource(ctx, t, cfg)
				resources.CreateObject(ctx, t, cfg, namespaces.GetSource(ctx).GetName(), resource.sourceObject)
				_, ctx = namespaces.CreateRandom(ctx, t, cfg)
				ctx = controller.SetupReplicator(ctx, t, cfg)
				_, ctx = namespaces.CreateRandom(ctx, t, cfg)
				resources.UpdateObject(ctx, t, cfg, namespaces.GetSource(ctx).GetName(), resource.sourceObjectUpdate)
				return ctx
			}).
			Teardown(cleanup.CleanTestObjects).
			Assess("updated replicas", assessResourcesReplication).
			Feature())

		testFeatures = append(testFeatures, features.New("controller removes replicas when source object source type label is removed").
			WithLabel("resource", resource.name).
			WithLabel("operation", "update").
			Setup(func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
				ctx = setupInitialNamspaces(ctx, t, cfg)
				ctx = namespaces.CreateSource(ctx, t, cfg)
				resources.CreateObject(ctx, t, cfg, namespaces.GetSource(ctx).GetName(), resource.sourceObject)
				_, ctx = namespaces.CreateRandom(ctx, t, cfg)
				ctx = controller.SetupReplicator(ctx, t, cfg)
				_, ctx = namespaces.CreateRandom(ctx, t, cfg)

				updatedObject := resource.sourceObject.DeepCopyObject().(k8s.Object)
				sourceObjectLabels := updatedObject.GetLabels()
				delete(sourceObjectLabels, replicator.ObjectTypeLabelKey)
				resources.UpdateObject(ctx, t, cfg, namespaces.GetSource(ctx).GetName(), updatedObject)
				return ctx
			}).
			Teardown(cleanup.CleanTestObjects).
			Assess("deleted replicas", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
				validation.ValidateResourceDeletion(ctx, t, cfg, resource.sourceObject,
					validation.WithDeletionIgnoredNamespaces(namespaces.GetSource(ctx).GetName()))
				return ctx
			}).
			Feature())

		testFeatures = append(testFeatures, features.New("controller removes replicas when source object type is set to different value").
			WithLabel("resource", resource.name).
			WithLabel("operation", "update").
			Setup(func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
				ctx = setupInitialNamspaces(ctx, t, cfg)
				ctx = namespaces.CreateSource(ctx, t, cfg)
				resources.CreateObject(ctx, t, cfg, namespaces.GetSource(ctx).GetName(), resource.sourceObject)
				_, ctx = namespaces.CreateRandom(ctx, t, cfg)
				ctx = controller.SetupReplicator(ctx, t, cfg)
				_, ctx = namespaces.CreateRandom(ctx, t, cfg)

				updatedObject := resource.sourceObject.DeepCopyObject().(k8s.Object)
				sourceObjectLabels := updatedObject.GetLabels()
				sourceObjectLabels[replicator.ObjectTypeLabelKey] = "ignored"
				resources.UpdateObject(ctx, t, cfg, namespaces.GetSource(ctx).GetName(), updatedObject)
				return ctx
			}).
			Teardown(cleanup.CleanTestObjects).
			Assess("deleted replicas", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
				validation.ValidateResourceDeletion(ctx, t, cfg, resource.sourceObject,
					validation.WithDeletionIgnoredNamespaces(namespaces.GetSource(ctx).GetName()))
				return ctx
			}).
			Feature())
	}

	testenv.Test(t, testFeatures...)
}

func TestResourcesDeletion(t *testing.T) {
	testResources := generateResourcesCreationTestData(t)

	testFeatures := []features.Feature{}
	for _, resource := range testResources {
		setupInitialNamspaces := func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			for i := 1; i < 10; i++ {
				_, ctx = namespaces.CreateRandom(ctx, t, cfg)
			}
			return ctx
		}

		testFeatures = append(testFeatures, features.New("controller deletes all replicas when source object is deleted").
			WithLabel("resource", resource.name).
			WithLabel("operation", "delete").
			Setup(func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
				ctx = setupInitialNamspaces(ctx, t, cfg)
				ctx = namespaces.CreateSource(ctx, t, cfg)
				resources.CreateObject(ctx, t, cfg, namespaces.GetSource(ctx).GetName(), resource.sourceObject)
				ctx = controller.SetupReplicator(ctx, t, cfg)
				validation.ValidateReplication(ctx, t, cfg, resource.sourceObject, resource.objectList,
					validation.WithObjectMatcher(resource.matcher))
				resources.DeleteObjectWithWait(ctx, t, cfg, namespaces.GetSource(ctx).GetName(), resource.sourceObject)
				return ctx
			}).
			Teardown(cleanup.CleanTestObjects).
			Assess("deleted replicas", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
				validation.ValidateResourceDeletion(ctx, t, cfg, resource.sourceObject)
				return ctx
			}).
			Feature())

		testFeatures = append(testFeatures, features.New("controller deletes all replicas when namespace with source object is deleted").
			WithLabel("resource", resource.name).
			WithLabel("operation", "delete").
			Setup(func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
				ctx = setupInitialNamspaces(ctx, t, cfg)
				ctx = namespaces.CreateSource(ctx, t, cfg)
				resources.CreateObject(ctx, t, cfg, namespaces.GetSource(ctx).GetName(), resource.sourceObject)
				ctx = controller.SetupReplicator(ctx, t, cfg)
				validation.ValidateReplication(ctx, t, cfg, resource.sourceObject, resource.objectList,
					validation.WithObjectMatcher(resource.matcher))
				ctx = namespaces.DeleteWithWait(ctx, t, cfg, namespaces.GetSource(ctx))
				return ctx
			}).
			Teardown(cleanup.CleanTestObjects).
			Assess("deleted replicas", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
				validation.ValidateResourceDeletion(ctx, t, cfg, resource.sourceObject)
				return ctx
			}).
			Feature())

		testFeatures = append(testFeatures, features.New("controller recreated deleted replicas").
			WithLabel("resource", resource.name).
			WithLabel("operation", "delete").
			Setup(func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
				ctx = setupInitialNamspaces(ctx, t, cfg)
				ctx = namespaces.CreateSource(ctx, t, cfg)
				resources.CreateObject(ctx, t, cfg, namespaces.GetSource(ctx).GetName(), resource.sourceObject)
				ctx = controller.SetupReplicator(ctx, t, cfg)
				cloneNamespace, ctx := namespaces.CreateRandom(ctx, t, cfg)
				validation.ValidateReplication(ctx, t, cfg, resource.sourceObject, resource.objectList,
					validation.WithObjectMatcher(resource.matcher))
				resources.DeleteObject(ctx, t, cfg, cloneNamespace.GetName(), resource.sourceObject)
				return ctx
			}).
			Teardown(cleanup.CleanTestObjects).
			Assess("recreated replicas", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
				validation.ValidateReplication(ctx, t, cfg, resource.sourceObject, resource.objectList,
					validation.WithObjectMatcher(resource.matcher))
				return ctx
			}).
			Feature())

		testFeatures = append(testFeatures, features.New("controller allows namespace with replica to be deleted").
			WithLabel("resource", resource.name).
			WithLabel("operation", "delete").
			Setup(func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
				ctx = setupInitialNamspaces(ctx, t, cfg)
				ctx = namespaces.CreateSource(ctx, t, cfg)
				resources.CreateObject(ctx, t, cfg, namespaces.GetSource(ctx).GetName(), resource.sourceObject)
				ctx = controller.SetupReplicator(ctx, t, cfg)
				cloneNamespace, ctx := namespaces.CreateRandom(ctx, t, cfg)
				validation.ValidateReplication(ctx, t, cfg, resource.sourceObject, resource.objectList,
					validation.WithObjectMatcher(resource.matcher))
				ctx = namespaces.DeleteWithWait(ctx, t, cfg, cloneNamespace)
				return ctx
			}).
			Teardown(cleanup.CleanTestObjects).
			Assess("remaining replicas", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
				validation.ValidateReplication(ctx, t, cfg, resource.sourceObject, resource.objectList,
					validation.WithObjectMatcher(resource.matcher))
				return ctx
			}).
			Feature())
	}

	testenv.Test(t, testFeatures...)
}
