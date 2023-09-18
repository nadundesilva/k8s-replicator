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
package controllers

import (
	"context"
	"fmt"
	"strings"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func replicateObject(ctx context.Context, k8sClient client.Client, ns string, sourceSecret *corev1.Secret) error {
	clonedSecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      sourceSecret.GetName(),
			Namespace: ns,
		},
	}
	_, err := ctrl.CreateOrUpdate(ctx, k8sClient, clonedSecret, func() error {
		addToMap := func(sourceMap map[string]string, targetMap map[string]string) {
			for k, v := range sourceMap {
				if !strings.HasPrefix(k, groupFqn) {
					targetMap[k] = v
				}
			}
		}

		labels := clonedSecret.ObjectMeta.GetLabels()
		if labels == nil {
			labels = map[string]string{}
		}
		addToMap(sourceSecret.GetLabels(), labels)
		labels[ObjectTypeLabelKey] = ObjectTypeLabelValueReplica
		clonedSecret.ObjectMeta.SetLabels(labels)

		annotations := clonedSecret.ObjectMeta.GetAnnotations()
		if annotations == nil {
			annotations = map[string]string{}
		}
		addToMap(sourceSecret.GetAnnotations(), annotations)
		annotations[SourceNamespaceAnnotationKey] = sourceSecret.GetNamespace()
		clonedSecret.SetAnnotations(annotations)

		clonedSecret.Immutable = sourceSecret.Immutable
		clonedSecret.Data = sourceSecret.Data
		clonedSecret.StringData = sourceSecret.StringData
		clonedSecret.Type = sourceSecret.Type
		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to replicate resource to namespace %v: %+w", ns, err)
	}

	addFinalizer(ctx, k8sClient, clonedSecret)
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

func getReplicaSourceStatus(ctx context.Context, k8sClient client.Client, replica client.Object) (sourceStatus, error) {
	sourceNamespace, sourceNamespaceOk := replica.GetAnnotations()[SourceNamespaceAnnotationKey]
	if !sourceNamespaceOk {
		return "", fmt.Errorf("replica does not contain %s annotation", SourceNamespaceAnnotationKey)
	}

	sourceSecret := &corev1.Secret{}
	sourceSeretKey := client.ObjectKey{Namespace: sourceNamespace, Name: replica.GetName()}
	if err := k8sClient.Get(ctx, sourceSeretKey, sourceSecret); err != nil {
		if errors.IsNotFound(err) {
			return sourceStatusNotFound, nil
		} else {
			return "", fmt.Errorf("failed to get source secret: %+w", err)
		}
	}

	if sourceSecret.GetDeletionTimestamp() != nil {
		return sourceStatusDeleted, nil
	}

	sourceObjectType, sourceObjectTypeOk := sourceSecret.GetLabels()[ObjectTypeLabelKey]
	if sourceObjectTypeOk {
		if sourceObjectType != ObjectTypeLabelValueReplicated {
			return "", fmt.Errorf("Unexpected object type %s in source", sourceObjectType)
		}
	} else {
		return sourceStatusUnmarked, nil
	}
	return sourceStatusAvailable, nil
}

func isNamespaceIgnored(ns *corev1.Namespace) bool {
	namespaceType, namespaceTypeOk := ns.GetLabels()[NamespaceTypeLabelKey]
	if namespaceTypeOk {
		if namespaceType == NamespaceTypeLabelValueIgnored {
			return true
		} else if namespaceType == NamespaceTypeLabelValueManaged {
			return false
		}
	}
	return strings.HasPrefix(ns.GetName(), "kube-") || (operatorNamespace != "" && ns.GetName() == operatorNamespace)
}
