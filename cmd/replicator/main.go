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
	"flag"
	"fmt"
	"log"
	"net/http"
	"path/filepath"

	"github.com/nadundesilva/k8s-replicator/pkg/config"
	"github.com/nadundesilva/k8s-replicator/pkg/kubernetes"
	"github.com/nadundesilva/k8s-replicator/pkg/replicator"
	"github.com/nadundesilva/k8s-replicator/pkg/replicator/resources"
	"github.com/nadundesilva/k8s-replicator/pkg/signals"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
	"k8s.io/klog"
)

var configFilePath string

func main() {
	klog.InitFlags(nil)
	flag.Parse()

	stopCh := signals.SetupSignalHandler()
	conf, err := config.NewFromFile(configFilePath)
	if err != nil {
		log.Fatalf("failed to create configuration: %v", err)
	}

	zapConf := zap.NewProductionConfig()
	logLevel, err := zapcore.ParseLevel(conf.Logging.Level)
	if err != nil {
		log.Printf("defaulting to info log level as parsing log level %s failed: %v", conf.Logging.Level, err)
		logLevel = zapcore.InfoLevel
	}
	zapConf.Level = zap.NewAtomicLevelAt(logLevel)

	zapLogger, err := zapConf.Build()
	if err != nil {
		log.Fatalf("failed to build logger config: %v", err)
	}
	defer func() {
		_ = zapLogger.Sync()
	}()
	logger := zapLogger.Sugar()

	resourceSelectorReq, err := labels.NewRequirement(
		replicator.ObjectTypeLabelKey,
		selection.In,
		[]string{
			replicator.ObjectTypeLabelValueSource,
			replicator.ObjectTypeLabelValueReplica,
		},
	)
	if err != nil {
		logger.Fatalw("failed to initialize resources filter", "error", err)
	}

	namespaceSelectorReq, err := labels.NewRequirement(
		replicator.NamespaceTypeLabelKey,
		selection.NotEquals,
		[]string{
			replicator.NamespaceTypeLabelValueIgnored,
		},
	)
	if err != nil {
		logger.Fatalw("failed to initialize namespace filter", "error", err)
	}

	k8sClient, err := kubernetes.NewClient(
		[]labels.Requirement{*resourceSelectorReq},
		[]labels.Requirement{*namespaceSelectorReq},
		logger,
	)
	if err != nil {
		logger.Fatalw("failed to initialize kuberentes client", "error", err)
	}

	resourceReplicators := []resources.ResourceReplicator{}
	if len(conf.Resources) == 0 {
		logger.Fatalw("no resources specified in configuration to replicate")
	} else {
		availableResourceReplicators := []resources.ResourceReplicator{
			resources.NewSecretReplicator(k8sClient, logger),
			resources.NewConfigMapReplicator(k8sClient, logger),
			resources.NewNetworkPolicyReplicator(k8sClient, logger),
		}
		for _, resource := range conf.Resources {
			found := false
			for _, resourceReplicator := range availableResourceReplicators {
				if resourceReplicator.ResourceApiVersion() == resource.ApiVersion &&
					resourceReplicator.ResourceKind() == resource.Kind {
					resourceReplicators = append(resourceReplicators, resourceReplicator)
					found = true
					break
				}
			}
			if !found {
				logger.Fatalw("unsupported resource specified in configuration to be replicated", "apiVersion", resource.ApiVersion,
					"kind", resource.Kind)
			}
		}
	}
	replicator := replicator.NewController(resourceReplicators, k8sClient, logger)

	err = k8sClient.Start(stopCh)
	if err != nil {
		logger.Fatalw("failed to start k8s client", "error", err)
	}

	go func() {
		http.HandleFunc("/health", func(w http.ResponseWriter, req *http.Request) {
			fmt.Fprintf(w, "{\"status\":\"OK\"}")
		})
		err := http.ListenAndServe(":8080", nil)
		if err != nil {
			logger.Fatalw("failed to start health endpoint")
		}
	}()

	err = replicator.Start(stopCh)
	if err != nil {
		logger.Fatalw("failed to start the replicator", "error", err)
	}
}

func init() {
	defaultConfigFile := filepath.Join("/", "etc", "replicator", "config.yaml")
	flag.StringVar(&configFilePath, "config", defaultConfigFile,
		fmt.Sprintf("Path to config file. Defaults to %s", defaultConfigFile))
}
