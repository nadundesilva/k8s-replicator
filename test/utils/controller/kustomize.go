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
package controller

import (
	"fmt"
	"testing"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/kustomize/api/krusty"
	"sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/kustomize/kyaml/filesys"
)

type kustomizeBuildObject struct {
	obj  runtime.Object
	kind string
}

func buildKustomizeResources(t *testing.T, kustomizeDir string) ([]kustomizeBuildObject, error) {
	fileSys := filesys.MakeFsOnDisk()
	if !fileSys.Exists(kustomizeDir) {
		return nil, fmt.Errorf("kustomization dir %s does not exist on file system", kustomizeDir)
	}

	k := krusty.MakeKustomizer(&krusty.Options{
		Reorder:           krusty.ReorderOptionUnspecified,
		AddManagedbyLabel: true,
		PluginConfig: &types.PluginConfig{
			FnpLoadingOptions: types.FnPluginLoadingOptions{},
		},
	})
	m, err := k.Run(fileSys, kustomizeDir)
	if err != nil {
		return nil, err
	}

	resources := []kustomizeBuildObject{}
	for _, resource := range m.Resources() {
		yaml, err := resource.AsYAML()
		if err != nil {
			return nil, fmt.Errorf("failed get kustomization output yaml: %+w", err)
		}
		obj, groupVersionKind, err := scheme.Codecs.UniversalDeserializer().Decode(yaml, nil, nil)
		if err != nil {
			return nil, fmt.Errorf("failed parse kustomization output yaml: %+w", err)
		}

		resources = append(resources, kustomizeBuildObject{
			obj:  obj,
			kind: groupVersionKind.String(),
		})
	}
	return resources, err
}
