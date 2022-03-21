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

	"github.com/nadundesilva/k8s-replicator/pkg/kubernetes"
	"github.com/nadundesilva/k8s-replicator/pkg/replicator/resources"
	"go.uber.org/zap"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
)

const (
	ReplicatedResourceLabelKey   = "nadundesilva.github.io/copied-object"
	ReplicatedResourceLabelValue = "true"
)

type ResourceEventHandler struct {
	replicator resources.ResourceReplicator
	k8sClient  kubernetes.ClientInterface
	logger     *zap.SugaredLogger
}

func NewResourcesEventHandler(replicator resources.ResourceReplicator, k8sClient kubernetes.ClientInterface, logger *zap.SugaredLogger) *ResourceEventHandler {
	return &ResourceEventHandler{
		replicator: replicator,
		k8sClient:  k8sClient,
		logger:     logger,
	}
}

func (h *ResourceEventHandler) OnAdd(obj interface{}) {
	ctx := context.Background()
	currentObj := obj.(metav1.Object)
	clonedObj := cloneObject(h.replicator, currentObj)

	namespaces, err := h.k8sClient.ListNamespaces(labels.Everything())
	if err != nil {
		h.logger.Errorw("failed to list namespace", "error", err)
	} else {
		for _, namespace := range namespaces {
			if namespace.GetName() != currentObj.GetName() {
				_, err := h.replicator.Get(ctx, namespace.GetName(), currentObj.GetName())
				if err != nil {
					if errors.IsNotFound(err) {
						h.replicator.Create(ctx, namespace.GetName(), clonedObj)
					} else {
						h.logger.Errorw("failed to check if resource exists", "error", err)
					}
				}
			}
		}
	}
}

func (h *ResourceEventHandler) OnUpdate(oldObj, newObj interface{}) {
	h.OnAdd(newObj)
}

func (h *ResourceEventHandler) OnDelete(obj interface{}) {
	// Handle Delete
}

func cloneObject(replicator resources.ResourceReplicator, source metav1.Object) metav1.Object {
	clonedObj := replicator.Clone(source)
	clonedObj.SetName(source.GetName())

	newLabels := map[string]string{}
	for k, v := range source.GetLabels() {
		newLabels[k] = v
	}
	newLabels[ReplicatedResourceLabelKey] = ReplicatedResourceLabelValue
	clonedObj.SetLabels(newLabels)

	newAnnotations := map[string]string{}
	for k, v := range source.GetAnnotations() {
		newAnnotations[k] = v
	}
	clonedObj.SetAnnotations(newAnnotations)

	return clonedObj
}
