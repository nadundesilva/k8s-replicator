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
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/cache"
)

type networkPolicyReplicator struct {
	k8sClient kubernetes.ClientInterface
	logger    *zap.SugaredLogger
}

var _ ResourceReplicator = (*networkPolicyReplicator)(nil)

func NewNetworkPolicyReplicator(k8sClient kubernetes.ClientInterface, logger *zap.SugaredLogger) *networkPolicyReplicator {
	_ = k8sClient.NetworkPolicyInformer()

	return &networkPolicyReplicator{
		k8sClient: k8sClient,
		logger:    logger,
	}
}

func (r *networkPolicyReplicator) ResourceApiVersion() string {
	return networkingv1.SchemeGroupVersion.String()
}

func (r *networkPolicyReplicator) ResourceKind() string {
	return kubernetes.KindNetworkPolicy
}

func (r *networkPolicyReplicator) Informer() cache.SharedInformer {
	return r.k8sClient.NetworkPolicyInformer()
}

func (r *networkPolicyReplicator) Apply(ctx context.Context, namespace string, object metav1.Object) error {
	_, err := r.k8sClient.ApplyNetworkPolicy(ctx, namespace, object.(*networkingv1.NetworkPolicy))
	return err
}

func (r *networkPolicyReplicator) List(namespace string, selector labels.Selector) ([]metav1.Object, error) {
	netpols, err := r.k8sClient.ListNetworkPolicies(namespace, selector)
	if err != nil {
		return []metav1.Object{}, err
	}
	listObjects := []metav1.Object{}
	for _, netpol := range netpols {
		listObjects = append(listObjects, netpol)
	}
	return listObjects, nil
}

func (r *networkPolicyReplicator) Get(ctx context.Context, namespace, name string) (metav1.Object, error) {
	return r.k8sClient.GetNetworkPolicy(ctx, namespace, name)
}

func (r *networkPolicyReplicator) Delete(ctx context.Context, namespace, name string) error {
	return r.k8sClient.DeleteNetworkPolicy(ctx, namespace, name)
}
