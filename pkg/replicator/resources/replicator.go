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
package resources

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/cache"
)

type ResourceReplicator interface {
	ResourceApiVersion() string
	ResourceName() string

	Informer() cache.SharedInformer
	Clone(source metav1.Object) metav1.Object

	Create(ctx context.Context, namespace string, object metav1.Object) error
	List(namespace string, selector labels.Selector) ([]metav1.Object, error)
	Get(namespace, name string) (metav1.Object, error)
	Delete(ctx context.Context, namespace, name string) error
}
