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
	"os"
	"strings"

	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
)

const (
	NamespaceTypeLabelKey = "replicator.nadundesilva.github.io/namespace-type"

	NamespaceTypeLabelValueManaged = "managed"
	NamespaceTypeLabelValueIgnored = "ignored"
)

var (
	controllerNamespace = os.Getenv("CONTROLLER_NAMESPACE")
)

func (r *controller) handleNewNamespace(obj interface{}) {
	ctx := context.Background()
	namespace := obj.(*corev1.Namespace)
	logger := r.logger.With("replicaNamespace", namespace.GetName())

	if !isManagedNamespace(logger, namespace) {
		return
	}

	for _, replicator := range r.resourceReplicators {
		logger := logger.With("apiVersion", replicator.ResourceApiVersion(), "kind", replicator.ResourceKind())

		objects, err := replicator.List("", sourceObjectsLabelSelector)
		if err != nil {
			logger.Errorw("failed to list the resources")
		}
		for _, object := range objects {
			clonedObj := cloneObject(replicator, object)

			replicationAttempted, err := createReplica(ctx, logger, object.GetNamespace(), namespace, clonedObj,
				replicator)
			if replicationAttempted {
				if err != nil {
					logger.Errorw("failed to replicate object to new namespace", "error", err)
				} else {
					logger.Infow("replicated object to new namespace", "object", object.GetName())
				}
			}
		}
	}
}

func (r *controller) handleUpdateNamespace(prevObj, newObj interface{}) {
	prevNamespace := prevObj.(*corev1.Namespace)
	newNamespace := newObj.(*corev1.Namespace)

	logger := r.logger.With("namespace", newNamespace.GetName())
	if !isManagedNamespace(logger, prevNamespace) && isManagedNamespace(logger, newNamespace) {
		r.handleNewNamespace(newObj)
	} else if isManagedNamespace(logger, prevNamespace) && !isManagedNamespace(logger, newNamespace) {
		r.handleDeleteNamespace(newObj)
	}
}

func (r *controller) handleDeleteNamespace(obj interface{}) {
	ctx := context.Background()
	deletedNamespace := obj.(*corev1.Namespace)

	if deletedNamespace.GetDeletionTimestamp() != nil {
		return
	}
	// Namespaces which are removed only due to being marked as ignored needs to be cleaned up

	logger := r.logger.With("replicaNamespace", deletedNamespace.GetName())
	for _, replicator := range r.resourceReplicators {
		logger := logger.With("apiVersion", replicator.ResourceApiVersion(), "kind", replicator.ResourceKind())

		objects, err := replicator.List(deletedNamespace.GetName(), replicasLabelSelector)
		if err != nil {
			logger.Errorw("failed to list the resources")
		}
		for _, object := range objects {
			deletionAttempted, err := deleteReplica(ctx, logger, deletedNamespace.GetName(), object.GetName(), replicator)
			if deletionAttempted {
				if err != nil {
					logger.Errorw("failed to delete object from namespace", "error", err)
				} else {
					logger.Infow("deleted object from namespace", "object", object.GetName())
				}
			}
		}
	}
}

func isManagedNamespace(logger *zap.SugaredLogger, namespace *corev1.Namespace) bool {
	if namespace == nil || namespace.GetDeletionTimestamp() != nil {
		return false
	}
	val, ok := namespace.GetLabels()[NamespaceTypeLabelKey]
	if ok {
		if val == NamespaceTypeLabelValueManaged {
			return true
		} else if val == NamespaceTypeLabelValueIgnored {
			return false
		} else {
			logger.Warnw("ignored unrecorgnized label in replica's namespace",
				"labelKey", NamespaceTypeLabelKey, "labelValue", val)
		}
	}
	return !strings.HasPrefix(namespace.GetName(), "kube-") && namespace.GetName() != controllerNamespace
}
