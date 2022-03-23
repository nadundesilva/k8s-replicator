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
	"k8s.io/client-go/tools/cache"
)

type secretReplicator struct {
	k8sClient kubernetes.ClientInterface
	logger    *zap.SugaredLogger
}

var _ ResourceReplicator = (*secretReplicator)(nil)

func NewSecretReplicator(k8sClient kubernetes.ClientInterface, logger *zap.SugaredLogger) *secretReplicator {
	_ = k8sClient.SecretInformer()

	return &secretReplicator{
		k8sClient: k8sClient,
		logger:    logger,
	}
}

func (r *secretReplicator) ResourceApiVersion() string {
	return "v1"
}

func (r *secretReplicator) ResourceName() string {
	return "Secret"
}

func (r *secretReplicator) Informer() cache.SharedInformer {
	return r.k8sClient.SecretInformer()
}

func (r *secretReplicator) Clone(source metav1.Object) metav1.Object {
	sourceSecret := source.(*corev1.Secret)
	clonedSecret := &corev1.Secret{
		Type:       sourceSecret.Type,
		Data:       map[string][]byte{},
		StringData: map[string]string{},
	}
	for k, v := range sourceSecret.Data {
		clonedSecret.Data[k] = v
	}
	for k, v := range sourceSecret.StringData {
		clonedSecret.StringData[k] = v
	}
	return clonedSecret
}

func (r *secretReplicator) Create(ctx context.Context, namespace string, object metav1.Object) error {
	_, err := r.k8sClient.CreateSecret(ctx, namespace, object.(*corev1.Secret))
	return err
}

func (r *secretReplicator) Get(ctx context.Context, namespace, name string) (metav1.Object, error) {
	return r.k8sClient.GetSecret(ctx, namespace, name)
}
