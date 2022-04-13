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
package kubernetes

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/cache"
)

const FieldManager = "replicator.nadundesilva.github.io/apply"

var defaultApplyOptions = metav1.ApplyOptions{
	FieldManager: FieldManager,
}
var defaultGetOptions = metav1.GetOptions{}
var defaultDeleteOptions = metav1.DeleteOptions{
	PropagationPolicy: toPointer(metav1.DeletePropagationBackground),
}

type ClientInterface interface {
	NamespaceInformer() cache.SharedIndexInformer
	ListNamespaces(selector labels.Selector) ([]*corev1.Namespace, error)
	GetNamespace(ctx context.Context, name string) (*corev1.Namespace, error)

	SecretInformer() cache.SharedIndexInformer
	ApplySecret(ctx context.Context, namespace string, secret *corev1.Secret) (*corev1.Secret, error)
	ListSecrets(namespace string, selector labels.Selector) ([]*corev1.Secret, error)
	GetSecret(ctx context.Context, namespace, name string) (*corev1.Secret, error)
	DeleteSecret(ctx context.Context, namespace, name string) error

	ConfigMapInformer() cache.SharedIndexInformer
	ApplyConfigMap(ctx context.Context, namespace string, configMap *corev1.ConfigMap) (*corev1.ConfigMap, error)
	ListConfigMaps(namespace string, selector labels.Selector) ([]*corev1.ConfigMap, error)
	GetConfigMap(ctx context.Context, namespace, name string) (*corev1.ConfigMap, error)
	DeleteConfigMap(ctx context.Context, namespace, name string) error

	NetworkPolicyInformer() cache.SharedIndexInformer
	ApplyNetworkPolicy(ctx context.Context, namespace string, netpol *networkingv1.NetworkPolicy) (*networkingv1.NetworkPolicy, error)
	ListNetworkPolicies(namespace string, selector labels.Selector) ([]*networkingv1.NetworkPolicy, error)
	GetNetworkPolicy(ctx context.Context, namespace, name string) (*networkingv1.NetworkPolicy, error)
	DeleteNetworkPolicy(ctx context.Context, namespace, name string) error
}

func toPointer[T interface{}](val T) *T {
	return &val
}
