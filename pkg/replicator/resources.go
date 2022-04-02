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

const (
	ObjectTypeLabelKey      = "replicator.nadundesilva.github.io/object-type"
	SourceNamespaceLabelKey = "replicator.nadundesilva.github.io/source-namespace"

	ObjectTypeLabelValueSource  = "source"
	ObjectTypeLabelValueReplica = "replica"
)

type ResourceEventHandler struct {
	replicator resources.ResourceReplicator
	k8sClient  kubernetes.ClientInterface
	logger     *zap.SugaredLogger
}

func NewResourcesEventHandler(replicator resources.ResourceReplicator, k8sClient kubernetes.ClientInterface,
	logger *zap.SugaredLogger) *ResourceEventHandler {
	logger = logger.With("apiVersion", replicator.ResourceApiVersion(), "kind", replicator.ResourceKind())

	return &ResourceEventHandler{
		replicator: replicator,
		k8sClient:  k8sClient,
		logger:     logger,
	}
}

func (h *ResourceEventHandler) OnAdd(newObj interface{}) {
	err := h.handleUpdate(newObj)
	if err != nil {
		h.logger.Errorw("failed to handle added object", "error", err)
	}
}

func (h *ResourceEventHandler) OnUpdate(oldObj, newObj interface{}) {
	err := h.handleUpdate(newObj)
	if err != nil {
		h.logger.Errorw("failed to handle updated object", "error", err)
	}
}

func (h *ResourceEventHandler) OnDelete(obj interface{}) {
	ctx := context.Background()
	deletedObj := obj.(metav1.Object)
	logger := h.logger.With("name", deletedObj.GetName())

	if isReplicationSource(deletedObj) {
		logger := logger.With("sourceNamespace", deletedObj.GetNamespace())

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
							logger.Errorw("failed to delete object", "error", err)
						} else {
							logger.Debugw("deleted object from namespace")
						}
					}
				}
			}
			logger.Infow("completed deleting source object replicas")
		}
	} else if isReplica(deletedObj) {
		if sourceNamespaceName, ok := deletedObj.GetLabels()[SourceNamespaceLabelKey]; ok {
			logger := logger.With("sourceNamespace", sourceNamespaceName)

			_, err := h.replicator.Get(sourceNamespaceName, deletedObj.GetName())
			if err != nil {
				if !errors.IsNotFound(err) {
					logger.Errorw("failed to get source object of replica", "error", err)
				}
			} else {
				logger := logger.With("replicaNamespace", deletedObj.GetNamespace(), "name", deletedObj.GetName())

				namespace, err := h.k8sClient.GetNamespace(deletedObj.GetNamespace())
				if err != nil {
					if !errors.IsNotFound(err) {
						logger.Errorw("failed to check if source namespace of replica exists", "error", err)
					}
				} else if isManagedNamespace(logger, namespace) {
					clonedObj := cloneObject(h.replicator, deletedObj)
					clonedObj.GetLabels()[SourceNamespaceLabelKey] = sourceNamespaceName
					err = h.replicator.Apply(ctx, namespace.GetName(), clonedObj)
					if err != nil {
						logger.Errorw("failed to recreate deleted replica", "error", err)
					} else {
						logger.Infow("recreated deleted replica")
					}
				}
			}
		} else {
			logger.Errorw("deleted replica does not contain label %s", SourceNamespaceLabelKey)
		}
	} else {
		logger.Errorw("ignored object's event received by replicator")
	}
}

func (h *ResourceEventHandler) handleUpdate(newObj interface{}) error {
	ctx := context.Background()
	object := newObj.(metav1.Object)
	logger := h.logger.With("name", object.GetName())

	if isReplicationSource(object) {
		logger := h.logger.With("sourceNamespace", object.GetNamespace())

		sourceNamespace, err := h.k8sClient.GetNamespace(object.GetNamespace())
		if err != nil {
			if errors.IsNotFound(err) {
				logger.Warnw("object marked as a replication source in an non-existent/ignored namespace")
			} else {
				return fmt.Errorf("failed to get source namespace: %w", err)
			}
		} else if isManagedNamespace(logger, sourceNamespace) {
			namespaces, err := h.k8sClient.ListNamespaces(labels.Everything())
			if err != nil {
				return fmt.Errorf("failed to list namespaces: %w", err)
			} else {
				clonedObj := cloneObject(h.replicator, object)
				for _, namespace := range namespaces {
					logger := logger.With("targetNamespace", namespace.GetName())
					replicationAttempted, err := createReplica(ctx, logger, object.GetNamespace(), namespace, clonedObj,
						h.replicator)
					if replicationAttempted {
						if err != nil {
							return fmt.Errorf("failed to create replica in namespace: %w", err)
						} else {
							logger.Debugw("created replica in namespace")
						}
					}
				}
			}
		} else {
			logger.Warnw("object marked as a replication source in an ignored namespace")
		}
	} else if isReplica(object) {
		logger = h.logger.With("targetNamespace", object.GetNamespace())
		if sourceNamespaceName, ok := object.GetAnnotations()[SourceNamespaceLabelKey]; ok {
			logger = h.logger.With("sourceNamespace", sourceNamespaceName)

			deletionRequired := false
			sourceNamespace, err := h.k8sClient.GetNamespace(sourceNamespaceName)
			if err != nil {
				if errors.IsNotFound(err) {
					deletionRequired = true
				} else {
					return fmt.Errorf("failed to get source namespace: %w", err)
				}
			} else if !isManagedNamespace(logger, sourceNamespace) {
				deletionRequired = true
			}

			if deletionRequired {
				deletionAttempted, err := deleteReplica(ctx, logger, object.GetNamespace(), object.GetName(), h.replicator)
				if deletionAttempted {
					if err != nil {
						return fmt.Errorf("failed to delete replica with no source: %w", err)
					} else {
						logger.Debugw("deleted replica with no source")
					}
				}
			}
		} else {
			logger.Errorw("replica does not contain label %s", SourceNamespaceLabelKey)
		}
	} else {
		logger.Errorw("ignored object's event received by replicator", "namespace", object.GetNamespace())
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
	newLabels[ObjectTypeLabelKey] = ObjectTypeLabelValueReplica
	newLabels[SourceNamespaceLabelKey] = source.GetNamespace()
	clonedObj.SetLabels(newLabels)

	newAnnotations := map[string]string{}
	for k, v := range source.GetAnnotations() {
		newAnnotations[k] = v
	}
	clonedObj.SetAnnotations(newAnnotations)

	return clonedObj
}

func createReplica(ctx context.Context, logger *zap.SugaredLogger, sourceNamespace string,
	targetNamespace *corev1.Namespace, obj metav1.Object, replicator resources.ResourceReplicator) (bool, error) {
	if sourceNamespace == targetNamespace.GetName() || !isManagedNamespace(logger, targetNamespace) {
		return false, nil
	}
	return true, replicator.Apply(ctx, targetNamespace.GetName(), obj)
}

func deleteReplica(ctx context.Context, logger *zap.SugaredLogger, namespace, name string,
	replicator resources.ResourceReplicator) (bool, error) {
	replica, err := replicator.Get(namespace, name)
	if err != nil {
		if errors.IsNotFound(err) {
			return false, nil
		} else {
			logger.Warnw("failed to check if object replica exists", "error", err)
		}
	}
	if !isReplica(replica) {
		return false, nil
	}
	return true, replicator.Delete(ctx, namespace, name)
}

func isReplicationSource(obj metav1.Object) bool {
	val, ok := obj.GetLabels()[ObjectTypeLabelKey]
	return ok && val == ObjectTypeLabelValueSource
}

func isReplica(obj metav1.Object) bool {
	val, ok := obj.GetLabels()[ObjectTypeLabelKey]
	return ok && val == ObjectTypeLabelValueReplica
}
