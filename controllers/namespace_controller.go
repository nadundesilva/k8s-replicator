/*
 * Copyright (c) 2023, Nadun De Silva. All Rights Reserved.
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

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// SecretReconciler reconciles a Namespace object
type NamespaceReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups="",resources=namespaces,verbs=get;list;watch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the namespace object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.13.0/pkg/reconcile
func (r *NamespaceReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	ctx = log.IntoContext(ctx, log.FromContext(ctx).WithValues("namespace", req.NamespacedName.Name))

	// Fetching object
	isNamespaceDeleted := false
	namespace := &corev1.Namespace{}
	if err := r.Get(ctx, req.NamespacedName, namespace); err != nil {
		if errors.IsNotFound(err) {
			isNamespaceDeleted = true
		} else {
			return ctrl.Result{}, fmt.Errorf("failed to get namespace being reconciled: %+w", err)
		}
	}
	namespaceName := req.NamespacedName.Name

	isNamespaceDeleted = isNamespaceDeleted || namespace.GetDeletionTimestamp() != nil
	isNamespaceIgnored := isNamespaceIgnored(namespace)

	// Reconciling
	if isNamespaceDeleted || isNamespaceIgnored {
		replicaSecrets := &corev1.SecretList{}
		err := r.List(ctx, replicaSecrets, &client.ListOptions{
			Namespace:     namespaceName,
			LabelSelector: replicaResourcesSelector,
		})
		if err != nil {
			return ctrl.Result{}, err
		}

		errs := []error{}
		for _, secret := range replicaSecrets.Items {
			if isNamespaceDeleted {
				log.FromContext(ctx).Info("Removing finalizer from replica in deleted namespace")
				err := removeFinalizer(ctx, r.Client, &secret)
				if err != nil {
					errs = append(errs, fmt.Errorf("failed to delete secret: %+w", err))
				}
			}
			if isNamespaceIgnored {
				log.FromContext(ctx).Info("Deleting replica in ignored namespace")
				err := deleteObject(ctx, r.Client, &secret)
				if err != nil {
					errs = append(errs, fmt.Errorf("failed to delete secret: %+w", err))
				}
			}
		}
		if len(errs) > 0 {
			return ctrl.Result{}, fmt.Errorf("failed to iterate replicated secrets in removed namespace: %+v", errs)
		}
	} else {
		replicatedSecrets := &corev1.SecretList{}
		err := r.List(ctx, replicatedSecrets)
		if err != nil {
			return ctrl.Result{}, err
		}

		errs := []error{}
		for _, secret := range replicatedSecrets.Items {
			if secret.GetDeletionTimestamp() != nil { // Secret already
				continue
			}
			if secret.GetNamespace() == namespaceName { // Replicated resource is from current namespace
				continue
			}
			if objectType, objectTypeOk := secret.GetLabels()[ObjectTypeLabelKey]; objectTypeOk && objectType == ObjectTypeLabelValueReplicated {
				log.FromContext(ctx).Info("Creating/Updating replica")
				err := replicateObject(ctx, r.Client, namespaceName, &secret)
				if err != nil {
					errs = append(errs, err)
				}
			}
		}
		if len(errs) > 0 {
			return ctrl.Result{}, fmt.Errorf("failed to iterate replicated secrets in namespace: %+v", errs)
		}
	}
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *NamespaceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1.Namespace{}).
		Complete(r)
}
