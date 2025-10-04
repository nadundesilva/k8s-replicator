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

	"github.com/nadundesilva/k8s-replicator/controllers/replication"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

// NamespaceReconciler reconciles a Namespace object
type NamespaceReconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	recorder record.EventRecorder

	Replicators []replication.Replicator
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
	ctx = log.IntoContext(ctx, log.FromContext(ctx).V(1).WithValues("targetNamespace", req.Name))
	log.FromContext(ctx).V(2).Info("Reconciling namespace")

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
	namespaceName := req.Name

	if !isNamespaceDeleted {
		isNamespaceDeleted = namespace.GetDeletionTimestamp() != nil
	}
	isNamespaceIgnored := isNamespaceIgnored(namespace)

	// Reconciling
	if isNamespaceDeleted || isNamespaceIgnored {
		for _, replicator := range r.Replicators {
			ctx := log.IntoContext(ctx, log.FromContext(ctx).WithValues("objectKind", replicator.GetKind()))
			log.FromContext(ctx).V(2).Info("Replicating object kind")

			replicaObjects := replicator.EmptyObjectList()
			err := r.List(ctx, replicaObjects, &client.ListOptions{
				Namespace:     namespaceName,
				LabelSelector: replicaResourcesSelector,
			})
			if err != nil {
				return ctrl.Result{}, err
			}

			errs := []error{}
			for _, object := range replicator.ObjectListToArray(replicaObjects) {
				ctx := log.IntoContext(ctx, log.FromContext(ctx).WithValues("sourceNamespace", object.GetNamespace(),
					"replicaName", object.GetName()))

				if isNamespaceDeleted {
					log.FromContext(ctx).V(1).Info("Removing finalizer from replica in deleted namespace")
					err := removeFinalizer(ctx, r.Client, object)
					if err != nil {
						errs = append(errs, fmt.Errorf("failed to remove finalizer from replica in deleted namespace: %+w", err))
					}
				}
				if isNamespaceIgnored {
					log.FromContext(ctx).V(1).Info("Deleting replica in ignored namespace")
					err := deleteObject(ctx, r.Client, object)
					if err != nil {
						errs = append(errs, fmt.Errorf("failed to delete object: %+w", err))
					}
				}
			}
			if len(errs) > 0 {
				return ctrl.Result{}, fmt.Errorf("failed to iterate replicated objects in removed namespace: %+v", errs)
			}
		}
	} else {
		for _, replicator := range r.Replicators {
			ctx := log.IntoContext(ctx, log.FromContext(ctx).WithValues("objectKind", replicator.GetKind()))
			log.FromContext(ctx).V(2).Info("Replicating object kind")

			replicatedObjects := replicator.EmptyObjectList()
			err := r.List(ctx, replicatedObjects)
			if err != nil {
				return ctrl.Result{}, err
			}

			errs := []error{}
			for _, object := range replicator.ObjectListToArray(replicatedObjects) {
				ctx := log.IntoContext(ctx, log.FromContext(ctx).WithValues("sourceNamespace", object.GetNamespace(),
					"replicaName", object.GetName()))

				if objectType, objectTypeOk := object.GetLabels()[objectTypeLabelKey]; objectTypeOk && objectType == objectTypeLabelValueReplicated {
					if object.GetDeletionTimestamp() != nil { // Object already deleted
						log.FromContext(ctx).V(2).Info("Ignoring deleted object")
						continue
					}
					if object.GetNamespace() == namespaceName { // Replicated resource is from current namespace
						log.FromContext(ctx).V(2).Info("Ignoring source object in current namespace")
						continue
					}

					log.FromContext(ctx).V(1).Info("Creating/Updating replica")
					err := replicateObject(ctx, r.Client, r.recorder, namespaceName, object, replicator)
					if err != nil {
						errs = append(errs, err)
					}
				}
			}
			if len(errs) > 0 {
				return ctrl.Result{}, fmt.Errorf("failed to iterate replicated objects in namespace: %+v", errs)
			}
		}
	}
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *NamespaceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	isReconciled := func(object client.Object) bool {
		return !isNamespaceIgnored(object.(*corev1.Namespace))
	}
	predicate := predicate.Funcs{
		CreateFunc: func(ce event.CreateEvent) bool {
			return isReconciled(ce.Object)
		},
		UpdateFunc: func(ue event.UpdateEvent) bool {
			return isReconciled(ue.ObjectOld) || isReconciled(ue.ObjectNew)
		},
		DeleteFunc: func(de event.DeleteEvent) bool {
			return isReconciled(de.Object)
		},
		GenericFunc: func(ge event.GenericEvent) bool {
			return isReconciled(ge.Object)
		},
	}

	name := "replicator-namespace-controller"
	r.recorder = mgr.GetEventRecorderFor(name)
	if r.Client == nil {
		r.Client = mgr.GetClient()
	}
	if r.Scheme == nil {
		r.Scheme = mgr.GetScheme()
	}
	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		For(&corev1.Namespace{}, builder.WithPredicates(predicate)).
		WithOptions(newManagerOptions(mgr, name, "Namespace")).
		Complete(r)
}
