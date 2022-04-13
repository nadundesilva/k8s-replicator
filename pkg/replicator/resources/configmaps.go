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

	"github.com/nadundesilva/k8s-replicator/pkg/kubernetes"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/cache"
)

type configMapReplicator struct {
	k8sClient kubernetes.ClientInterface
	logger    *zap.SugaredLogger
}

var _ ResourceReplicator = (*configMapReplicator)(nil)

func NewConfigMapReplicator(k8sClient kubernetes.ClientInterface, logger *zap.SugaredLogger) *configMapReplicator {
	_ = k8sClient.ConfigMapInformer()

	return &configMapReplicator{
		k8sClient: k8sClient,
		logger:    logger,
	}
}

func (r *configMapReplicator) ResourceApiVersion() string {
	return corev1.SchemeGroupVersion.String()
}

func (r *configMapReplicator) ResourceKind() string {
	return kubernetes.KindConfigMap
}

func (r *configMapReplicator) Informer() cache.SharedInformer {
	return r.k8sClient.ConfigMapInformer()
}

func (r *configMapReplicator) Clone(source metav1.Object) metav1.Object {
	sourceConfigMap := source.(*corev1.ConfigMap)
	return sourceConfigMap.DeepCopy()
}

func (r *configMapReplicator) Apply(ctx context.Context, namespace string, object metav1.Object) error {
	_, err := r.k8sClient.ApplyConfigMap(ctx, namespace, object.(*corev1.ConfigMap))
	return err
}

func (r *configMapReplicator) List(namespace string, selector labels.Selector) ([]metav1.Object, error) {
	configMaps, err := r.k8sClient.ListConfigMaps(namespace, selector)
	if err != nil {
		return []metav1.Object{}, err
	}
	listObjects := []metav1.Object{}
	for _, configMap := range configMaps {
		listObjects = append(listObjects, configMap)
	}
	return listObjects, nil
}

func (r *configMapReplicator) Get(ctx context.Context, namespace, name string) (metav1.Object, error) {
	return r.k8sClient.GetConfigMap(ctx, namespace, name)
}

func (r *configMapReplicator) Delete(ctx context.Context, namespace, name string) error {
	return r.k8sClient.DeleteConfigMap(ctx, namespace, name)
}
