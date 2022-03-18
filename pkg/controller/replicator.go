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
	"github.com/nadundesilva/k8s-replicator/pkg/kubernetes"
	"go.uber.org/zap"
)

type replicator struct {
	k8sClient kubernetes.ClientInterface
	logger    zap.SugaredLogger
}

func NewReplicator(k8sClient kubernetes.ClientInterface, logger *zap.SugaredLogger) *replicator {
	return &replicator{
		k8sClient: k8sClient,
		logger:    *logger,
	}
}

func (r *replicator) Start(stopCh <-chan struct{}) error {
	r.logger.Info("starting replicator")

	err := r.registerSecretInformer(stopCh)
	if err != nil {
		return err
	}

	r.logger.Info("started replicator")
	<-stopCh
	r.logger.Info("shutting down replicator")
	return nil
}
