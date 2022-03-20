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
package replicator

import (
	"fmt"

	"github.com/nadundesilva/k8s-replicator/pkg/kubernetes"
	"go.uber.org/zap"
	"k8s.io/client-go/tools/cache"
)

type controller struct {
	resourceReplicators []ResourceReplicator
	namespaceClient     *kubernetes.NamespaceClient
	logger              *zap.SugaredLogger
}

func NewController(resourceReplicators []ResourceReplicator, namespaceClient *kubernetes.NamespaceClient, logger *zap.SugaredLogger) *controller {
	_ = namespaceClient.Informer()

	return &controller{
		resourceReplicators: resourceReplicators,
		namespaceClient:     namespaceClient,
		logger:              logger,
	}
}

func (r *controller) Start(stopCh <-chan struct{}) error {
	r.logger.Info("starting replicator")

	informerSyncs := []cache.InformerSynced{}
	for _, resourceReplicator := range r.resourceReplicators {
		informer := resourceReplicator.GetInformer(stopCh)
		informerSyncs = append(informerSyncs, informer.HasSynced)
	}

	namespaceInformer := r.namespaceClient.Informer()
	namespaceInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: r.handleNewNamespace,
	})
	informerSyncs = append(informerSyncs, namespaceInformer.HasSynced)

	if cache.WaitForCacheSync(stopCh, informerSyncs...) {
		r.logger.Infow("cache sync complete", "informersCount", len(informerSyncs))
	} else {
		return fmt.Errorf("timeout waiting for informer cache sync")
	}

	r.logger.Info("started replicator")
	<-stopCh
	r.logger.Info("shutting down replicator")
	return nil
}
