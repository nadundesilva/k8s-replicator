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
package testdata

import (
	"testing"

	"github.com/nadundesilva/k8s-replicator/pkg/replicator"
	"github.com/nadundesilva/k8s-replicator/test/utils/validation"
	"sigs.k8s.io/e2e-framework/klient/k8s"
)

type Resource struct {
	Name               string
	ObjectList         k8s.ObjectList
	SourceObject       k8s.Object
	SourceObjectUpdate k8s.Object
	Matcher            validation.ObjectMatcher
}

func GenerateResourceTestData(t *testing.T) []Resource {
	resources := []Resource{
		GenerateSecretTestDatum(),
		GenerateConfigMapTestDatum(),
		GenerateNetworkPolicyTestDatum(),
	}
	filteredResources := []Resource{}
	for _, resource := range resources {
		if !isTested(resource) {
			continue
		}
		filteredResources = append(filteredResources, resource)
	}
	return filteredResources
}

func process(resource Resource) Resource {
	resource.SourceObjectUpdate.SetName(resource.SourceObject.GetName())

	updateSourceObjectLabels := func(sourceObject k8s.Object) {
		labels := sourceObject.GetLabels()
		labels[replicator.ObjectTypeLabelKey] = replicator.ObjectTypeLabelValueSource
	}
	updateSourceObjectLabels(resource.SourceObject)
	updateSourceObjectLabels(resource.SourceObjectUpdate)

	return resource
}

func toPointer[T interface{}](val T) *T {
	return &val
}
