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
package controller

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/tools/cache"
)

func (r *replicator) registerNamespaceInformer(stopCh <-chan struct{}) error {
	namespaceInformer := r.k8sClient.NamespaceInformer().Informer()
	namespaceInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: r.replicateNamespace,
		UpdateFunc: func(oldObj, newObj interface{}) {
			r.replicateNamespace(newObj)
		},
		DeleteFunc: r.removeNamespace,
	})

	if cache.WaitForCacheSync(stopCh, namespaceInformer.HasSynced) {
		r.logger.Debugw("namespace cache sync complete")
	} else {
		return fmt.Errorf("timeout waiting for namespace informer cache sync")
	}
	return nil
}

func (r *replicator) replicateNamespace(obj interface{}) {
	namespace := obj.(*corev1.Namespace)
	r.logger.Infow("Replicating namespace", "name", namespace.GetName())
}

func (r *replicator) removeNamespace(obj interface{}) {
	namespace := obj.(*corev1.Namespace)
	r.logger.Infow("Removing namespace", "name", namespace.GetName())
}
