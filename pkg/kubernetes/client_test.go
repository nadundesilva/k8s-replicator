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
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"testing"
	"time"

	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes/fake"
	clienttesting "k8s.io/client-go/testing"
	"k8s.io/client-go/tools/cache"
)

func TestClientResourceOperations(t *testing.T) {
	type testDatum struct {
		name            string
		resource        string
		labelParserFunc LabelParserFunc
		newEmptyObject  func() runtime.Object
		equals          func(objA runtime.Object, objB runtime.Object) bool
		initialObjects  []runtime.Object
		informer        func(k8sClient *Client) cache.SharedIndexInformer
		apply           func(ctx context.Context, k8sClient *Client, namespace, name string) (string, metav1.Object, error)
		list            func(k8sClient *Client, namespace string) ([]metav1.Object, error)
		get             func(ctx context.Context, k8sClient *Client, namespace, name string) (metav1.Object, error)
		delete          func(ctx context.Context, k8sClient *Client, namespace, name string) error
	}
	originalLabelParserFunc := defaultLabelParserFunc
	testData := []testDatum{
		{
			name:            "Namespace operations",
			resource:        "namespaces",
			labelParserFunc: originalLabelParserFunc,
			newEmptyObject: func() runtime.Object {
				return &corev1.Namespace{}
			},
			equals: func(objA, objB runtime.Object) bool {
				namespaceA := objA.(*corev1.Namespace)
				namespaceB := objB.(*corev1.Namespace)
				return reflect.DeepEqual(namespaceA.Spec, namespaceB.Spec)
			},
			initialObjects: []runtime.Object{
				&corev1.Namespace{
					ObjectMeta: metav1.ObjectMeta{
						Name: "test-ns",
					},
				},
			},
			informer: func(k8sClient *Client) cache.SharedIndexInformer {
				return k8sClient.NamespaceInformer()
			},
			apply: nil,
			list: func(k8sClient *Client, namespace string) ([]metav1.Object, error) {
				namespaces, err := k8sClient.ListNamespaces(labels.Everything())
				objects := []metav1.Object{}
				for _, ns := range namespaces {
					objects = append(objects, ns)
				}
				return objects, err
			},
			get: func(ctx context.Context, k8sClient *Client, namespace, name string) (metav1.Object, error) {
				return k8sClient.GetNamespace(ctx, name)
			},
			delete: nil,
		},
		{
			name:     "Namespace operations with label parser failure",
			resource: "namespaces",
			labelParserFunc: func(selector string, opts ...field.PathOption) ([]labels.Requirement, error) {
				return nil, fmt.Errorf("label-parsing-error")
			},
			newEmptyObject: func() runtime.Object {
				return &corev1.Namespace{}
			},
			equals: func(objA, objB runtime.Object) bool {
				namespaceA := objA.(*corev1.Namespace)
				namespaceB := objB.(*corev1.Namespace)
				return reflect.DeepEqual(namespaceA.Spec, namespaceB.Spec)
			},
			initialObjects: []runtime.Object{
				&corev1.Namespace{
					ObjectMeta: metav1.ObjectMeta{
						Name: "test-ns",
					},
				},
			},
			informer: func(k8sClient *Client) cache.SharedIndexInformer {
				return k8sClient.NamespaceInformer()
			},
			apply: nil,
			list: func(k8sClient *Client, namespace string) ([]metav1.Object, error) {
				namespaces, err := k8sClient.ListNamespaces(labels.Everything())
				objects := []metav1.Object{}
				for _, ns := range namespaces {
					objects = append(objects, ns)
				}
				return objects, err
			},
			get: func(ctx context.Context, k8sClient *Client, namespace, name string) (metav1.Object, error) {
				return k8sClient.GetNamespace(ctx, name)
			},
			delete: nil,
		},
		{
			name:            "Secret operations",
			resource:        "secrets",
			labelParserFunc: originalLabelParserFunc,
			newEmptyObject: func() runtime.Object {
				return &corev1.Secret{}
			},
			equals: func(objA runtime.Object, objB runtime.Object) bool {
				secretA := objA.(*corev1.Secret)
				secretB := objB.(*corev1.Secret)
				return reflect.DeepEqual(secretA.StringData, secretB.StringData) &&
					reflect.DeepEqual(secretA.Data, secretB.Data) &&
					reflect.DeepEqual(secretA.Type, secretB.Type) &&
					reflect.DeepEqual(secretA.Immutable, secretB.Immutable)
			},
			initialObjects: []runtime.Object{
				&corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-secret",
						Namespace: "test-ns",
					},
					StringData: map[string]string{
						"test-secret-key-1": "test-secret-key-value-1",
					},
				},
			},
			informer: func(k8sClient *Client) cache.SharedIndexInformer {
				return k8sClient.SecretInformer()
			},
			apply: func(ctx context.Context, k8sClient *Client, namespace, name string) (string, metav1.Object, error) {
				secret := &corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      namespace,
						Namespace: name,
					},
					StringData: map[string]string{
						"test-secret-key-2": "test-secret-key-value-2",
					},
					Immutable: toPointer(false),
				}
				_, err := k8sClient.ApplySecret(ctx, "test-ns", secret)
				return KindSecret, secret, err
			},
			list: func(k8sClient *Client, namespace string) ([]metav1.Object, error) {
				secrets, err := k8sClient.ListSecrets(namespace, labels.Everything())
				objects := []metav1.Object{}
				for _, secret := range secrets {
					objects = append(objects, secret)
				}
				return objects, err
			},
			get: func(ctx context.Context, k8sClient *Client, namespace, name string) (metav1.Object, error) {
				return k8sClient.GetSecret(ctx, namespace, name)
			},
			delete: func(ctx context.Context, k8sClient *Client, namespace, name string) error {
				return k8sClient.DeleteSecret(ctx, namespace, name)
			},
		},
		{
			name:     "Secret operations with label parser failure",
			resource: "secrets",
			labelParserFunc: func(selector string, opts ...field.PathOption) ([]labels.Requirement, error) {
				return nil, fmt.Errorf("label-parsing-error")
			},
			newEmptyObject: func() runtime.Object {
				return &corev1.Secret{}
			},
			equals: func(objA runtime.Object, objB runtime.Object) bool {
				secretA := objA.(*corev1.Secret)
				secretB := objB.(*corev1.Secret)
				return reflect.DeepEqual(secretA.StringData, secretB.StringData) &&
					reflect.DeepEqual(secretA.Data, secretB.Data) &&
					reflect.DeepEqual(secretA.Type, secretB.Type) &&
					reflect.DeepEqual(secretA.Immutable, secretB.Immutable)
			},
			initialObjects: []runtime.Object{
				&corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-secret",
						Namespace: "test-ns",
					},
					StringData: map[string]string{
						"test-secret-key-1": "test-secret-key-value-1",
					},
				},
			},
			informer: func(k8sClient *Client) cache.SharedIndexInformer {
				return k8sClient.SecretInformer()
			},
			apply: func(ctx context.Context, k8sClient *Client, namespace, name string) (string, metav1.Object, error) {
				secret := &corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      namespace,
						Namespace: name,
					},
					StringData: map[string]string{
						"test-secret-key-2": "test-secret-key-value-2",
					},
					Immutable: toPointer(false),
				}
				_, err := k8sClient.ApplySecret(ctx, "test-ns", secret)
				return KindSecret, secret, err
			},
			list: func(k8sClient *Client, namespace string) ([]metav1.Object, error) {
				secrets, err := k8sClient.ListSecrets(namespace, labels.Everything())
				objects := []metav1.Object{}
				for _, secret := range secrets {
					objects = append(objects, secret)
				}
				return objects, err
			},
			get: func(ctx context.Context, k8sClient *Client, namespace, name string) (metav1.Object, error) {
				return k8sClient.GetSecret(ctx, namespace, name)
			},
			delete: func(ctx context.Context, k8sClient *Client, namespace, name string) error {
				return k8sClient.DeleteSecret(ctx, namespace, name)
			},
		},
		{
			name:            "ConfigMap operations",
			resource:        "configmaps",
			labelParserFunc: originalLabelParserFunc,
			newEmptyObject: func() runtime.Object {
				return &corev1.ConfigMap{}
			},
			equals: func(objA, objB runtime.Object) bool {
				configMapA := objA.(*corev1.ConfigMap)
				configMapB := objB.(*corev1.ConfigMap)
				return reflect.DeepEqual(configMapA.Data, configMapB.Data) &&
					reflect.DeepEqual(configMapA.BinaryData, configMapB.BinaryData) &&
					reflect.DeepEqual(configMapA.Immutable, configMapB.Immutable)
			},
			initialObjects: []runtime.Object{
				&corev1.ConfigMap{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-configmap",
						Namespace: "test-ns",
					},
					Data: map[string]string{
						"test-configmap-key-1": "test-configmap-key-value-1",
					},
				},
			},
			informer: func(k8sClient *Client) cache.SharedIndexInformer {
				return k8sClient.ConfigMapInformer()
			},
			apply: func(ctx context.Context, k8sClient *Client, namespace, name string) (string, metav1.Object, error) {
				configMap := &corev1.ConfigMap{
					ObjectMeta: metav1.ObjectMeta{
						Name:      namespace,
						Namespace: name,
					},
					Data: map[string]string{
						"test-configmap-key-2": "test-configmap-key-value-2",
					},
					Immutable: toPointer(false),
				}
				_, err := k8sClient.ApplyConfigMap(ctx, "test-ns", configMap)
				return KindConfigMap, configMap, err
			},
			list: func(k8sClient *Client, namespace string) ([]metav1.Object, error) {
				configMaps, err := k8sClient.ListConfigMaps(namespace, labels.Everything())
				objects := []metav1.Object{}
				for _, configMap := range configMaps {
					objects = append(objects, configMap)
				}
				return objects, err
			},
			get: func(ctx context.Context, k8sClient *Client, namespace, name string) (metav1.Object, error) {
				return k8sClient.GetConfigMap(ctx, namespace, name)
			},
			delete: func(ctx context.Context, k8sClient *Client, namespace, name string) error {
				return k8sClient.DeleteConfigMap(ctx, namespace, name)
			},
		},
		{
			name:            "NetworkPolicy operations",
			resource:        "networkpolicies",
			labelParserFunc: originalLabelParserFunc,
			newEmptyObject: func() runtime.Object {
				return &networkingv1.NetworkPolicy{}
			},
			equals: func(objA, objB runtime.Object) bool {
				networkPolicyA := objA.(*networkingv1.NetworkPolicy)
				networkPolicyB := objB.(*networkingv1.NetworkPolicy)
				return reflect.DeepEqual(networkPolicyA.Spec, networkPolicyB.Spec)
			},
			initialObjects: []runtime.Object{
				&networkingv1.NetworkPolicy{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-networkpolicy",
						Namespace: "test-ns",
					},
					Spec: networkingv1.NetworkPolicySpec{
						PolicyTypes: []networkingv1.PolicyType{
							networkingv1.PolicyTypeIngress,
						},
						PodSelector: metav1.LabelSelector{
							MatchLabels: map[string]string{
								"pod-type": "foo",
							},
						},
						Ingress: []networkingv1.NetworkPolicyIngressRule{
							{
								From: []networkingv1.NetworkPolicyPeer{
									{
										PodSelector: &metav1.LabelSelector{
											MatchExpressions: []metav1.LabelSelectorRequirement{
												{
													Key:      "pod-type",
													Operator: metav1.LabelSelectorOpIn,
													Values:   []string{"bar", "baz"},
												},
											},
										},
										NamespaceSelector: &metav1.LabelSelector{
											MatchLabels: map[string]string{
												"visibility": "restricted",
											},
										},
										IPBlock: &networkingv1.IPBlock{
											CIDR: "204.214.1.1/16",
											Except: []string{
												"204.214.214.147/24",
												"204.214.216.149/24",
											},
										},
									},
								},
								Ports: []networkingv1.NetworkPolicyPort{
									{
										Protocol: toPointer(corev1.ProtocolSCTP),
										Port:     toPointer(intstr.FromInt(10098)),
									},
								},
							},
						},
					},
				},
			},
			informer: func(k8sClient *Client) cache.SharedIndexInformer {
				return k8sClient.NetworkPolicyInformer()
			},
			apply: func(ctx context.Context, k8sClient *Client, namespace, name string) (string, metav1.Object, error) {
				networkPolicy := &networkingv1.NetworkPolicy{
					ObjectMeta: metav1.ObjectMeta{
						Name:      name,
						Namespace: namespace,
					},
					Spec: networkingv1.NetworkPolicySpec{
						PolicyTypes: []networkingv1.PolicyType{
							networkingv1.PolicyTypeIngress,
							networkingv1.PolicyTypeEgress,
						},
						PodSelector: metav1.LabelSelector{
							MatchLabels: map[string]string{
								"pod-type": "foo",
							},
						},
						Ingress: []networkingv1.NetworkPolicyIngressRule{
							{
								From: []networkingv1.NetworkPolicyPeer{
									{
										PodSelector: &metav1.LabelSelector{
											MatchExpressions: []metav1.LabelSelectorRequirement{
												{
													Key:      "pod-type",
													Operator: metav1.LabelSelectorOpIn,
													Values:   []string{"bar", "baz"},
												},
											},
										},
										NamespaceSelector: &metav1.LabelSelector{
											MatchLabels: map[string]string{
												"visibility": "restricted",
											},
										},
										IPBlock: &networkingv1.IPBlock{
											CIDR: "204.213.1.1/16",
											Except: []string{
												"204.213.214.147/24",
												"204.213.216.149/24",
											},
										},
									},
								},
								Ports: []networkingv1.NetworkPolicyPort{
									{
										Protocol: toPointer(corev1.ProtocolSCTP),
										Port:     toPointer(intstr.FromInt(10098)),
									},
								},
							},
						},
						Egress: []networkingv1.NetworkPolicyEgressRule{
							{
								To: []networkingv1.NetworkPolicyPeer{
									{
										PodSelector: &metav1.LabelSelector{
											MatchExpressions: []metav1.LabelSelectorRequirement{
												{
													Key:      "pod-type",
													Operator: metav1.LabelSelectorOpIn,
													Values:   []string{"bar", "baz", "bam"},
												},
											},
										},
										NamespaceSelector: &metav1.LabelSelector{
											MatchLabels: map[string]string{
												"visibility": "hidden",
											},
										},
										IPBlock: &networkingv1.IPBlock{
											CIDR: "204.215.1.1/16",
											Except: []string{
												"204.215.214.147/24",
												"204.215.216.149/24",
											},
										},
									},
								},
								Ports: []networkingv1.NetworkPolicyPort{
									{
										Protocol: toPointer(corev1.ProtocolTCP),
										Port:     toPointer(intstr.FromInt(8000)),
										EndPort:  toPointer[int32](9000),
									},
								},
							},
						},
					},
				}
				_, err := k8sClient.ApplyNetworkPolicy(ctx, "test-ns", networkPolicy)
				return KindNetworkPolicy, networkPolicy, err
			},
			list: func(k8sClient *Client, namespace string) ([]metav1.Object, error) {
				networkPolicies, err := k8sClient.ListNetworkPolicies(namespace, labels.Everything())
				objects := []metav1.Object{}
				for _, networkPolicy := range networkPolicies {
					objects = append(objects, networkPolicy)
				}
				return objects, err
			},
			get: func(ctx context.Context, k8sClient *Client, namespace, name string) (metav1.Object, error) {
				return k8sClient.GetNetworkPolicy(ctx, namespace, name)
			},
			delete: func(ctx context.Context, k8sClient *Client, namespace, name string) error {
				return k8sClient.DeleteNetworkPolicy(ctx, namespace, name)
			},
		},
	}
	for _, testDatum := range testData {
		t.Run(testDatum.name, func(t *testing.T) {
			defaultLabelParserFunc = testDatum.labelParserFunc
			stopCh := make(chan struct{})
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

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
			clientset := fake.NewSimpleClientset(testDatum.initialObjects...)
			watcherStarted := make(chan struct{})
			clientset.PrependWatchReactor("*", func(action clienttesting.Action) (handled bool, ret watch.Interface, err error) {
				gvr := action.GetResource()
				ns := action.GetNamespace()
				watch, err := clientset.Tracker().Watch(gvr, ns)
				if err != nil {
					return false, nil, err
				}
				close(watcherStarted)
				return true, watch, nil
			})
			clientset.PrependReactor("patch", testDatum.resource, func(action clienttesting.Action) (handled bool, ret runtime.Object, err error) {
				if patchAction, ok := action.(clienttesting.PatchAction); ok {
					object := testDatum.newEmptyObject()
					err := json.Unmarshal(patchAction.GetPatch(), object)
					if err != nil {
						t.Errorf("failed to transform patch to secret: %v", err)
					}

					err = clientset.Tracker().Create(action.GetResource(), object, action.GetNamespace())
					if err != nil {
						return true, nil, err
					}
					return true, object, nil
				} else {
					t.Errorf("not patch type type")
				}
				return false, nil, err
			})

			// Initialize replicator kubernetes client
			k8sClient, err := NewClient(clientset, []labels.Requirement{}, []labels.Requirement{}, logger)
			if err != nil {
				t.Errorf("creating new replicator kubernetes client failed with error: %v", err)
			}
			informer := testDatum.informer(k8sClient)

			// Start informers with event handlers
			type EventType string
			const (
				Add    EventType = "Add"
				Update EventType = "Update"
				Delete EventType = "Delete"
			)
			type Event struct {
				eventType EventType
				prevObj   interface{}
				newObj    interface{}
			}
			events := make(chan *Event, 1)
			informer.AddEventHandler(&cache.ResourceEventHandlerFuncs{
				AddFunc: func(obj interface{}) {
					events <- &Event{
						eventType: Add,
						prevObj:   nil,
						newObj:    obj,
					}
				},
				UpdateFunc: func(oldObj, newObj interface{}) {
					events <- &Event{
						eventType: Update,
						prevObj:   oldObj,
						newObj:    newObj,
					}
				},
				DeleteFunc: func(obj interface{}) {
					events <- &Event{
						eventType: Delete,
						prevObj:   obj,
						newObj:    nil,
					}
				},
			})
			err = k8sClient.Start(stopCh)
			if err != nil {
				t.Errorf("failed to start k8s client with error: %v", err)
			}

			if !cache.WaitForCacheSync(stopCh, informer.HasSynced) {
				t.Errorf("timeout waiting for informers cache sync")
			}
			<-watcherStarted

			// Validate informer notifying the initial objects
			for i := range testDatum.initialObjects {
				select {
				case event := <-events:
					if event.eventType != Add {
						t.Errorf("received event of unexpected type %s", event.eventType)
					}
					if !testDatum.equals(event.newObj.(runtime.Object), testDatum.initialObjects[i].(runtime.Object)) {
						t.Errorf("received add event with unexpected object: want %+v; got %+v", testDatum.initialObjects[i], event.newObj)
					}
				case <-time.After(wait.ForeverTestTimeout):
					t.Errorf("informer did not get the expected initial event of type %s", Add)
				}
			}

			// Validating informer list calls
			if len(clientset.Actions()) != 2 {
				t.Errorf("unexpected number of client set actions called for informer sync: want %d; got %d",
					2, len(clientset.Actions()))
			}

			firstObject := testDatum.initialObjects[0]
			firstObjectNamespace := firstObject.(metav1.Object).GetNamespace()
			firstObjectName := firstObject.(metav1.Object).GetName()
			expectedObjectsCount := len(testDatum.initialObjects)

			// Validate applying patch and the informer notifying of new object
			if testDatum.apply != nil {
				clientset.ClearActions()
				kind, applyObject, err := testDatum.apply(ctx, k8sClient, firstObjectNamespace, fmt.Sprintf("%s-new", firstObjectName))
				if err != nil {
					t.Errorf("applying resource failed with error: %v", err)
				}
				select {
				case event := <-events:
					if event.eventType != Add {
						t.Errorf("received event of unexpected type %s", event.eventType)
					}
					if !testDatum.equals(event.newObj.(runtime.Object), applyObject.(runtime.Object)) {
						t.Errorf("received add event with unexpected object: want %+v; got %+v", applyObject, event.newObj)
					}
				case <-time.After(wait.ForeverTestTimeout):
					t.Errorf("informer did not get the expected event of type %s and kind %s", Add, kind)
				}
				expectedObjectsCount++

				// Validate applying patch action
				if len(clientset.Actions()) != 1 {
					t.Errorf("unexpected number of client set actions called for apply: want %d; got %d", 1, len(clientset.Actions()))
				} else {
					action := clientset.Actions()[0]
					if action.GetVerb() != "patch" {
						t.Errorf("unexpected verb used in the client set action for apply: want %s; got %s", "patch", action.GetVerb())
					}
					if action.GetResource().Resource != testDatum.resource {
						t.Errorf("unexpected resource used in the client set action for apply: want %s; got %s",
							testDatum.resource, action.GetResource().Resource)
					}
				}
			}

			// Validate listing the objects
			clientset.ClearActions()
			objects, err := testDatum.list(k8sClient, firstObjectNamespace)
			if err != nil {
				t.Errorf("listing resource failed with error: %v", err)
			}
			if len(objects) != expectedObjectsCount {
				t.Errorf("listing resource returned an unexpected amount of resources: want %d; got %d",
					expectedObjectsCount, len(objects))
			}

			// Validate listing action (using informer lister)
			if len(clientset.Actions()) != 0 {
				t.Errorf("unexpected number of client set actions called for list: want %d; got %d", 0, len(clientset.Actions()))
			}

			// Validate getting an object
			clientset.ClearActions()
			object, err := testDatum.get(ctx, k8sClient, firstObjectNamespace, firstObjectName)
			if err != nil {
				t.Errorf("getting resource failed with error: %v", err)
			}
			if !testDatum.equals(object.(runtime.Object), firstObject.(runtime.Object)) {
				t.Errorf("resource returned in get call did not match the requested initial object: want %+v; got %+v",
					firstObject, object)
			}

			// Validate getting action
			if len(clientset.Actions()) != 1 {
				t.Errorf("unexpected number of client set actions called for get: want %d; got %d", 1, len(clientset.Actions()))
			} else {
				action := clientset.Actions()[0]
				if action.GetVerb() != "get" {
					t.Errorf("unexpected verb used in the client set action for get: want %s; got %s", "get", action.GetVerb())
				}
				if action.GetResource().Resource != testDatum.resource {
					t.Errorf("unexpected resource used in the client set action for get: want %s; got %s",
						testDatum.resource, action.GetResource().Resource)
				}
			}

			// Validate deleting an object and the informer notifying the deletion
			if testDatum.delete != nil {
				clientset.ClearActions()
				err = testDatum.delete(ctx, k8sClient, firstObjectNamespace, firstObjectName)
				if err != nil {
					t.Errorf("getting resource failed with error: %v", err)
				}
				select {
				case event := <-events:
					if event.eventType != Delete {
						t.Errorf("received event of unexpected type %s", event.eventType)
					}
					if !testDatum.equals(event.prevObj.(runtime.Object), firstObject.(runtime.Object)) {
						t.Errorf("received delete event with unexpected object: want %+v; got %+v", firstObject, event.prevObj)
					}
				case <-time.After(wait.ForeverTestTimeout):
					t.Errorf("informer did not get the expected event of type %s", Delete)
				}

				// Validate deleting action
				if len(clientset.Actions()) != 1 {
					t.Errorf("unexpected number of client set actions called for delete: want %d; got %d", 1, len(clientset.Actions()))
				} else {
					action := clientset.Actions()[0]
					if action.GetVerb() != "delete" {
						t.Errorf("unexpected verb used in the client set action for delete: want %s; got %s", "delete", action.GetVerb())
					}
					if action.GetResource().Resource != testDatum.resource {
						t.Errorf("unexpected resource used in the client set action for delete: want %s; got %s",
							testDatum.resource, action.GetResource().Resource)
					}
				}
			}

			close(stopCh)
		})
	}
	defaultLabelParserFunc = originalLabelParserFunc
}
