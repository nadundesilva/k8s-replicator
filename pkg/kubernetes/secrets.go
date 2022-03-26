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
	"k8s.io/apimachinery/pkg/labels"
	applyconfigcorev1 "k8s.io/client-go/applyconfigurations/core/v1"
	applyconfigmetav1 "k8s.io/client-go/applyconfigurations/meta/v1"
	"k8s.io/client-go/tools/cache"
)

const KindSecret = "Secret"

func (c *client) SecretInformer() cache.SharedIndexInformer {
	return c.resourceInformerFactory.Core().V1().Secrets().Informer()
}

func (c *client) ApplySecret(ctx context.Context, namespace string, secret *corev1.Secret) (*corev1.Secret, error) {
	applyConfig := &applyconfigcorev1.SecretApplyConfiguration{
		TypeMetaApplyConfiguration: applyconfigmetav1.TypeMetaApplyConfiguration{
			Kind:       toPointer(KindSecret),
			APIVersion: toPointer(corev1.SchemeGroupVersion.String()),
		},
		ObjectMetaApplyConfiguration: &applyconfigmetav1.ObjectMetaApplyConfiguration{
			Name:        &secret.Name,
			Labels:      secret.Labels,
			Annotations: secret.Annotations,
		},
		Type:       &secret.Type,
		Data:       secret.Data,
		StringData: secret.StringData,
		Immutable:  secret.Immutable,
	}
	return c.clientset.CoreV1().Secrets(namespace).Apply(ctx, applyConfig, defaultApplyOptions)
}

func (c *client) ListSecrets(namespace string, selector labels.Selector) ([]*corev1.Secret, error) {
	return c.resourceInformerFactory.Core().V1().Secrets().Lister().Secrets(namespace).List(selector)
}

func (c *client) GetSecret(namespace, name string) (*corev1.Secret, error) {
	return c.resourceInformerFactory.Core().V1().Secrets().Lister().Secrets(namespace).Get(name)
}

func (c *client) DeleteSecret(ctx context.Context, namespace, name string) error {
	return c.clientset.CoreV1().Secrets(namespace).Delete(ctx, name, defaultDeleteOptions)
}
