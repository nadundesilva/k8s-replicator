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

	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/features"
)

func TestResourcesCreation(t *testing.T) {
	resources := generateResourcesCreationTestData(t)

	testFeatures := []features.Feature{}
	for _, resource := range resources {
		setupInitialNamspaces := func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			for i := 1; i < 10; i++ {
				_, ctx = createRandomNamespace(ctx, t, cfg)
			}
			return ctx
		}
		assessResourcesReplication := func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			validateReplication(ctx, t, cfg, resource.sourceObject, resource.objectList, resource.matcher)
			return ctx
		}

		testFeatures = append(testFeatures, features.New("controller starts before initial namespaces creation").
			WithLabel("resource", resource.name).
			WithLabel("operation", "create").
			Setup(func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
				ctx = setupInitialNamspaces(ctx, t, cfg)
				ctx = setupReplicatorController(ctx, t, cfg)
				sourceNamespace, ctx := createRandomNamespace(ctx, t, cfg)
				createSourceObject(ctx, t, cfg, sourceNamespace.GetName(), resource.sourceObject)
				_, ctx = createRandomNamespace(ctx, t, cfg)
				_, ctx = createRandomNamespace(ctx, t, cfg)
				return ctx
			}).
			Teardown(cleanupTestObjects).
			Assess("replicated objects", assessResourcesReplication).
			Feature())

		testFeatures = append(testFeatures, features.New(fmt.Sprintf("controller starts netween initial namespaces creation and source namespace %s creation", resource.name)).
			WithLabel("resource", resource.name).
			WithLabel("operation", "create").
			Setup(func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
				ctx = setupInitialNamspaces(ctx, t, cfg)
				ctx = setupReplicatorController(ctx, t, cfg)
				sourceNamespace, ctx := createRandomNamespace(ctx, t, cfg)
				createSourceObject(ctx, t, cfg, sourceNamespace.GetName(), resource.sourceObject)
				_, ctx = createRandomNamespace(ctx, t, cfg)
				_, ctx = createRandomNamespace(ctx, t, cfg)
				return ctx
			}).
			Teardown(cleanupTestObjects).
			Assess("replicated objects", assessResourcesReplication).
			Feature())

		testFeatures = append(testFeatures, features.New(fmt.Sprintf("controller starts between source namespace creation and source object %s creation", resource.name)).
			WithLabel("resource", resource.name).
			WithLabel("operation", "create").
			Setup(func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
				ctx = setupInitialNamspaces(ctx, t, cfg)
				sourceNamespace, ctx := createRandomNamespace(ctx, t, cfg)
				ctx = setupReplicatorController(ctx, t, cfg)
				createSourceObject(ctx, t, cfg, sourceNamespace.GetName(), resource.sourceObject)
				_, ctx = createRandomNamespace(ctx, t, cfg)
				_, ctx = createRandomNamespace(ctx, t, cfg)
				return ctx
			}).
			Teardown(cleanupTestObjects).
			Assess("replicated objects", assessResourcesReplication).
			Feature())

		testFeatures = append(testFeatures, features.New(fmt.Sprintf("controller starts between source object %s creation and first new namespace creation", resource.name)).
			WithLabel("resource", resource.name).
			WithLabel("operation", "create").
			Setup(func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
				ctx = setupInitialNamspaces(ctx, t, cfg)
				sourceNamespace, ctx := createRandomNamespace(ctx, t, cfg)
				createSourceObject(ctx, t, cfg, sourceNamespace.GetName(), resource.sourceObject)
				ctx = setupReplicatorController(ctx, t, cfg)
				_, ctx = createRandomNamespace(ctx, t, cfg)
				_, ctx = createRandomNamespace(ctx, t, cfg)
				return ctx
			}).
			Teardown(cleanupTestObjects).
			Assess("replicated objects", assessResourcesReplication).
			Feature())

		testFeatures = append(testFeatures, features.New("controller starts between first new namespace creation and second new namespace creation").
			WithLabel("resource", resource.name).
			WithLabel("operation", "create").
			Setup(func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
				ctx = setupInitialNamspaces(ctx, t, cfg)
				sourceNamespace, ctx := createRandomNamespace(ctx, t, cfg)
				createSourceObject(ctx, t, cfg, sourceNamespace.GetName(), resource.sourceObject)
				_, ctx = createRandomNamespace(ctx, t, cfg)
				ctx = setupReplicatorController(ctx, t, cfg)
				_, ctx = createRandomNamespace(ctx, t, cfg)
				return ctx
			}).
			Teardown(cleanupTestObjects).
			Assess("replicated objects", assessResourcesReplication).
			Feature())

		testFeatures = append(testFeatures, features.New("controller starts after second new namespace creation").
			WithLabel("resource", resource.name).
			WithLabel("operation", "create").
			Setup(func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
				ctx = setupInitialNamspaces(ctx, t, cfg)
				sourceNamespace, ctx := createRandomNamespace(ctx, t, cfg)
				createSourceObject(ctx, t, cfg, sourceNamespace.GetName(), resource.sourceObject)
				_, ctx = createRandomNamespace(ctx, t, cfg)
				_, ctx = createRandomNamespace(ctx, t, cfg)
				ctx = setupReplicatorController(ctx, t, cfg)
				return ctx
			}).
			Teardown(cleanupTestObjects).
			Assess("replicated objects", assessResourcesReplication).
			Feature())
	}

	testenv.Test(t, testFeatures...)
}

func TestResourcesDeletion(t *testing.T) {
	resources := generateResourcesCreationTestData(t)

	testFeatures := []features.Feature{}
	for _, resource := range resources {
		setupInitialNamspaces := func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			for i := 1; i < 10; i++ {
				_, ctx = createRandomNamespace(ctx, t, cfg)
			}
			return ctx
		}

		testFeatures = append(testFeatures, features.New("controller deletes all clones when source object is deleted").
			WithLabel("resource", resource.name).
			WithLabel("operation", "delete").
			Setup(func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
				ctx = setupInitialNamspaces(ctx, t, cfg)
				sourceNamespace, ctx := createRandomNamespace(ctx, t, cfg)
				createSourceObject(ctx, t, cfg, sourceNamespace.GetName(), resource.sourceObject)
				ctx = setupReplicatorController(ctx, t, cfg)
				validateReplication(ctx, t, cfg, resource.sourceObject, resource.objectList, resource.matcher)
				deleteObjectWithWait(ctx, t, cfg, sourceNamespace.GetName(), resource.sourceObject)
				return ctx
			}).
			Teardown(cleanupTestObjects).
			Assess("deleted clones", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
				validateResourceDeletion(ctx, t, cfg, resource.sourceObject)
				return ctx
			}).
			Feature())

		testFeatures = append(testFeatures, features.New("controller deletes all clones when namespace with source object is deleted").
			WithLabel("resource", resource.name).
			WithLabel("operation", "delete").
			Setup(func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
				ctx = setupInitialNamspaces(ctx, t, cfg)
				sourceNamespace, ctx := createRandomNamespace(ctx, t, cfg)
				createSourceObject(ctx, t, cfg, sourceNamespace.GetName(), resource.sourceObject)
				ctx = setupReplicatorController(ctx, t, cfg)
				validateReplication(ctx, t, cfg, resource.sourceObject, resource.objectList, resource.matcher)
				ctx = deleteNamespaceWithWait(ctx, t, cfg, sourceNamespace)
				return ctx
			}).
			Teardown(cleanupTestObjects).
			Assess("deleted clones", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
				validateResourceDeletion(ctx, t, cfg, resource.sourceObject)
				return ctx
			}).
			Feature())

		testFeatures = append(testFeatures, features.New("controller recreated deleted clones").
			WithLabel("resource", resource.name).
			WithLabel("operation", "delete").
			Setup(func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
				ctx = setupInitialNamspaces(ctx, t, cfg)
				sourceNamespace, ctx := createRandomNamespace(ctx, t, cfg)
				createSourceObject(ctx, t, cfg, sourceNamespace.GetName(), resource.sourceObject)
				ctx = setupReplicatorController(ctx, t, cfg)
				cloneNamespace, ctx := createRandomNamespace(ctx, t, cfg)
				validateReplication(ctx, t, cfg, resource.sourceObject, resource.objectList, resource.matcher)
				deleteObject(ctx, t, cfg, cloneNamespace.GetName(), resource.sourceObject)
				return ctx
			}).
			Teardown(cleanupTestObjects).
			Assess("recreated clones", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
				validateReplication(ctx, t, cfg, resource.sourceObject, resource.objectList, resource.matcher)
				return ctx
			}).
			Feature())

		testFeatures = append(testFeatures, features.New("controller allows namespace with clone to be deleted").
			WithLabel("resource", resource.name).
			WithLabel("operation", "delete").
			Setup(func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
				ctx = setupInitialNamspaces(ctx, t, cfg)
				sourceNamespace, ctx := createRandomNamespace(ctx, t, cfg)
				createSourceObject(ctx, t, cfg, sourceNamespace.GetName(), resource.sourceObject)
				ctx = setupReplicatorController(ctx, t, cfg)
				cloneNamespace, ctx := createRandomNamespace(ctx, t, cfg)
				validateReplication(ctx, t, cfg, resource.sourceObject, resource.objectList, resource.matcher)
				ctx = deleteNamespaceWithWait(ctx, t, cfg, cloneNamespace)
				return ctx
			}).
			Teardown(cleanupTestObjects).
			Assess("remaining clones", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
				validateReplication(ctx, t, cfg, resource.sourceObject, resource.objectList, resource.matcher)
				return ctx
			}).
			Feature())
	}

	testenv.Test(t, testFeatures...)
}
