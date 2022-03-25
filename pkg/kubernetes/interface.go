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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/cache"
)

var defaultCreateOptions = metav1.CreateOptions{}
var defaultDeleteOptions = metav1.DeleteOptions{
	PropagationPolicy: toPointer(metav1.DeletePropagationBackground),
}

type ClientInterface interface {
	SecretInformer() cache.SharedIndexInformer
	NamespaceInformer() cache.SharedIndexInformer
	GetNamespace(name string) (*corev1.Namespace, error)

	ListNamespaces(selector labels.Selector) ([]*corev1.Namespace, error)

	CreateSecret(ctx context.Context, namespace string, secret *corev1.Secret) (*corev1.Secret, error)
	ListSecrets(namespace string, selector labels.Selector) ([]*corev1.Secret, error)
	GetSecret(namespace, name string) (*corev1.Secret, error)
	DeleteSecret(ctx context.Context, namespace, name string) error
}

func toPointer[T interface{}](val T) *T {
	return &val
}
