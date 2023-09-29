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
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
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
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.12.2/pkg/reconcile
func (r *SecretReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx).WithValues("secret", req.NamespacedName)
	ctx = log.IntoContext(ctx, logger)
	logger.Info("Reconciling secret")

	secret := &corev1.Secret{}
	if err := r.Get(ctx, req.NamespacedName, secret); err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		} else {
			return ctrl.Result{}, fmt.Errorf("failed to get existing secret: %+w", err)
		}
	}

	isSecretDeleted := secret.GetDeletionTimestamp() != nil
	objectType, objectTypeOk := secret.GetLabels()[ObjectTypeLabelKey]
	isReplica := objectTypeOk && objectType == ObjectTypeLabelValueReplica
	if isSecretDeleted {
		if isReplica {
			sourceNamespace, sourceNamespaceOk := secret.GetAnnotations()[SourceNamespaceAnnotationKey]
			if !sourceNamespaceOk {
				logger.Error(fmt.Errorf("%s annotation not found", SourceNamespaceAnnotationKey),
					"ignored deletion of replicated secret without source namespace label")
				return ctrl.Result{}, nil
			}

			sourceSecret := &corev1.Secret{}
			sourceSeretKey := client.ObjectKey{Namespace: sourceNamespace, Name: secret.GetName()}
			if err := r.Get(ctx, sourceSeretKey, sourceSecret); err != nil {
				if errors.IsNotFound(err) {
					return ctrl.Result{}, nil
				} else {
					return ctrl.Result{}, fmt.Errorf("failed to get existing secret: %+w", err)
				}
			}

			err := r.CreateResource(ctx, secret, secret.GetNamespace())
			if err != nil {
				return ctrl.Result{}, err
			}
		}
		return ctrl.Result{}, nil
	} else if isReplica {
		return ctrl.Result{}, nil
	}

	namespaceList := &corev1.NamespaceList{}
	err := r.List(ctx, namespaceList, &client.ListOptions{
		LabelSelector: namespaceSelector,
	})
	if err != nil {
		return ctrl.Result{}, err
	}

	errs := []error{}
	for _, ns := range namespaceList.Items {
		if strings.HasPrefix(ns.GetName(), "kube-") || (operatorNamespace != "" && ns.GetName() == operatorNamespace) {
			namespaceType, namespaceTypeOk := ns.GetLabels()[NamespaceTypeLabelKey]
			if !namespaceTypeOk || namespaceType != NamespaceTypeLabelValueManaged {
				continue
			}
		}

		err := r.CreateResource(ctx, secret, ns.GetName())
		if err != nil {
			errs = append(errs, fmt.Errorf("failed to replicate resource to namespace: %+w", err))
		}
	}
	if len(errs) > 0 {
		return ctrl.Result{}, fmt.Errorf("failed to replicate resource: %+v", errs)
	}
	return ctrl.Result{}, nil
}

func (r *SecretReconciler) CreateResource(ctx context.Context, sourceSecret *corev1.Secret, ns string) error {
	clonedSecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      sourceSecret.GetName(),
			Namespace: ns,
		},
	}
	_, err := ctrl.CreateOrUpdate(ctx, r.Client, clonedSecret, func() error {
		clonedSecret.ObjectMeta.SetLabels(map[string]string{
			ObjectTypeLabelKey:           ObjectTypeLabelValueReplica,
			SourceNamespaceAnnotationKey: sourceSecret.GetNamespace(),
		})
		clonedSecret.Immutable = sourceSecret.Immutable
		clonedSecret.Data = sourceSecret.Data
		clonedSecret.StringData = sourceSecret.StringData
		clonedSecret.Type = sourceSecret.Type
		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to replicate resource to namespace %v: %+w", ns, err)
	}
	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *SecretReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1.Secret{}, builder.WithPredicates(sourceResourcesPredicate)).
		Complete(r)
}
