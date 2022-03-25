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
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/cache"
)

func (c *client) NamespaceInformer() cache.SharedIndexInformer {
	return c.namespaceInformerFactory.Core().V1().Namespaces().Informer()
}

func (c *client) ListNamespaces(selector labels.Selector) ([]*corev1.Namespace, error) {
	return c.namespaceInformerFactory.Core().V1().Namespaces().Lister().List(selector)
}

func (c *client) GetNamespace(name string) (*corev1.Namespace, error) {
	return c.namespaceInformerFactory.Core().V1().Namespaces().Lister().Get(name)
}
