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
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	applyconfigcorev1 "k8s.io/client-go/applyconfigurations/core/v1"
	applyconfigmetav1 "k8s.io/client-go/applyconfigurations/meta/v1"
	"k8s.io/client-go/tools/cache"
)

const KindConfigMap = "ConfigMap"

func (c *client) ConfigMapInformer() cache.SharedIndexInformer {
	return c.resourceInformerFactory.Core().V1().ConfigMaps().Informer()
}

func (c *client) ApplyConfigMap(ctx context.Context, namespace string, configMap *corev1.ConfigMap) (*corev1.ConfigMap, error) {
	applyConfig := &applyconfigcorev1.ConfigMapApplyConfiguration{
		TypeMetaApplyConfiguration: applyconfigmetav1.TypeMetaApplyConfiguration{
			Kind:       toPointer(KindConfigMap),
			APIVersion: toPointer(corev1.SchemeGroupVersion.String()),
		},
		ObjectMetaApplyConfiguration: &applyconfigmetav1.ObjectMetaApplyConfiguration{
			Name:        &configMap.Name,
			Labels:      configMap.Labels,
			Annotations: configMap.Annotations,
		},
		Data:       configMap.Data,
		BinaryData: configMap.BinaryData,
		Immutable:  configMap.Immutable,
	}
	return c.clientset.CoreV1().ConfigMaps(namespace).Apply(ctx, applyConfig, defaultApplyOptions)
}

func (c *client) ListConfigMaps(namespace string, selector labels.Selector) ([]*corev1.ConfigMap, error) {
	return c.resourceInformerFactory.Core().V1().ConfigMaps().Lister().ConfigMaps(namespace).List(selector)
}

func (c *client) GetConfigMap(ctx context.Context, namespace, name string) (*corev1.ConfigMap, error) {
	configMap, err := c.resourceInformerFactory.Core().V1().ConfigMaps().Lister().ConfigMaps(namespace).Get(name)
	if errors.IsNotFound(err) {
		return c.clientset.CoreV1().ConfigMaps(namespace).Get(ctx, name, defaultGetOptions)
	}
	return configMap, err
}

func (c *client) DeleteConfigMap(ctx context.Context, namespace, name string) error {
	return c.clientset.CoreV1().ConfigMaps(namespace).Delete(ctx, name, defaultDeleteOptions)
}
