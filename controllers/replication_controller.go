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

	"github.com/nadundesilva/k8s-replicator/controllers/replication"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

// ReplicationReconciler reconciles a replicated object
type ReplicationReconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	recorder record.EventRecorder

	Replicator replication.Replicator
}

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *ReplicationReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	ctx = log.IntoContext(ctx, log.FromContext(ctx).V(1).WithValues("replicaName", req.Name))
	log.FromContext(ctx).V(2).Info("Reconciling object")

	// Fetching object
	isObjectDeleted := false
	object := r.Replicator.EmptyObject()
	if err := r.Get(ctx, req.NamespacedName, object); err != nil {
		if errors.IsNotFound(err) {
			isObjectDeleted = true
		} else {
			return ctrl.Result{}, fmt.Errorf("failed to get object being reconciled: %+w", err)
		}
	}
	if !isObjectDeleted {
		isObjectDeleted = object.GetDeletionTimestamp() != nil
	}

	// Identifying object type
	objectType, objectTypeOk := object.GetLabels()[objectTypeLabelKey]
	if !objectTypeOk {
		logger := log.FromContext(ctx).WithValues("reason", "object type not present in object")
		if controllerutil.ContainsFinalizer(object, resourceFinalizer) {
			logger.V(1).Info("Removing replicas of unmarked object")
			return ctrl.Result{}, r.handleSourceRemoval(ctx, object)
		} else {
			logger.V(1).Info("Ignoring unmarked object")
			return ctrl.Result{}, nil
		}
	}

	// Reconciling
	switch objectType {
	case objectTypeLabelValueReplica:
		ctx = log.IntoContext(ctx, log.FromContext(ctx).WithValues("replicaNamespace", object.GetNamespace()))

		sourceStatus, err := getReplicaSourceStatus(ctx, r.Client, object, r.Replicator)
		if err != nil {
			return ctrl.Result{}, err
		}

		if sourceStatus != sourceStatusAvailable {
			logger := log.FromContext(ctx).WithValues("reason", "source object not available",
				"sourceStatus", sourceStatus)
			if isObjectDeleted {
				logger.V(1).Info("Removing finalizer from replica")
				return ctrl.Result{}, removeFinalizer(ctx, r.Client, object)
			} else {
				logger.V(1).Info("Deleting replica")
				return ctrl.Result{}, deleteObject(ctx, r.Client, object)
			}
		}
		return ctrl.Result{}, nil
	case objectTypeLabelValueReplicated:
		ctx = log.IntoContext(ctx, log.FromContext(ctx).WithValues("sourceNamespace", object.GetNamespace()))

		if isObjectDeleted {
			return ctrl.Result{}, r.handleSourceRemoval(ctx, object)
		} else {
			return ctrl.Result{}, r.handleSourceUpdate(ctx, object)
		}
	default:
		logger := log.FromContext(ctx).WithValues("objectType", objectType)
		if controllerutil.ContainsFinalizer(object, resourceFinalizer) {
			logger.V(1).Info("Removing any replicas of unknown object type if present")
			return ctrl.Result{}, r.handleSourceRemoval(ctx, object)
		} else {
			logger.V(1).Info("Ignoring unknown object type")
			return ctrl.Result{}, nil
		}
	}
}

func (r *ReplicationReconciler) handleSourceRemoval(ctx context.Context, object client.Object) error {
	err := r.iterateNamespaces(ctx, func(ns corev1.Namespace) error {
		if ns.GetName() == object.GetNamespace() {
			return nil
		}

		replica := r.Replicator.EmptyObject()
		err := r.Get(ctx, client.ObjectKey{Namespace: ns.GetName(), Name: object.GetName()}, replica)
		if err != nil {
			if errors.IsNotFound(err) {
				return nil
			} else {
				return err
			}
		}

		log.FromContext(ctx).V(1).Info("Deleting replica", "replicaNamespace", ns.GetName(),
			"reason", "source object deleted")
		err = deleteObject(ctx, r.Client, replica)
		if err != nil {
			return err
		}
		r.recorder.Eventf(object, "Normal", SourceObjectDelete, "replica in namespace %s deleted", ns.GetName())
		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to finalize source object: %+w", err)
	}

	return removeFinalizer(ctx, r.Client, object)
}

func (r *ReplicationReconciler) handleSourceUpdate(ctx context.Context, object client.Object) error {
	err := addFinalizer(ctx, r.Client, object)
	if err != nil {
		return err
	}

	err = r.iterateNamespaces(ctx, func(ns corev1.Namespace) error {
		if ns.GetName() == object.GetNamespace() {
			return nil
		}

		log.FromContext(ctx).V(1).Info("Creating/Updating replica", "replicaNamespace", ns.GetName())
		return replicateObject(ctx, r.Client, r.recorder, ns.GetName(), object, r.Replicator)
	})
	return err
}

func (r *ReplicationReconciler) iterateNamespaces(ctx context.Context, handler func(ns corev1.Namespace) error) error {
	namespaceList := &corev1.NamespaceList{}
	err := r.List(ctx, namespaceList, &client.ListOptions{
		LabelSelector: namespaceSelector,
	})
	if err != nil {
		return err
	}

	errs := []error{}
	for _, ns := range namespaceList.Items {
		if isNamespaceIgnored(&ns) || ns.GetDeletionTimestamp() != nil {
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
func (r *ReplicationReconciler) SetupWithManager(mgr ctrl.Manager) error {
	predicate := predicate.NewPredicateFuncs(func(object client.Object) bool {
		objectType, objectTypeOk := object.GetLabels()[objectTypeLabelKey]
		if objectTypeOk && (objectType == objectTypeLabelValueReplicated || objectType == objectTypeLabelValueReplica) {
			return true
		}
		return controllerutil.ContainsFinalizer(object, resourceFinalizer)
	})

	name := fmt.Sprintf("replicator-%s-controller", strings.ToLower(r.Replicator.GetKind()))
	r.recorder = mgr.GetEventRecorderFor(name)
	if r.Client == nil {
		r.Client = mgr.GetClient()
	}
	if r.Scheme == nil {
		r.Scheme = mgr.GetScheme()
	}
	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		For(r.Replicator.EmptyObject(), builder.WithPredicates(predicate)).
		WithOptions(newManagerOptions(mgr, name, r.Replicator.GetKind())).
		Complete(r)
}
