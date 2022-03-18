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

func (r *replicator) registerSecretInformer(stopCh <-chan struct{}) error {
	secretInformer := r.k8sClient.SecretInformer().Informer()
	secretInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: r.replicateSecret,
		UpdateFunc: func(oldObj, newObj interface{}) {
			r.replicateSecret(newObj)
		},
		DeleteFunc: r.removeSecret,
	})

	if cache.WaitForCacheSync(stopCh, secretInformer.HasSynced) {
		r.logger.Debugw("secret cache sync complete")
	} else {
		return fmt.Errorf("timeout waiting for secret informer cache sync")
	}
	return nil
}

func (r *replicator) replicateSecret(obj interface{}) {
	secret := obj.(*corev1.Secret)
	r.logger.Infow("Replicating secret", "name", secret.GetName(), "namespace", secret.GetNamespace())
}

func (r *replicator) removeSecret(obj interface{}) {
	secret := obj.(*corev1.Secret)
	r.logger.Infow("Removing secret", "name", secret.GetName(), "namespace", secret.GetNamespace())
}
