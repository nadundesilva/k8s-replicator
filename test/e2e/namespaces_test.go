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
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/features"
)

func TestNamespaceCreate(t *testing.T) {
	testResources := testdata.GenerateResourceTestData()

	testFeatures := []features.Feature{}
	for _, resource := range testResources {
		var testedNs *corev1.Namespace

		testFeatures = append(testFeatures, newFeatureBuilder("controller creates a replica when a new namespace is created", resource).
			WithLabel("operation", "create").
			Setup(func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
				ctx = controller.SetupReplicator(ctx, t, cfg)
				ctx = namespaces.CreateSource(ctx, t, cfg)
				resources.CreateObject(ctx, t, cfg, common.GetSourceObjectNamespace(ctx).GetName(), resource.SourceObject())
				_, ctx = namespaces.CreateRandom(ctx, t, cfg)
				return ctx
			}).
			Teardown(cleanup.CleanTestObjects).
			Assess("replicated objects", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
				validation.ValidateReplication(ctx, t, cfg, resource.SourceObject(), resource.EmptyObjectList(),
					validation.WithReplicationObjectMatcher(resource.IsEqual))
				return ctx
			}).
			Feature())

		testFeatures = append(testFeatures, newFeatureBuilder("controller ignores when a new namespace with ignored label", resource).
			WithLabel("operation", "create").
			Setup(func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
				ctx = controller.SetupReplicator(ctx, t, cfg)
				ctx = namespaces.CreateSource(ctx, t, cfg)
				resources.CreateObject(ctx, t, cfg, common.GetSourceObjectNamespace(ctx).GetName(), resource.SourceObject())
				testedNs, ctx = namespaces.CreateRandom(ctx, t, cfg, namespaces.WithLabels(map[string]string{
					common.NamespaceTypeLabelKey: common.NamespaceTypeLabelValueIgnored,
				}))
				return ctx
			}).
			Teardown(cleanup.CleanTestObjects).
			Assess("replicated objects", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
				validation.ValidateReplication(ctx, t, cfg, resource.SourceObject(), resource.EmptyObjectList(),
					validation.WithReplicationObjectMatcher(resource.IsEqual),
					validation.WithReplicationIgnoredNamespaces(testedNs.GetName()))
				return ctx
			}).
			Feature())

		testFeatures = append(testFeatures, newFeatureBuilder("controller ignores when a new namespace with kube prefixed name", resource).
			WithLabel("operation", "create").
			Setup(func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
				ctx = controller.SetupReplicator(ctx, t, cfg)
				ctx = namespaces.CreateSource(ctx, t, cfg)
				resources.CreateObject(ctx, t, cfg, common.GetSourceObjectNamespace(ctx).GetName(), resource.SourceObject())
				testedNs, ctx = namespaces.CreateRandom(ctx, t, cfg, namespaces.WithPrefix("kube"))
				return ctx
			}).
			Teardown(cleanup.CleanTestObjects).
			Assess("replicated objects", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
				validation.ValidateReplication(ctx, t, cfg, resource.SourceObject(), resource.EmptyObjectList(),
					validation.WithReplicationObjectMatcher(resource.IsEqual),
					validation.WithReplicationIgnoredNamespaces(testedNs.GetName()))
				return ctx
			}).
			Feature())

		testFeatures = append(testFeatures, newFeatureBuilder("controller ignores when a new namespace with kube prefixed name and replicate label", resource).
			WithLabel("operation", "create").
			Setup(func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
				ctx = controller.SetupReplicator(ctx, t, cfg)
				ctx = namespaces.CreateSource(ctx, t, cfg)
				resources.CreateObject(ctx, t, cfg, common.GetSourceObjectNamespace(ctx).GetName(), resource.SourceObject())
				testedNs, ctx = namespaces.CreateRandom(ctx, t, cfg, namespaces.WithPrefix("kube"),
					namespaces.WithLabels(map[string]string{
						common.NamespaceTypeLabelKey: common.NamespaceTypeLabelValueManaged,
					}))
				return ctx
			}).
			Teardown(cleanup.CleanTestObjects).
			Assess("replicated objects", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
				validation.ValidateReplication(ctx, t, cfg, resource.SourceObject(), resource.EmptyObjectList(),
					validation.WithReplicationObjectMatcher(resource.IsEqual),
					validation.WithReplicatedNamespaces(testedNs.GetName()))
				return ctx
			}).
			Feature())
	}

	testenv.Test(t, testFeatures...)
}

func TestNamespaceLabelsUpdate(t *testing.T) {
	testResources := testdata.GenerateResourceTestData()

	testFeatures := []features.Feature{}
	for _, resource := range testResources {
		var testedNs *corev1.Namespace

		testFeatures = append(testFeatures, newFeatureBuilder("controller ignores namespaces with ignored label", resource).
			WithLabel("operation", "create").
			Setup(func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
				ctx = controller.SetupReplicator(ctx, t, cfg)
				ctx = namespaces.CreateSource(ctx, t, cfg)
				resources.CreateObject(ctx, t, cfg, common.GetSourceObjectNamespace(ctx).GetName(), resource.SourceObject())
				_, ctx = namespaces.CreateRandom(ctx, t, cfg)
				testedNs, ctx = namespaces.CreateRandom(ctx, t, cfg, namespaces.WithLabels(map[string]string{
					common.NamespaceTypeLabelKey: common.NamespaceTypeLabelValueIgnored,
				}))
				return ctx
			}).
			Teardown(cleanup.CleanTestObjects).
			Assess("replicated objects", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
				validation.ValidateReplication(ctx, t, cfg, resource.SourceObject(), resource.EmptyObjectList(),
					validation.WithReplicationObjectMatcher(resource.IsEqual),
					validation.WithReplicationIgnoredNamespaces(testedNs.GetName()))
				return ctx
			}).
			Feature())

		testFeatures = append(testFeatures, newFeatureBuilder("controller ignores namespaces with kube prefixed names", resource).
			WithLabel("operation", "create").
			Setup(func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
				ctx = controller.SetupReplicator(ctx, t, cfg)
				ctx = namespaces.CreateSource(ctx, t, cfg)
				resources.CreateObject(ctx, t, cfg, common.GetSourceObjectNamespace(ctx).GetName(), resource.SourceObject())
				_, ctx = namespaces.CreateRandom(ctx, t, cfg)
				testedNs, ctx = namespaces.CreateRandom(ctx, t, cfg, namespaces.WithPrefix("kube"))
				return ctx
			}).
			Teardown(cleanup.CleanTestObjects).
			Assess("replicated objects", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
				validation.ValidateReplication(ctx, t, cfg, resource.SourceObject(), resource.EmptyObjectList(),
					validation.WithReplicationObjectMatcher(resource.IsEqual),
					validation.WithReplicationIgnoredNamespaces(testedNs.GetName()))
				return ctx
			}).
			Feature())

		testFeatures = append(testFeatures, newFeatureBuilder("controller replicates namespaces with kube prefixed name and replicate label", resource).
			WithLabel("operation", "create").
			Setup(func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
				ctx = controller.SetupReplicator(ctx, t, cfg)
				ctx = namespaces.CreateSource(ctx, t, cfg)
				resources.CreateObject(ctx, t, cfg, common.GetSourceObjectNamespace(ctx).GetName(), resource.SourceObject())
				_, ctx = namespaces.CreateRandom(ctx, t, cfg)
				testedNs, ctx = namespaces.CreateRandom(ctx, t, cfg, namespaces.WithPrefix("kube"),
					namespaces.WithLabels(map[string]string{
						common.NamespaceTypeLabelKey: common.NamespaceTypeLabelValueManaged,
					}))
				return ctx
			}).
			Teardown(cleanup.CleanTestObjects).
			Assess("replicated objects", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
				validation.ValidateReplication(ctx, t, cfg, resource.SourceObject(), resource.EmptyObjectList(),
					validation.WithReplicationObjectMatcher(resource.IsEqual),
					validation.WithReplicatedNamespaces(testedNs.GetName()))
				return ctx
			}).
			Feature())

		testFeatures = append(testFeatures, newFeatureBuilder("controller ignores controller namespace", resource).
			WithLabel("operation", "create").
			Setup(func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
				ctx = controller.SetupReplicator(ctx, t, cfg)
				ctx = namespaces.CreateSource(ctx, t, cfg)
				resources.CreateObject(ctx, t, cfg, common.GetSourceObjectNamespace(ctx).GetName(), resource.SourceObject())
				_, ctx = namespaces.CreateRandom(ctx, t, cfg)
				_, ctx = namespaces.CreateRandom(ctx, t, cfg)
				return ctx
			}).
			Teardown(cleanup.CleanTestObjects).
			Assess("replicated objects", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
				validation.ValidateReplication(ctx, t, cfg, resource.SourceObject(), resource.EmptyObjectList(),
					validation.WithReplicationObjectMatcher(resource.IsEqual),
					validation.WithReplicationIgnoredNamespaces(common.GetControllerNamespace(ctx)))
				return ctx
			}).
			Feature())

		testFeatures = append(testFeatures, newFeatureBuilder("controller creates replica in controller namespace with replicate label", resource).
			WithLabel("operation", "create").
			Setup(func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
				ctx = controller.SetupReplicator(ctx, t, cfg, controller.WithNamespaceLabels(map[string]string{
					common.NamespaceTypeLabelKey: common.NamespaceTypeLabelValueManaged,
				}))
				ctx = namespaces.CreateSource(ctx, t, cfg)
				resources.CreateObject(ctx, t, cfg, common.GetSourceObjectNamespace(ctx).GetName(), resource.SourceObject())
				_, ctx = namespaces.CreateRandom(ctx, t, cfg)
				_, ctx = namespaces.CreateRandom(ctx, t, cfg)
				return ctx
			}).
			Teardown(cleanup.CleanTestObjects).
			Assess("replicated objects", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
				validation.ValidateReplication(ctx, t, cfg, resource.SourceObject(), resource.EmptyObjectList(),
					validation.WithReplicationObjectMatcher(resource.IsEqual),
					validation.WithReplicatedNamespaces(common.GetControllerNamespace(ctx)))
				return ctx
			}).
			Feature())

		testFeatures = append(testFeatures, newFeatureBuilder("controller deletes replica when ignored label is added to namespace", resource).
			WithLabel("operation", "create").
			Setup(func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
				ctx = controller.SetupReplicator(ctx, t, cfg)
				ctx = namespaces.CreateSource(ctx, t, cfg)
				resources.CreateObject(ctx, t, cfg, common.GetSourceObjectNamespace(ctx).GetName(), resource.SourceObject())
				_, ctx = namespaces.CreateRandom(ctx, t, cfg)
				testedNs, ctx = namespaces.CreateRandom(ctx, t, cfg)
				validation.ValidateReplication(ctx, t, cfg, resource.SourceObject(), resource.EmptyObjectList(),
					validation.WithReplicationObjectMatcher(resource.IsEqual))

				testedNs.GetLabels()[common.NamespaceTypeLabelKey] = common.NamespaceTypeLabelValueIgnored
				namespaces.Update(ctx, t, cfg, testedNs)
				return ctx
			}).
			Teardown(cleanup.CleanTestObjects).
			Assess("ignored object", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
				validation.ValidateReplication(ctx, t, cfg, resource.SourceObject(), resource.EmptyObjectList(),
					validation.WithReplicationObjectMatcher(resource.IsEqual),
					validation.WithReplicationIgnoredNamespaces(testedNs.GetName()))
				return ctx
			}).
			Feature())

		testFeatures = append(testFeatures, newFeatureBuilder("controller creates replica when managed label is added to ignored namespace", resource).
			WithLabel("operation", "create").
			Setup(func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
				ctx = controller.SetupReplicator(ctx, t, cfg)
				ctx = namespaces.CreateSource(ctx, t, cfg)
				resources.CreateObject(ctx, t, cfg, common.GetSourceObjectNamespace(ctx).GetName(), resource.SourceObject())
				_, ctx = namespaces.CreateRandom(ctx, t, cfg)
				testedNs, ctx = namespaces.CreateRandom(ctx, t, cfg, namespaces.WithPrefix("kube"))
				validation.ValidateReplication(ctx, t, cfg, resource.SourceObject(), resource.EmptyObjectList())

				testedNs.GetLabels()[common.NamespaceTypeLabelKey] = common.NamespaceTypeLabelValueManaged
				namespaces.Update(ctx, t, cfg, testedNs)
				return ctx
			}).
			Teardown(cleanup.CleanTestObjects).
			Assess("ignored object", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
				validation.ValidateReplication(ctx, t, cfg, resource.SourceObject(), resource.EmptyObjectList(),
					validation.WithReplicationObjectMatcher(resource.IsEqual),
					validation.WithReplicatedNamespaces(testedNs.GetName()))
				return ctx
			}).
			Feature())
	}

	testenv.Test(t, testFeatures...)
}

func TestNamespaceDelete(t *testing.T) {
	testResources := testdata.GenerateResourceTestData()

	testFeatures := []features.Feature{}
	for _, resource := range testResources {
		var ignoredNs *corev1.Namespace

		assessResourcesReplication := func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			ignoredNamespaces := []string{}
			if ignoredNs != nil {
				ignoredNamespaces = append(ignoredNamespaces, ignoredNs.GetName())
			}

			validation.ValidateReplication(ctx, t, cfg, resource.SourceObject(), resource.EmptyObjectList(),
				validation.WithReplicationObjectMatcher(resource.IsEqual),
				validation.WithReplicationIgnoredNamespaces(ignoredNamespaces...))
			return ctx
		}

		testFeatures = append(testFeatures, newFeatureBuilder("controller deleted the replica in the namespace when it is deleted", resource).
			WithLabel("operation", "delete").
			Setup(func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
				ctx = controller.SetupReplicator(ctx, t, cfg)
				ctx = namespaces.CreateSource(ctx, t, cfg)
				resources.CreateObject(ctx, t, cfg, common.GetSourceObjectNamespace(ctx).GetName(), resource.SourceObject())
				testedNs, ctx := namespaces.CreateRandom(ctx, t, cfg)
				assessResourcesReplication(ctx, t, cfg)

				namespaces.DeleteWithWait(ctx, t, cfg, testedNs)
				assessResourcesReplication(ctx, t, cfg)
				return ctx
			}).
			Teardown(cleanup.CleanTestObjects).
			Assess("replicated objects", assessResourcesReplication).
			Feature())

		testFeatures = append(testFeatures, newFeatureBuilder("controller controller ignores the deletion of ignored namespaces", resource).
			WithLabel("operation", "delete").
			Setup(func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
				ctx = controller.SetupReplicator(ctx, t, cfg)
				ctx = namespaces.CreateSource(ctx, t, cfg)
				resources.CreateObject(ctx, t, cfg, common.GetSourceObjectNamespace(ctx).GetName(), resource.SourceObject())
				ignoredNs, ctx = namespaces.CreateRandom(ctx, t, cfg, namespaces.WithLabels(map[string]string{
					common.NamespaceTypeLabelKey: common.NamespaceTypeLabelValueIgnored,
				}))
				assessResourcesReplication(ctx, t, cfg)

				namespaces.DeleteWithWait(ctx, t, cfg, ignoredNs)
				assessResourcesReplication(ctx, t, cfg)
				return ctx
			}).
			Teardown(cleanup.CleanTestObjects).
			Assess("replicated objects", assessResourcesReplication).
			Feature())
	}

	testenv.Test(t, testFeatures...)
}
