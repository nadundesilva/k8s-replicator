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

	"k8s.io/client-go/informers"
)

type InformerFactoryStarterFunc func(<-chan struct{}, informers.SharedInformerFactory) error

func startInformerFactory(stopCh <-chan struct{}, factory informers.SharedInformerFactory) error {
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

type StartOptions struct {
	informerFactoryStarterFunc InformerFactoryStarterFunc
}

type StartOption func(*StartOptions)

func WithInformerFactoryStarter(informerFactoryStarterFunc InformerFactoryStarterFunc) StartOption {
	return func(options *StartOptions) {
		options.informerFactoryStarterFunc = informerFactoryStarterFunc
	}
}

func (c *Client) Start(stopCh <-chan struct{}, opts ...StartOption) error {
	options := &StartOptions{
		informerFactoryStarterFunc: startInformerFactory,
	}
	for _, opt := range opts {
		opt(options)
	}

	err := options.informerFactoryStarterFunc(stopCh, c.namespaceInformerFactory)
	if err != nil {
		return err
	}
	return options.informerFactoryStarterFunc(stopCh, c.resourceInformerFactory)
}
