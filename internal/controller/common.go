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
	"context"
	"fmt"
	"strings"

	"github.com/go-logr/logr"
	"github.com/nadundesilva/k8s-replicator/internal/controller/replication"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/record"
	"k8s.io/klog/v2"
	"k8s.io/utils/ptr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	ctrlController "sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func newManagerOptions(mgr ctrl.Manager, name string, kind string) ctrlController.Options {
	logger := mgr.GetLogger().WithValues(
		"controller", name,
	)
	return ctrlController.Options{
		MaxConcurrentReconciles: 100,
		RecoverPanic:            ptr.To(true),
		NeedLeaderElection:      ptr.To(true),
		LogConstructor: func(req *reconcile.Request) logr.Logger {
			logger := logger
			if req != nil {
				logger = logger.WithValues(
					"reconcileObject", klog.KRef(req.Namespace, req.Name),
					"reconcileKind", kind,
				)
			}
			return logger
		},
	}
}

func replicateObject(ctx context.Context, k8sClient client.Client, eventRecorder record.EventRecorder,
	ns string, sourceObject client.Object, replicator replication.Replicator) error {
	clonedObject := replicator.EmptyObject()
	clonedObject.SetNamespace(ns)
	clonedObject.SetName(sourceObject.GetName())

	result, err := ctrl.CreateOrUpdate(ctx, k8sClient, clonedObject, func() error {
		copyMap := func(sourceMap map[string]string, targetMap map[string]string) {
			for k, v := range sourceMap {
				if !strings.HasPrefix(k, groupFqn) {
					targetMap[k] = v
				}
			}
		}
		replicator.Replicate(sourceObject, clonedObject)

		labels := clonedObject.GetLabels()
		if labels == nil {
			labels = map[string]string{}
		}
		copyMap(sourceObject.GetLabels(), labels)
		labels[objectTypeLabelKey] = objectTypeLabelValueReplica
		clonedObject.SetLabels(labels)

		annotations := clonedObject.GetAnnotations()
		if annotations == nil {
			annotations = map[string]string{}
		}
		copyMap(sourceObject.GetAnnotations(), annotations)
		annotations[sourceNamespaceAnnotationKey] = sourceObject.GetNamespace()
		clonedObject.SetAnnotations(annotations)
		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to replicate resource to namespace %v: %+w", ns, err)
	}
	switch result {
	case controllerutil.OperationResultCreated:
		eventRecorder.Eventf(sourceObject, "Normal", SourceObjectCreate, "replica in namespace %s created", ns)
	case controllerutil.OperationResultUpdated:
		eventRecorder.Eventf(sourceObject, "Normal", SourceObjectUpdate, "replica in namespace %s updated", ns)
	}

	err = addFinalizer(ctx, k8sClient, clonedObject)
	if err != nil {
		return err
	}
	return nil
}

func deleteObject(ctx context.Context, k8sClient client.Client, object client.Object) error {
	err := removeFinalizer(ctx, k8sClient, object)
	if err != nil {
		return err
	}

	propagationStrategy := metav1.DeletePropagationBackground
	err = k8sClient.Delete(ctx, object, &client.DeleteOptions{
		PropagationPolicy: &propagationStrategy,
	})
	if err != nil {
		if errors.IsNotFound(err) {
			return nil
		} else {
			return err
		}
	}
	return nil
}

func addFinalizer(ctx context.Context, k8sClient client.Client, object client.Object) error {
	if !controllerutil.ContainsFinalizer(object, resourceFinalizer) {
		controllerutil.AddFinalizer(object, resourceFinalizer)
		err := k8sClient.Update(ctx, object)
		if err != nil {
			return fmt.Errorf("failed to add finalizer: %+w", err)
		}
	}
	return nil
}

func removeFinalizer(ctx context.Context, k8sClient client.Client, object client.Object) error {
	if controllerutil.ContainsFinalizer(object, resourceFinalizer) {
		controllerutil.RemoveFinalizer(object, resourceFinalizer)
		err := k8sClient.Update(ctx, object)
		if err != nil {
			return err
		}
	}
	return nil
}

type sourceStatus string

const (
	sourceStatusNotFound  sourceStatus = "NotFound"
	sourceStatusDeleted   sourceStatus = "Deleted"
	sourceStatusUnmarked  sourceStatus = "Unmarked"
	sourceStatusAvailable sourceStatus = "Available"
)

func getReplicaSourceStatus(ctx context.Context, k8sClient client.Client, replica client.Object,
	replicator replication.Replicator) (sourceStatus, error) {
	sourceNamespace, sourceNamespaceOk := replica.GetAnnotations()[sourceNamespaceAnnotationKey]
	if !sourceNamespaceOk {
		return "", fmt.Errorf("replica does not contain %s annotation", sourceNamespaceAnnotationKey)
	}

	sourceObject := replicator.EmptyObject()
	sourceSeretKey := client.ObjectKey{Namespace: sourceNamespace, Name: replica.GetName()}
	if err := k8sClient.Get(ctx, sourceSeretKey, sourceObject); err != nil {
		if errors.IsNotFound(err) {
			return sourceStatusNotFound, nil
		} else {
			return "", fmt.Errorf("failed to get source object: %+w", err)
		}
	}

	if sourceObject.GetDeletionTimestamp() != nil {
		return sourceStatusDeleted, nil
	}

	sourceObjectType, sourceObjectTypeOk := sourceObject.GetLabels()[objectTypeLabelKey]
	if sourceObjectTypeOk {
		if sourceObjectType != objectTypeLabelValueReplicated {
			return "", fmt.Errorf("unexpected object type %s in source", sourceObjectType)
		}
	} else {
		return sourceStatusUnmarked, nil
	}
	return sourceStatusAvailable, nil
}

func isNamespaceIgnored(ns *corev1.Namespace) bool {
	namespaceType, namespaceTypeOk := ns.GetLabels()[namespaceTypeLabelKey]
	if namespaceTypeOk {
		switch namespaceType {
		case namespaceTypeLabelValueIgnored:
			return true
		case namespaceTypeLabelValueManaged:
			return false
		}
	}
	return strings.HasPrefix(ns.GetName(), "kube-") || (operatorNamespace != "" && ns.GetName() == operatorNamespace)
}
