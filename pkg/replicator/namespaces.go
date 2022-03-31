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
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
)

const (
	ReplicationTargetNamespaceTypeLabelKey = "replicator.nadundesilva.github.io/target-namespace"

	ReplicationTargetNamespaceTypeLabelValueReplicated = "replicated"
	ReplicationTargetNamespaceTypeLabelValueIgnored    = "ignored"
)

var (
	controllerNamespace = os.Getenv("CONTROLLER_NAMESPACE")
)

func (r *controller) handleNewNamespace(obj interface{}) {
	ctx := context.Background()
	namespace := obj.(*corev1.Namespace)
	logger := r.logger.With("targetNamespace", namespace.GetName())

	sourceSelectorRequirement, err := labels.NewRequirement(
		ReplicationObjectTypeLabelKey,
		selection.Equals,
		[]string{ReplicationObjectTypeLabelValueSource},
	)
	if err != nil {
		logger.Errorw("failed to initialize source objects filter", "error", err)
	}

	for _, replicator := range r.resourceReplicators {
		logger := logger.With("apiVersion", replicator.ResourceApiVersion(), "resource", replicator.ResourceName())
		objects, err := replicator.List("", labels.NewSelector().Add(*sourceSelectorRequirement))
		if err != nil {
			logger.Errorw("failed to list the resources")
		}
		for _, object := range objects {
			clonedObj := cloneObject(replicator, object)

			err = replicateToNamespace(ctx, logger, object.GetNamespace(), namespace, clonedObj,
				replicator)
			if err != nil {
				logger.Errorw("failed to replicate object to new namespace", "error", err)
			} else {
				logger.Infow("replicated object to new namespace", "object", object.GetName())
			}
		}
	}
}

func (r *controller) handleUpdateNamespace(prevObj, newObj interface{}) {
	ctx := context.Background()
	prevNamespace := prevObj.(*corev1.Namespace)
	newNamespace := newObj.(*corev1.Namespace)
	logger := r.logger.With("targetNamespace", newNamespace.GetName())

	clonesSelectorRequirement, err := labels.NewRequirement(
		ReplicationObjectTypeLabelKey,
		selection.Equals,
		[]string{ReplicationObjectTypeLabelValueClone},
	)
	if err != nil {
		logger.Errorw("failed to initialize cloned objects filter", "error", err)
	}

	if !isReplicationTargetNamespace(logger, prevNamespace) && isReplicationTargetNamespace(logger, newNamespace) {
		r.handleNewNamespace(newObj)
	} else if isReplicationTargetNamespace(logger, prevNamespace) && !isReplicationTargetNamespace(logger, newNamespace) {
		for _, replicator := range r.resourceReplicators {
			logger := logger.With("apiVersion", replicator.ResourceApiVersion(), "resource", replicator.ResourceName())
			objects, err := replicator.List("", labels.NewSelector().Add(*clonesSelectorRequirement))
			if err != nil {
				logger.Errorw("failed to list the resources")
			}
			for _, object := range objects {
				err = replicator.Delete(ctx, newNamespace.GetName(), object.GetName())
				if err != nil {
					logger.Errorw("failed to delete object from namespace", "error", err)
				} else {
					logger.Infow("deleted object from namespace", "object", object.GetName())
				}
			}
		}
	}
}

func isReplicationTargetNamespace(logger *zap.SugaredLogger, namespace *corev1.Namespace) bool {
	val, ok := namespace.GetLabels()[ReplicationTargetNamespaceTypeLabelKey]
	if ok {
		if val == ReplicationTargetNamespaceTypeLabelValueReplicated {
			return true
		} else if val == ReplicationTargetNamespaceTypeLabelValueIgnored {
			return false
		} else {
			logger.Warnw("ignored unrecorgnized label in target namespace",
				"labelKey", ReplicationTargetNamespaceTypeLabelKey, "labelValue", val)
		}
	}
	return !strings.HasPrefix(namespace.GetName(), "kube-") && namespace.GetName() != controllerNamespace
}
