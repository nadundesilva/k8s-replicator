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
	"fmt"

	"github.com/nadundesilva/k8s-replicator/pkg/kubernetes"
	"github.com/nadundesilva/k8s-replicator/pkg/replicator/resources"
	"go.uber.org/zap"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
)

type ResourceEventHandler struct {
	replicator resources.ResourceReplicator
	k8sClient  kubernetes.ClientInterface
	logger     *zap.SugaredLogger
}

func NewResourcesEventHandler(replicator resources.ResourceReplicator, k8sClient kubernetes.ClientInterface,
	logger *zap.SugaredLogger) *ResourceEventHandler {
	return &ResourceEventHandler{
		replicator: replicator,
		k8sClient:  k8sClient,
		logger: logger.With("replicatingResourceApiVersion", replicator.ResourceApiVersion(),
			"replicatingResourceName", replicator.ResourceName()),
	}
}

func (h *ResourceEventHandler) OnAdd(obj interface{}) {
	newObj := obj.(metav1.Object)
	logger := h.logger.With("sourceNamespace", newObj.GetNamespace(), "name", newObj.GetName())
	err := h.handleUpdate(newObj, logger)
	if err != nil {
		logger.Errorw("failed to handle replicating added object", "error", err)
	} else {
		logger.Infow("completed replicating added object")
	}
}

func (h *ResourceEventHandler) OnUpdate(oldObj, newObj interface{}) {
	updatedObj := newObj.(metav1.Object)
	logger := h.logger.With("sourceNamespace", updatedObj.GetNamespace(), "name", updatedObj.GetName())
	err := h.handleUpdate(updatedObj, logger)
	if err != nil {
		logger.Errorw("failed to handle replicating updated object", "error", err)
	} else {
		logger.Infow("completed replicating updated object")
	}
}

func (h *ResourceEventHandler) OnDelete(obj interface{}) {
	ctx := context.Background()
	deletedObj := obj.(metav1.Object)
	logger := h.logger.With("sourceNamespace", deletedObj.GetNamespace(), "name", deletedObj.GetName())

	namespaces, err := h.k8sClient.ListNamespaces(labels.Everything())
	if err != nil {
		logger.Errorw("failed to handle deleting object: failed to list namespaces", "error", err)
	} else {
		for _, namespace := range namespaces {
			if namespace.GetName() != deletedObj.GetNamespace() {
				logger := logger.With("targerNamespace", namespace.GetName())
				err := h.k8sClient.DeleteSecret(ctx, namespace.GetName(), deletedObj.GetName())
				if err != nil {
					logger.Errorw("failed to delete secret", "error", err)
				} else {
					logger.Debugw("deleted object from namespace")
				}
			}
		}
		logger.Infow("completed deleting object")
	}
}

func (h *ResourceEventHandler) handleUpdate(currentObj metav1.Object, logger *zap.SugaredLogger) error {
	ctx := context.Background()
	clonedObj := h.cloneObject(h.replicator, currentObj)

	namespaces, err := h.k8sClient.ListNamespaces(labels.Everything())
	if err != nil {
		return fmt.Errorf("failed to list namespaces %+w", err)
	} else {
		for _, namespace := range namespaces {
			if namespace.GetName() != currentObj.GetNamespace() {
				_, err := h.replicator.Get(ctx, namespace.GetName(), currentObj.GetName())
				if err != nil {
					logger := logger.With("targetNamespace", namespace.GetName())
					if errors.IsNotFound(err) {
						err = h.replicator.Create(ctx, namespace.GetName(), clonedObj)
						if err != nil {
							logger.Errorw("failed to create new object", "error", err)
						} else {
							logger.Debugw("replicated object to namespace")
						}
					} else {
						logger.Errorw("failed to check if object exists", "error", err)
					}
				}
			}
		}
	}
	return nil
}

func (h *ResourceEventHandler) cloneObject(replicator resources.ResourceReplicator, source metav1.Object) metav1.Object {
	clonedObj := replicator.Clone(source)
	clonedObj.SetName(source.GetName())

	newLabels := map[string]string{}
	for k, v := range source.GetLabels() {
		newLabels[k] = v
	}
	newLabels[ReplicationObjectTypeLabelKey] = ReplicationObjectTypeLabelValueClone
	clonedObj.SetLabels(newLabels)

	newAnnotations := map[string]string{}
	for k, v := range source.GetAnnotations() {
		newAnnotations[k] = v
	}
	clonedObj.SetAnnotations(newAnnotations)

	return clonedObj
}
