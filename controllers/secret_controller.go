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

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// SecretReconciler reconciles a Secret object
type SecretReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="",resources=secrets/status,verbs=get;update;patch
//+kubebuilder:rbac:groups="",resources=secrets/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Secret object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.13.0/pkg/reconcile
func (r *SecretReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	// Fetching object
	isSecretDeleted := false
	secret := &corev1.Secret{}
	if err := r.Get(ctx, req.NamespacedName, secret); err != nil {
		if errors.IsNotFound(err) {
			isSecretDeleted = true
		} else {
			return ctrl.Result{}, fmt.Errorf("failed to get secret being reconciled: %+w", err)
		}
	}
	isSecretDeleted = isSecretDeleted || secret.GetDeletionTimestamp() != nil

	// Identifying object type
	objectType, objectTypeOk := secret.GetLabels()[ObjectTypeLabelKey]
	if !objectTypeOk {
		logger := log.FromContext(ctx).WithValues("reason", "object type not present in object")
		if controllerutil.ContainsFinalizer(secret, resourceFinalizer) {
			logger.Info("Removing replicas of unmarked object")
			return ctrl.Result{}, r.handleSourceRemoval(ctx, secret)
		} else {
			logger.Info("Ignoring unmarked object")
			return ctrl.Result{}, nil
		}
	}

	// Reconciling
	switch objectType {
	case ObjectTypeLabelValueReplica:
		ctx = log.IntoContext(ctx, log.FromContext(ctx).WithValues("replicaNamespace", secret.GetNamespace()))
		sourceStatus, err := getReplicaSourceStatus(ctx, r.Client, secret)
		if err != nil {
			return ctrl.Result{}, err
		}
		if sourceStatus != sourceStatusAvailable {
			logger := log.FromContext(ctx).WithValues("reason", "source object not available",
				"sourceStatus", sourceStatus)
			if isSecretDeleted {
				logger.Info("Removing finalizer from replica")
				return ctrl.Result{}, removeFinalizer(ctx, r.Client, secret)
			} else {
				logger.Info("Deleting replica")
				return ctrl.Result{}, deleteObject(ctx, r.Client, secret)
			}
		}
		return ctrl.Result{}, nil
	case ObjectTypeLabelValueReplicated:
		ctx = log.IntoContext(ctx, log.FromContext(ctx).WithValues("sourceNamespace", secret.GetNamespace()))
		if isSecretDeleted {
			return ctrl.Result{}, r.handleSourceRemoval(ctx, secret)
		} else {
			return ctrl.Result{}, r.handleSourceUpdate(ctx, secret)
		}
	default:
		logger := log.FromContext(ctx).WithValues("objectType", objectType)
		if controllerutil.ContainsFinalizer(secret, resourceFinalizer) {
			logger.Info("Removing any replicas of unknown object type if present")
			return ctrl.Result{}, r.handleSourceRemoval(ctx, secret)
		} else {
			logger.Info("Ignoring unknown object type")
			return ctrl.Result{}, nil
		}
	}
}

func (r *SecretReconciler) handleSourceRemoval(ctx context.Context, secret *corev1.Secret) error {
	err := r.iterateNamespaces(ctx, func(ns corev1.Namespace) error {
		if ns.GetName() == secret.GetNamespace() {
			return nil
		}

		replica := &corev1.Secret{}
		err := r.Get(ctx, client.ObjectKey{Namespace: ns.GetName(), Name: secret.GetName()}, replica)
		if err != nil {
			if errors.IsNotFound(err) {
				return nil
			} else {
				return err
			}
		}

		log.FromContext(ctx).Info("Deleting replica", "replicaNamespace", ns.GetName(),
			"reason", "source object deleted")
		return deleteObject(ctx, r.Client, replica)
	})
	if err != nil {
		return fmt.Errorf("failed to finalize source secret: %+w", err)
	}
	return removeFinalizer(ctx, r.Client, secret)
}

func (r *SecretReconciler) handleSourceUpdate(ctx context.Context, secret *corev1.Secret) error {
	err := addFinalizer(ctx, r.Client, secret)
	if err != nil {
		return err
	}
	err = r.iterateNamespaces(ctx, func(ns corev1.Namespace) error {
		if ns.GetName() == secret.GetNamespace() {
			return nil
		}

		log.FromContext(ctx).Info("Creating/Updating replica", "replicaNamespace", ns.GetName())
		return replicateObject(ctx, r.Client, ns.GetName(), secret)
	})
	return err
}

func (r *SecretReconciler) iterateNamespaces(ctx context.Context, handler func(ns corev1.Namespace) error) error {
	namespaceList := &corev1.NamespaceList{}
	err := r.List(ctx, namespaceList, &client.ListOptions{
		LabelSelector: namespaceSelector,
	})
	if err != nil {
		return err
	}

	errs := []error{}
	for _, ns := range namespaceList.Items {
		if isNamespaceIgnored(&ns) {
			continue
		}

		err = handler(ns)
		if err != nil {
			errs = append(errs, fmt.Errorf("failed to run reconciliation for namespace %s: %+w", ns.GetName(), err))
			continue
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("failed to iterate namespaces: %+v", errs)
	}
	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *SecretReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1.Secret{}, builder.WithPredicates(managedResourcesPredicate)).
		Complete(r)
}
