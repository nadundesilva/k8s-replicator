/*
 * Copyright (c) 2022, Deep Net. All Rights Reserved.
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
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type client struct {
	clientset       *kubernetes.Clientset
	informerFactory informers.SharedInformerFactory
}

var _ Interface = (*client)(nil)

func NewFromKubeConfig() *client {
	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}
	informerFactory := informers.NewSharedInformerFactory(clientset, 0)

	return &client{
		clientset:       clientset,
		informerFactory: informerFactory,
	}
}

func (c *client) Start(stopCh <-chan struct{}) error {
	_ = c.SecretInformer()

	c.informerFactory.Start(stopCh)
	return func(results ...map[reflect.Type]bool) error {
		for i := range results {
			for t, ok := range results[i] {
				if !ok {
					return fmt.Errorf("failed to wait for cache with type %s", t)
				}
			}
		}
		return nil
	}(c.informerFactory.WaitForCacheSync(stopCh))
}
