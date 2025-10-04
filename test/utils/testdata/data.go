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
	"github.com/nadundesilva/k8s-replicator/test/utils/common"
	"github.com/nadundesilva/k8s-replicator/test/utils/validation"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type resourceData struct {
	Name               string
	SourceObject       client.Object
	SourceObjectUpdate client.Object
	EmptyObject        client.Object
	EmptyObjectList    client.ObjectList
	IsEqual            validation.ObjectMatcher
}

func (d *resourceData) GenerateResource() Resource {
	return Resource{
		data:    d,
		Name:    d.Name,
		IsEqual: d.IsEqual,
	}
}

type Resource struct {
	data    *resourceData
	Name    string
	IsEqual validation.ObjectMatcher
}

func (r *Resource) SourceObject() client.Object {
	obj := r.data.SourceObject.DeepCopyObject().(client.Object)
	objLabels := obj.GetLabels()
	if objLabels == nil {
		objLabels = map[string]string{}
	}
	objLabels[common.ObjectTypeLabelKey] = common.ObjectTypeLabelValueReplicated
	obj.SetLabels(objLabels)
	return obj
}

func (r *Resource) SourceObjectUpdate() client.Object {
	obj := r.data.SourceObjectUpdate.DeepCopyObject().(client.Object)
	objLabels := obj.GetLabels()
	if objLabels == nil {
		objLabels = map[string]string{}
	}
	objLabels[common.ObjectTypeLabelKey] = common.ObjectTypeLabelValueReplicated
	obj.SetLabels(objLabels)
	return obj
}

func (r *Resource) EmptyObject() client.Object {
	return r.data.EmptyObject.DeepCopyObject().(client.Object)
}

func (r *Resource) EmptyObjectList() client.ObjectList {
	return r.data.EmptyObjectList.DeepCopyObject().(client.ObjectList)
}

func GenerateResourceTestData() []Resource {
	resources := []Resource{
		generateSecretTestDatum(),
		generateConfigMapTestDatum(),
		generateNetworkPolicyTestDatum(),
		generateServiceAccountTestDatum(),
		generateRoleTestDatum(),
		generateRoleBindingTestDatum(),
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

func process(resourceData resourceData) Resource {
	resourceData.SourceObjectUpdate.SetName(resourceData.SourceObject.GetName())

	updateObjectMetadata := func(object client.Object) {
		labels := object.GetLabels()
		if labels == nil {
			labels = map[string]string{}
		}
		labels["unit-test-label-key"] = "test-label-value"

		annotations := object.GetAnnotations()
		if annotations == nil {
			annotations = map[string]string{}
		}
		annotations["unit-test-annotation-key"] = "test-annotation-value"
	}
	updateObjectMetadata(resourceData.SourceObject)
	updateObjectMetadata(resourceData.SourceObjectUpdate)

	return resourceData.GenerateResource()
}
