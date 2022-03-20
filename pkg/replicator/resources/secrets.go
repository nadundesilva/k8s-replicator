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
	"github.com/nadundesilva/k8s-replicator/pkg/kubernetes"
	"go.uber.org/zap"
	"k8s.io/client-go/tools/cache"
)

type secretReplicator struct {
	secretClient *kubernetes.SecretClient
	logger       *zap.SugaredLogger
}

var _ ResourceReplicator = (*secretReplicator)(nil)

func NewSecretReplicator(secretClient *kubernetes.SecretClient, logger *zap.SugaredLogger) *secretReplicator {
	_ = secretClient.Informer()

	return &secretReplicator{
		secretClient: secretClient,
		logger:       logger,
	}
}

func (r *secretReplicator) Informer() cache.SharedInformer {
	return r.secretClient.Informer()
}
