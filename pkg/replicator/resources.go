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
	"log"

	"github.com/nadundesilva/k8s-replicator/pkg/kubernetes"
	"github.com/nadundesilva/k8s-replicator/pkg/replicator/resources"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
)

const (
	ObjectTypeLabelKey = "replicator.nadundesilva.github.io/object-type"

	SourceNamespaceAnnotationKey       = "replicator.nadundesilva.github.io/source-namespace"
	SourceResourceVersionAnnotationKey = "replicator.nadundesilva.github.io/source-resource-version"

	ObjectTypeLabelValueSource  = "source"
	ObjectTypeLabelValueReplica = "replica"
)

var (
	sourceObjectsLabelSelector labels.Selector
	replicasLabelSelector      labels.Selector
)

func init() {
	sourceSelectorRequirement, err := labels.NewRequirement(
		ObjectTypeLabelKey,
		selection.Equals,
		[]string{ObjectTypeLabelValueSource},
	)
	if err != nil {
		log.Fatalf("failed to initialize source objects selector: %v", err)
	}
	sourceObjectsLabelSelector = labels.NewSelector().Add(*sourceSelectorRequirement)

	replicaSelectorRequirement, err := labels.NewRequirement(
		ObjectTypeLabelKey,
		selection.Equals,
		[]string{ObjectTypeLabelValueReplica},
	)
	if err != nil {
		log.Fatalf("failed to initialize replicas selector: %v", err)
	}
	replicasLabelSelector = labels.NewSelector().Add(*replicaSelectorRequirement)
}

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
		if sourceNamespaceName, ok := deletedObj.GetAnnotations()[SourceNamespaceAnnotationKey]; ok {
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
					clonedObj.GetAnnotations()[SourceNamespaceAnnotationKey] = sourceNamespaceName

					err = h.replicator.Apply(ctx, namespace.GetName(), clonedObj)
					if err != nil {
						logger.Errorw("failed to recreate deleted replica", "error", err)
					} else {
						logger.Infow("recreated deleted replica")
					}
				}
			}
		} else {
			logger.Errorw("deleted replica does not contain source namespace annotation", "annotation",
				SourceNamespaceAnnotationKey)
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
		logger := logger.With("sourceNamespace", object.GetNamespace())

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
					logger := logger.With("replicaNamespace", namespace.GetName())
					replicationAttempted, err := applyReplica(ctx, logger, object, namespace, clonedObj,
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
		logger = logger.With("replicaNamespace", object.GetNamespace())
		if sourceNamespaceName, ok := object.GetAnnotations()[SourceNamespaceAnnotationKey]; ok {
			logger = logger.With("sourceNamespace", sourceNamespaceName)

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
			logger.Errorw("replica does not contain source namespace annotation", "annotation",
				SourceNamespaceAnnotationKey)
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
	clonedObj.SetLabels(newLabels)

	newAnnotations := map[string]string{}
	for k, v := range source.GetAnnotations() {
		newAnnotations[k] = v
	}
	newAnnotations[SourceNamespaceAnnotationKey] = source.GetNamespace()
	newAnnotations[SourceResourceVersionAnnotationKey] = source.GetResourceVersion()
	clonedObj.SetAnnotations(newAnnotations)

	return clonedObj
}

func applyReplica(ctx context.Context, logger *zap.SugaredLogger, sourceObj metav1.Object,
	replicaNamespace *corev1.Namespace, obj metav1.Object, replicator resources.ResourceReplicator) (bool, error) {
	if sourceObj.GetNamespace() == replicaNamespace.GetName() || !isManagedNamespace(logger, replicaNamespace) {
		return false, nil
	}
	existingReplica, err := replicator.Get(replicaNamespace.GetName(), obj.GetName())
	if err != nil {
		if !errors.IsNotFound(err) {
			logger.Warnw("failed to check if object replica already exists", "error", err)
		}
	} else {
		if val, ok := existingReplica.GetAnnotations()[SourceResourceVersionAnnotationKey]; ok {
			if val == sourceObj.GetResourceVersion() {
				return false, nil
			}
		} else {
			logger.Errorw("replica does not contain source resource version annotation", "annotation",
				SourceResourceVersionAnnotationKey)
		}
	}
	return true, replicator.Apply(ctx, replicaNamespace.GetName(), obj)
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
	} else if replica.GetDeletionTimestamp() != nil || !isReplica(replica) {
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
