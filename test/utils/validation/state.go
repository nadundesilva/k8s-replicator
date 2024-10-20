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
package validation

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/e2e-framework/klient/k8s"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
)

type clusterObject struct {
	Name              string            `json:"name,omitempty"`
	Labels            map[string]string `json:"labels,omitempty"`
	Annotations       map[string]string `json:"annotations,omitempty"`
	Finalizers        []string          `json:"finalizers,omitempty"`
	DeletionTimestamp *metav1.Time      `json:"deletionTimestamp,omitempty"`
}

type namespacedObject struct {
	clusterObject
	Kind      schema.GroupVersionKind `json:"kind,omitempty"`
	Namespace string                  `json:"namespace,omitempty"`
}

func printState(ctx context.Context, t *testing.T, cfg *envconf.Config, sourceObject k8s.Object) error {
	namespaceList := &corev1.NamespaceList{}
	err := cfg.Client().Resources().List(ctx, namespaceList)
	if err != nil {
		return fmt.Errorf("failed to get the list of namespace: %w", err)
	}

	namespaceData := []*clusterObject{}
	objectData := []*namespacedObject{}
	for _, namespace := range namespaceList.Items {
		namespaceData = append(namespaceData, &clusterObject{
			Name:              namespace.GetName(),
			Labels:            namespace.GetLabels(),
			Annotations:       namespace.GetAnnotations(),
			Finalizers:        namespace.GetFinalizers(),
			DeletionTimestamp: namespace.GetDeletionTimestamp(),
		})

		clonedObj := sourceObject.DeepCopyObject().(k8s.Object)
		err := cfg.Client().Resources(namespace.GetName()).Get(ctx, clonedObj.GetName(), namespace.GetName(), clonedObj)
		if err != nil {
			t.Logf("failed to get object %s in namespace %s: %v", clonedObj.GetName(), namespace.GetName(), err)
		} else {
			objectData = append(objectData, &namespacedObject{
				Kind:      namespace.GetObjectKind().GroupVersionKind(),
				Namespace: namespace.GetName(),
				clusterObject: clusterObject{
					Name:              clonedObj.GetName(),
					Labels:            clonedObj.GetLabels(),
					Annotations:       clonedObj.GetAnnotations(),
					Finalizers:        clonedObj.GetFinalizers(),
					DeletionTimestamp: clonedObj.GetDeletionTimestamp(),
				},
			})
		}
	}

	formattedJSON, err := json.MarshalIndent(namespaceData, "", "\t")
	if err != nil {
		return fmt.Errorf("failed to format namespace data into json: %w", err)
	}
	t.Logf("namespaces: %s", formattedJSON)

	formattedJSON, err = json.MarshalIndent(objectData, "", "\t")
	if err != nil {
		return fmt.Errorf("failed to format object data into json: %w", err)
	}
	t.Logf("objects: %s", formattedJSON)

	return nil
}
