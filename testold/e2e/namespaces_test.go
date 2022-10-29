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

	"github.com/nadundesilva/k8s-replicator/controllers"
	"github.com/nadundesilva/k8s-replicator/testold/utils/cleanup"
	"github.com/nadundesilva/k8s-replicator/testold/utils/controller"
	"github.com/nadundesilva/k8s-replicator/testold/utils/namespaces"
	"github.com/nadundesilva/k8s-replicator/testold/utils/resources"
	"github.com/nadundesilva/k8s-replicator/testold/utils/testdata"
	"github.com/nadundesilva/k8s-replicator/testold/utils/validation"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/features"
)

func TestNamespaceLabels(t *testing.T) {
	testResources := testdata.GenerateResourceTestData(t)

	testFeatures := []features.Feature{}
	for _, resource := range testResources {
		var testedNs *corev1.Namespace

		testFeatures = append(testFeatures, newFeatureBuilder("controller ignores namespaces with skip label", resource).
			WithLabel("operation", "create").
			Setup(func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
				ctx = controller.SetupReplicator(ctx, t, cfg)
				ctx = namespaces.CreateSource(ctx, t, cfg)
				resources.CreateObject(ctx, t, cfg, namespaces.GetSource(ctx).GetName(), resource.SourceObject)
				_, ctx = namespaces.CreateRandom(ctx, t, cfg)
				testedNs, ctx = namespaces.CreateRandom(ctx, t, cfg, namespaces.WithLabels(map[string]string{
					controllers.NamespaceTypeLabelKey: controllers.NamespaceTypeLabelValueIgnored,
				}))
				return ctx
			}).
			Teardown(cleanup.CleanTestObjects).
			Assess("replicated objects", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
				validation.ValidateReplication(ctx, t, cfg, resource.SourceObject, resource.ObjectList,
					validation.WithObjectMatcher(resource.Matcher),
					validation.WithReplicationIgnoredNamespaces(testedNs.GetName()))
				return ctx
			}).
			Feature())

		testFeatures = append(testFeatures, newFeatureBuilder("controller ignores namespaces with kube prefix", resource).
			WithLabel("operation", "create").
			Setup(func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
				ctx = controller.SetupReplicator(ctx, t, cfg)
				ctx = namespaces.CreateSource(ctx, t, cfg)
				resources.CreateObject(ctx, t, cfg, namespaces.GetSource(ctx).GetName(), resource.SourceObject)
				_, ctx = namespaces.CreateRandom(ctx, t, cfg)
				testedNs, ctx = namespaces.CreateRandom(ctx, t, cfg, namespaces.WithPrefix("kube"))
				return ctx
			}).
			Teardown(cleanup.CleanTestObjects).
			Assess("replicated objects", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
				validation.ValidateReplication(ctx, t, cfg, resource.SourceObject, resource.ObjectList,
					validation.WithObjectMatcher(resource.Matcher),
					validation.WithReplicationIgnoredNamespaces(testedNs.GetName()))
				return ctx
			}).
			Feature())

		testFeatures = append(testFeatures, newFeatureBuilder("controller replicates namespaces with kube prefix with replicate label", resource).
			WithLabel("operation", "create").
			Setup(func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
				ctx = controller.SetupReplicator(ctx, t, cfg)
				ctx = namespaces.CreateSource(ctx, t, cfg)
				resources.CreateObject(ctx, t, cfg, namespaces.GetSource(ctx).GetName(), resource.SourceObject)
				_, ctx = namespaces.CreateRandom(ctx, t, cfg)
				testedNs, ctx = namespaces.CreateRandom(ctx, t, cfg, namespaces.WithPrefix("kube"),
					namespaces.WithLabels(map[string]string{
						controllers.NamespaceTypeLabelKey: controllers.NamespaceTypeLabelValueManaged,
					}))
				return ctx
			}).
			Teardown(cleanup.CleanTestObjects).
			Assess("replicated objects", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
				validation.ValidateReplication(ctx, t, cfg, resource.SourceObject, resource.ObjectList,
					validation.WithObjectMatcher(resource.Matcher), validation.WithReplicatedNamespaces(testedNs.GetName()))
				return ctx
			}).
			Feature())

		testFeatures = append(testFeatures, newFeatureBuilder("controller ignores controller namespace", resource).
			WithLabel("operation", "create").
			Setup(func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
				ctx = controller.SetupReplicator(ctx, t, cfg)
				ctx = namespaces.CreateSource(ctx, t, cfg)
				resources.CreateObject(ctx, t, cfg, namespaces.GetSource(ctx).GetName(), resource.SourceObject)
				_, ctx = namespaces.CreateRandom(ctx, t, cfg)
				_, ctx = namespaces.CreateRandom(ctx, t, cfg)
				return ctx
			}).
			Teardown(cleanup.CleanTestObjects).
			Assess("replicated objects", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
				validation.ValidateReplication(ctx, t, cfg, resource.SourceObject, resource.ObjectList,
					validation.WithObjectMatcher(resource.Matcher),
					validation.WithReplicationIgnoredNamespaces(controller.GetNamspace(ctx)))
				return ctx
			}).
			Feature())

		testFeatures = append(testFeatures, newFeatureBuilder("controller creates replica in controller namespace with replicate label", resource).
			WithLabel("operation", "create").
			Setup(func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
				ctx = controller.SetupReplicator(ctx, t, cfg, controller.WithNamespaceLabels(map[string]string{
					controllers.NamespaceTypeLabelKey: controllers.NamespaceTypeLabelValueManaged,
				}))
				ctx = namespaces.CreateSource(ctx, t, cfg)
				resources.CreateObject(ctx, t, cfg, namespaces.GetSource(ctx).GetName(), resource.SourceObject)
				_, ctx = namespaces.CreateRandom(ctx, t, cfg)
				_, ctx = namespaces.CreateRandom(ctx, t, cfg)
				return ctx
			}).
			Teardown(cleanup.CleanTestObjects).
			Assess("replicated objects", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
				validation.ValidateReplication(ctx, t, cfg, resource.SourceObject, resource.ObjectList,
					validation.WithObjectMatcher(resource.Matcher),
					validation.WithReplicatedNamespaces(controller.GetNamspace(ctx)))
				return ctx
			}).
			Feature())

		testFeatures = append(testFeatures, newFeatureBuilder("controller deletes replica when ignore label is added to namespace", resource).
			WithLabel("operation", "create").
			Setup(func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
				ctx = controller.SetupReplicator(ctx, t, cfg)
				ctx = namespaces.CreateSource(ctx, t, cfg)
				resources.CreateObject(ctx, t, cfg, namespaces.GetSource(ctx).GetName(), resource.SourceObject)
				_, ctx = namespaces.CreateRandom(ctx, t, cfg)
				testedNs, ctx = namespaces.CreateRandom(ctx, t, cfg)
				validation.ValidateReplication(ctx, t, cfg, resource.SourceObject, resource.ObjectList,
					validation.WithObjectMatcher(resource.Matcher))

				testedNs.GetLabels()[controllers.NamespaceTypeLabelKey] = controllers.NamespaceTypeLabelValueIgnored
				namespaces.Update(ctx, t, cfg, testedNs)
				return ctx
			}).
			Teardown(cleanup.CleanTestObjects).
			Assess("ignored object", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
				validation.ValidateReplication(ctx, t, cfg, resource.SourceObject, resource.ObjectList,
					validation.WithObjectMatcher(resource.Matcher),
					validation.WithReplicationIgnoredNamespaces(testedNs.GetName()))
				return ctx
			}).
			Feature())

		testFeatures = append(testFeatures, newFeatureBuilder("controller creates replica when replicate label is added to ignored namespace", resource).
			WithLabel("operation", "create").
			Setup(func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
				ctx = controller.SetupReplicator(ctx, t, cfg)
				ctx = namespaces.CreateSource(ctx, t, cfg)
				resources.CreateObject(ctx, t, cfg, namespaces.GetSource(ctx).GetName(), resource.SourceObject)
				_, ctx = namespaces.CreateRandom(ctx, t, cfg)
				testedNs, ctx = namespaces.CreateRandom(ctx, t, cfg, namespaces.WithPrefix("kube-"))
				validation.ValidateReplication(ctx, t, cfg, resource.SourceObject, resource.ObjectList)

				testedNs.GetLabels()[controllers.NamespaceTypeLabelKey] = controllers.NamespaceTypeLabelValueManaged
				namespaces.Update(ctx, t, cfg, testedNs)
				return ctx
			}).
			Teardown(cleanup.CleanTestObjects).
			Assess("ignored object", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
				validation.ValidateReplication(ctx, t, cfg, resource.SourceObject, resource.ObjectList,
					validation.WithObjectMatcher(resource.Matcher),
					validation.WithReplicatedNamespaces(testedNs.GetName()))
				return ctx
			}).
			Feature())
	}

	testenv.Test(t, testFeatures...)
}
