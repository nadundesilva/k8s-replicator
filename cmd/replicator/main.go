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
package main

import (
	"log"

	"github.com/nadundesilva/k8s-replicator/pkg/kubernetes"
	"github.com/nadundesilva/k8s-replicator/pkg/replicator"
	"github.com/nadundesilva/k8s-replicator/pkg/signals"
	"go.uber.org/zap"
)

func main() {
	stopCh := signals.SetupSignalHandler()

	zapConf := zap.NewProductionConfig()
	zapLogger, err := zapConf.Build()
	if err != nil {
		log.Printf("failed to build logger config: %v", err)
	}
	defer func() {
		err := zapLogger.Sync()
		if err != nil {
			log.Printf("failed to sync logger: %v", err)
		}
	}()
	logger := zapLogger.Sugar()

	k8sBaseClient := kubernetes.NewBaseClient()

	resourceReplicators := []replicator.ResourceReplicator{
		replicator.NewSecretReplicator(kubernetes.NewSecretClient(k8sBaseClient), logger),
	}
	replicator := replicator.NewController(resourceReplicators, kubernetes.NewNamespaceClient(k8sBaseClient), logger)

	err = k8sBaseClient.Start(stopCh)
	if err != nil {
		panic(err)
	}
	err = replicator.Start(stopCh)
	if err != nil {
		panic(err)
	}
}
