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
	"testing"

	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes/fake"
	clienttesting "k8s.io/client-go/testing"
)

func TestClientStartFailure(t *testing.T) {
	type testDatum struct {
		name                       string
		informerFactoryStarterFunc InformerFactoryStarterFunc
		expectedErr                error
		expectedStartCount         int
	}
	informerStartErrCount := 0
	testData := []testDatum{
		{
			name: "Start with first factory returning error",
			informerFactoryStarterFunc: func(stopCh <-chan struct{}, factory informerFactory) error {
				informerStartErrCount++
				if informerStartErrCount == 1 {
					return fmt.Errorf("test-err-1")
				} else {
					return fmt.Errorf("unexpected-number-of-calls")
				}
			},
			expectedErr:        fmt.Errorf("test-err-1"),
			expectedStartCount: 1,
		},
		{
			name: "Start with second factory returning error",
			informerFactoryStarterFunc: func(stopCh <-chan struct{}, factory informerFactory) error {
				informerStartErrCount++
				if informerStartErrCount == 1 {
					return nil
				} else if informerStartErrCount == 2 {
					return fmt.Errorf("test-err-2")
				} else {
					return fmt.Errorf("unexpected-number-of-calls")
				}
			},
			expectedErr:        fmt.Errorf("test-err-2"),
			expectedStartCount: 2,
		},
		{
			name: "Start with all factories not returning errors",
			informerFactoryStarterFunc: func(stopCh <-chan struct{}, factory informerFactory) error {
				informerStartErrCount++
				return nil
			},
			expectedErr:        nil,
			expectedStartCount: 2,
		},
	}
	for _, testDatum := range testData {
		t.Run(testDatum.name, func(t *testing.T) {
			stopCh := make(chan struct{})
			informerStartErrCount = 0

			// Initialize logger
			zapConf := zap.NewProductionConfig()
			zapLogger, err := zapConf.Build()
			if err != nil {
				t.Fatalf("failed to build logger config: %v", err)
			}
			defer func() {
				_ = zapLogger.Sync()
			}()
			logger := zapLogger.Sugar()

			// Setup mock client
			clientset := fake.NewSimpleClientset()
			watcherStarted := make(chan struct{})
			clientset.PrependWatchReactor("*", func(action clienttesting.Action) (handled bool, ret watch.Interface, err error) {
				gvr := action.GetResource()
				ns := action.GetNamespace()
				watch, err := clientset.Tracker().Watch(gvr, ns)
				if err != nil {
					return false, nil, err
				}
				close(watcherStarted)
				return true, watch, fmt.Errorf("test-informer-start-error")
			})

			// Initialize replicator kubernetes client
			k8sClient, err := NewClient(clientset, []labels.Requirement{}, []labels.Requirement{}, logger)
			if err != nil {
				t.Errorf("creating new replicator kubernetes client failed with error: %v", err)
			}

			err = k8sClient.Start(stopCh, WithInformerFactoryStarter(testDatum.informerFactoryStarterFunc))
			if testDatum.expectedErr == nil && err != nil {
				t.Errorf("failed to start k8s client with error: want %v; got %v", testDatum.expectedErr, err)
			}
			if testDatum.expectedErr != nil && err == nil {
				t.Errorf("start did not return expected error: want %v; got %v", testDatum.expectedErr, err)
			}
			if testDatum.expectedErr != nil && err != nil && testDatum.expectedErr.Error() != err.Error() {
				t.Errorf("start returned invalid error: want %v; got %v", testDatum.expectedErr, err)
			}

			if testDatum.expectedStartCount != informerStartErrCount {
				t.Errorf("unexpected number of informer start counts: want %d; got %d",
					testDatum.expectedStartCount, informerStartErrCount)
			}
		})
	}
}

type fakeInformerFactory struct {
	startCallStopCh            <-chan struct{}
	waitForCacheSyncCallStopCh <-chan struct{}
}

var _ informerFactory = (*fakeInformerFactory)(nil)

func (f *fakeInformerFactory) Start(stopCh <-chan struct{}) {
	f.startCallStopCh = stopCh
}

func (f *fakeInformerFactory) WaitForCacheSync(stopCh <-chan struct{}) map[reflect.Type]bool {
	f.waitForCacheSyncCallStopCh = stopCh
	return map[reflect.Type]bool{
		reflect.TypeOf(&corev1.ConfigMap{}).Elem(): true,
		reflect.TypeOf(&corev1.Secret{}).Elem():    false,
	}
}

func TestInformerClientStarterFuncFailure(t *testing.T) {
	expectedErr := fmt.Errorf("failed to wait for cache with type v1.Secret")
	stopCh := make(chan struct{})
	factory := &fakeInformerFactory{}

	err := startInformerFactory(stopCh, factory)
	if err == nil {
		t.Errorf("expected error not returned from start call: got %v", err)
	}
	if expectedErr.Error() != err.Error() {
		t.Errorf("returned error is not of expected type: want %v; got %v", expectedErr, err)
	}

	if stopCh != factory.startCallStopCh {
		t.Errorf("unexpected stop channel passed into start call: want %v; got %v", stopCh, factory.startCallStopCh)
	}
	if stopCh != factory.waitForCacheSyncCallStopCh {
		t.Errorf("unexpected stop channel passed into wait for cache sync call: want %v; got %v",
			stopCh, factory.waitForCacheSyncCallStopCh)
	}
}
