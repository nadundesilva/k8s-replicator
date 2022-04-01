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
	corev1 "k8s.io/api/core/v1"
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
	logger = logger.With("apiVersion", replicator.ResourceApiVersion(), "resource", replicator.ResourceName())
	return &ResourceEventHandler{
		replicator: replicator,
		k8sClient:  k8sClient,
		logger:     logger,
	}
}

func (h *ResourceEventHandler) OnAdd(obj interface{}) {
	newObj := obj.(metav1.Object)
	if !isReplicationSource(newObj) {
		return
	}

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
	if !isReplicationSource(updatedObj) {
		return
	}

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
	if isReplicationSource(deletedObj) {
		logger := h.logger.With("sourceNamespace", deletedObj.GetNamespace(), "name", deletedObj.GetName())
		namespaces, err := h.k8sClient.ListNamespaces(labels.Everything())
		if err != nil {
			logger.Errorw("failed to handle deleting object: failed to list namespaces", "error", err)
		} else {
			for _, namespace := range namespaces {
				if namespace.GetName() != deletedObj.GetNamespace() {
					logger := logger.With("targerNamespace", namespace.GetName())
					deletionAttempted, err := deleteReplica(ctx, logger, namespace.GetName(), deletedObj.GetName(), h.replicator)
					if deletionAttempted {
						if err != nil && !errors.IsNotFound(err) {
							logger.Errorw("failed to delete secret", "error", err)
						} else {
							logger.Debugw("deleted object from namespace")
						}
					}
				}
			}
			logger.Infow("completed deleting object")
		}
	} else if isReplicationClone(deletedObj) {
		if sourceNamespaceName, ok := deletedObj.GetLabels()[ReplicationSourceNamespaceLabelKey]; ok {
			_, err := h.replicator.Get(sourceNamespaceName, deletedObj.GetName())
			if err != nil {
				if !errors.IsNotFound(err) {
					h.logger.Errorw("failed to get source secret", "error", err, "sourceNamespace", sourceNamespaceName)
				}
			} else {
				logger := h.logger.With("cloneNamespace", deletedObj.GetNamespace(), "name", deletedObj.GetName())
				namespace, err := h.k8sClient.GetNamespace(deletedObj.GetNamespace())
				if err != nil {
					if !errors.IsNotFound(err) {
						logger.Errorw("failed to recreate deleted clone: failed to check namespace state", "error", err)
					}
				} else if namespace != nil && isReplicationTargetNamespace(logger, namespace) && namespace.GetDeletionTimestamp() == nil {
					clonedObj := cloneObject(h.replicator, deletedObj)
					clonedObj.GetLabels()[ReplicationSourceNamespaceLabelKey] = sourceNamespaceName
					err = h.replicator.Apply(ctx, namespace.GetName(), clonedObj)
					if err != nil {
						logger.Errorw("failed to recreate deleted clone", "error", err)
					} else {
						logger.Infow("recreated deleted clone")
					}
				}
			}
		}
	}
}

func (h *ResourceEventHandler) handleUpdate(currentObj metav1.Object, logger *zap.SugaredLogger) error {
	ctx := context.Background()
	clonedObj := cloneObject(h.replicator, currentObj)

	namespaces, err := h.k8sClient.ListNamespaces(labels.Everything())
	if err != nil {
		return fmt.Errorf("failed to list namespaces %+w", err)
	} else {
		for _, namespace := range namespaces {
			logger := logger.With("targetNamespace", namespace.GetName())
			replicationAttempted, err := replicateToNamespace(ctx, logger, currentObj.GetNamespace(), namespace, clonedObj,
				h.replicator)
			if replicationAttempted {
				if err != nil {
					logger.Errorw("failed to replicate object to namespace", "error", err)
				} else {
					logger.Debugw("replicated object to namespace")
				}
			}
		}
	}
	return nil
}

func cloneObject(replicator resources.ResourceReplicator, source metav1.Object) metav1.Object {
	clonedObj := replicator.Clone(source)
	clonedObj.SetName(source.GetName())

	newLabels := map[string]string{}
	for k, v := range source.GetLabels() {
		newLabels[k] = v
	}
	newLabels[ReplicationObjectTypeLabelKey] = ReplicationObjectTypeLabelValueClone
	newLabels[ReplicationSourceNamespaceLabelKey] = source.GetNamespace()
	clonedObj.SetLabels(newLabels)

	newAnnotations := map[string]string{}
	for k, v := range source.GetAnnotations() {
		newAnnotations[k] = v
	}
	clonedObj.SetAnnotations(newAnnotations)

	return clonedObj
}

func replicateToNamespace(ctx context.Context, logger *zap.SugaredLogger, sourceNamespace string,
	targetNamespace *corev1.Namespace, obj metav1.Object, replicator resources.ResourceReplicator) (bool, error) {
	if sourceNamespace == targetNamespace.GetName() || !isReplicationTargetNamespace(logger, targetNamespace) {
		return false, nil
	}
	return true, replicator.Apply(ctx, targetNamespace.GetName(), obj)
}

func deleteReplica(ctx context.Context, logger *zap.SugaredLogger, namespace, name string,
	replicator resources.ResourceReplicator) (bool, error) {
	_, err := replicator.Get(namespace, name)
	if err != nil {
		if errors.IsNotFound(err) {
			return false, nil
		} else {
			logger.Warnw("failed to check if object replica exists", "error", err)
		}
	}
	return true, replicator.Delete(ctx, namespace, name)
}

func isReplicationSource(obj metav1.Object) bool {
	val, ok := obj.GetLabels()[ReplicationObjectTypeLabelKey]
	return ok && val == ReplicationObjectTypeLabelValueSource
}

func isReplicationClone(obj metav1.Object) bool {
	val, ok := obj.GetLabels()[ReplicationObjectTypeLabelKey]
	return ok && val == ReplicationObjectTypeLabelValueClone
}
