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
package replicator

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
)

func (r *controller) handleNewNamespace(obj interface{}) {
	namespace := obj.(*corev1.Namespace)
	logger := r.logger.With("namespace", namespace.GetName())

	selectorRequirement, err := labels.NewRequirement(
		ReplicationObjectTypeLabelKey,
		selection.In,
		[]string{ReplicationObjectTypeLabelValueSource},
	)
	if err != nil {
		logger.Errorw("failed to initialize resources filter", "error", err)
	}

	for _, replicator := range r.resourceReplicators {
		logger := logger.With("apiVersion", replicator.ResourceApiVersion(), "resource", replicator.ResourceName())
		objects, err := replicator.List("", labels.NewSelector().Add(*selectorRequirement))
		if err != nil {
			logger.Errorw("failed to list the resources")
		}
		for _, object := range objects {
			logger := logger.With("targetNamespace", namespace.GetName())
			clonedObj := cloneObject(replicator, object)

			err = replicateToNamespace(context.Background(), object.GetNamespace(), namespace.GetName(), clonedObj,
				replicator)
			if err != nil {
				logger.Errorw("failed to replicate object to new namespace", "error", err)
			} else {
				logger.Infow("replicated object to new namespace", "object", object.GetName())
			}
		}
	}
}
