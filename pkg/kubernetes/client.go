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
	"fmt"
	"reflect"
	"time"

	"go.uber.org/zap"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type client struct {
	clientset                *kubernetes.Clientset
	namespaceInformerFactory informers.SharedInformerFactory
	resourceInformerFactory  informers.SharedInformerFactory
}

var _ ClientInterface = (*client)(nil)

func NewClient(resourceSelectorReqs, namespaceSelectorReqs []labels.Requirement, logger *zap.SugaredLogger) (*client, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to detect in cluster configuration: %w", err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize underlying kubernetes client: %w", err)
	}

	withNewRequirements := func(newReqs []labels.Requirement) informers.SharedInformerOption {
		return informers.WithTweakListOptions(func(options *metav1.ListOptions) {
			requirements, err := labels.ParseToRequirements(options.LabelSelector)
			if err != nil {
				logger.Errorw("failed to parse label selector", "error", err)
				return
			}

			selector := labels.NewSelector().Add(requirements...).Add(newReqs...)
			options.LabelSelector = selector.String()
		})
	}

	namespaceInformerFactory := informers.NewSharedInformerFactoryWithOptions(clientset, time.Minute*5,
		withNewRequirements(namespaceSelectorReqs))
	resourceInformerFactory := informers.NewSharedInformerFactoryWithOptions(clientset, time.Minute*5,
		withNewRequirements(resourceSelectorReqs))

	return &client{
		clientset:                clientset,
		namespaceInformerFactory: namespaceInformerFactory,
		resourceInformerFactory:  resourceInformerFactory,
	}, nil
}

func (c *client) Start(stopCh <-chan struct{}) error {
	err := c.startInformerFactory(stopCh, c.namespaceInformerFactory)
	if err != nil {
		return err
	}
	return c.startInformerFactory(stopCh, c.resourceInformerFactory)
}

func (c *client) startInformerFactory(stopCh <-chan struct{}, factory informers.SharedInformerFactory) error {
	factory.Start(stopCh)
	return func(results ...map[reflect.Type]bool) error {
		for i := range results {
			for t, ok := range results[i] {
				if !ok {
					return fmt.Errorf("failed to wait for cache with type %s", t)
				}
			}
		}
		return nil
	}(factory.WaitForCacheSync(stopCh))
}
