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
	"encoding/base64"
	"fmt"
	"reflect"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/e2e-framework/klient/k8s"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/features"
)

func TestResources(t *testing.T) {
	resources := []struct {
		name         string
		typeMeta     metav1.Object
		objectList   k8s.ObjectList
		sourceObject k8s.Object
		matcher      objectMatcher
	}{
		{
			name:       "secret",
			objectList: &corev1.SecretList{},
			sourceObject: &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name: envconf.RandomName("source-secret", 32),
					Labels: map[string]string{
						"e2e-tests.replicator.io/test-label-key": "test-label-value",
					},
					Annotations: map[string]string{
						"e2e-tests.replicator.io/test-annotation-key": "test-annotation-value",
					},
				},
				Data: map[string][]byte{
					"data-item-one-key": []byte(base64.StdEncoding.EncodeToString([]byte("data-item-one-value"))),
				},
			},
			matcher: func(sourceObject k8s.Object, targetObject k8s.Object) bool {
				sourceSecret := sourceObject.(*corev1.Secret)
				targetSecret := targetObject.(*corev1.Secret)
				if !reflect.DeepEqual(sourceSecret.Data, targetSecret.Data) {
					t.Errorf("secret data not equal; want %s, got %s",
						sourceSecret.Data, targetSecret.Data)
				}
				return true
			},
		},
	}

	testFeatures := []features.Feature{}
	for _, resource := range resources {
		sourceNamespaceName := envconf.RandomName(namespacePrefix+"source-ns", 32)
		feature := features.New(fmt.Sprintf("replicate when new %s is created", resource.name)).
			WithLabel("resource", resource.name).
			Setup(func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
				setupReplicatorController(ctx, t, cfg)
				createSourceNamespace(ctx, t, cfg, sourceNamespaceName)
				return ctx
			}).
			Assess("new resource added to existing namespace", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
				createSourceObject(ctx, t, cfg, sourceNamespaceName, resource.sourceObject)
				validateReplication(ctx, t, cfg, resource.sourceObject, resource.objectList, resource.matcher)
				return ctx
			}).
			Feature()
		testFeatures = append(testFeatures, feature)
	}

	testenv.Test(t, testFeatures...)
}
