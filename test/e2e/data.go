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
package e2e

import (
	"encoding/base64"
	"fmt"
	"log"
	"os"
	"reflect"
	"regexp"
	"testing"

	"github.com/nadundesilva/k8s-replicator/pkg/replicator"
	"github.com/nadundesilva/k8s-replicator/test/utils/validation"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/e2e-framework/klient/k8s"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
)

var (
	testResourcesFilterRegexString string
	testResourcesFilterRegex       *regexp.Regexp
)

func init() {
	testResourcesFilterRegexString = os.Getenv("TEST_RESOURCES_FILTER_REGEX")
	if testResourcesFilterRegexString == "" {
		testResourcesFilterRegexString = ".*"
	}
	testResourcesFilterRegexString = fmt.Sprintf("^%s$", testResourcesFilterRegexString)

	var err error
	testResourcesFilterRegex, err = regexp.Compile(testResourcesFilterRegexString)
	if err != nil {
		log.Fatalf("failed to initialize test resources filter: %v", err)
	}
}

type resourceTestDatum struct {
	name               string
	objectList         k8s.ObjectList
	sourceObject       k8s.Object
	sourceObjectUpdate k8s.Object
	matcher            validation.ObjectMatcher
}

func generateResourcesCreationTestData(t *testing.T) []resourceTestDatum {
	resources := []resourceTestDatum{
		{
			name:       "Secret",
			objectList: &corev1.SecretList{},
			sourceObject: &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name: envconf.RandomName("test-secret", 32),
					Labels: map[string]string{
						"e2e-tests.replicator.io/test-label-key": "test-label-value",
					},
					Annotations: map[string]string{
						"e2e-tests.replicator.io/test-annotation-key": "test-annotation-value",
					},
				},
				Data: map[string][]byte{
					"secret-data-item-one-key": []byte(base64.StdEncoding.EncodeToString([]byte("secret-data-item-one-value"))),
				},
			},
			sourceObjectUpdate: &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"e2e-tests.replicator.io/test-label-key": "test-label-value",
					},
					Annotations: map[string]string{
						"e2e-tests.replicator.io/test-annotation-key": "test-annotation-value",
					},
				},
				Data: map[string][]byte{
					"secret-data-item-two-key": []byte(base64.StdEncoding.EncodeToString([]byte("secret-data-item-two-value"))),
				},
			},
			matcher: func(sourceObject k8s.Object, replicaObject k8s.Object) bool {
				sourceSecret := sourceObject.(*corev1.Secret)
				replicaSecret := replicaObject.(*corev1.Secret)
				if !reflect.DeepEqual(sourceSecret.Data, replicaSecret.Data) {
					t.Errorf("secret data not equal; want %s, got %s",
						sourceSecret.Data, replicaSecret.Data)
				}
				return true
			},
		},
		{
			name:       "ConfigMap",
			objectList: &corev1.ConfigMapList{},
			sourceObject: &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name: envconf.RandomName("test-config-map", 32),
					Labels: map[string]string{
						"e2e-tests.replicator.io/test-label-key": "test-label-value",
					},
					Annotations: map[string]string{
						"e2e-tests.replicator.io/test-annotation-key": "test-annotation-value",
					},
				},
				BinaryData: map[string][]byte{
					"config-map-data-item-one-key": []byte(base64.StdEncoding.EncodeToString([]byte("config-map-data-item-one-value"))),
				},
			},
			sourceObjectUpdate: &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"e2e-tests.replicator.io/test-label-key": "test-label-value",
					},
					Annotations: map[string]string{
						"e2e-tests.replicator.io/test-annotation-key": "test-annotation-value",
					},
				},
				BinaryData: map[string][]byte{
					"config-map-data-item-two-key": []byte(base64.StdEncoding.EncodeToString([]byte("config-map-data-item-two-value"))),
				},
			},
			matcher: func(sourceObject k8s.Object, replicaObject k8s.Object) bool {
				sourceConfigMap := sourceObject.(*corev1.ConfigMap)
				replicaConfigMap := replicaObject.(*corev1.ConfigMap)
				if !reflect.DeepEqual(sourceConfigMap.BinaryData, replicaConfigMap.BinaryData) {
					t.Errorf("config map data not equal; want %s, got %s",
						sourceConfigMap.BinaryData, replicaConfigMap.BinaryData)
				}
				return true
			},
		},
		{
			name:       "NetworkPolicy",
			objectList: &networkingv1.NetworkPolicyList{},
			sourceObject: &networkingv1.NetworkPolicy{
				ObjectMeta: metav1.ObjectMeta{
					Name: envconf.RandomName("test-network-policy", 32),
					Labels: map[string]string{
						"e2e-tests.replicator.io/test-label-key": "test-label-value",
					},
					Annotations: map[string]string{
						"e2e-tests.replicator.io/test-annotation-key": "test-annotation-value",
					},
				},
				Spec: networkingv1.NetworkPolicySpec{
					PodSelector: metav1.LabelSelector{
						MatchLabels: map[string]string{
							"network-policy-one-selector-label-key-one": "network-policy-one-selector-label-value-one",
						},
					},
					PolicyTypes: []networkingv1.PolicyType{
						networkingv1.PolicyTypeIngress,
					},
					Ingress: []networkingv1.NetworkPolicyIngressRule{
						{
							Ports: []networkingv1.NetworkPolicyPort{
								{
									Protocol: toPointer(corev1.ProtocolTCP),
									Port:     toPointer(intstr.FromString("named-port-1")),
								},
							},
							From: []networkingv1.NetworkPolicyPeer{
								{
									NamespaceSelector: &metav1.LabelSelector{
										MatchLabels: map[string]string{
											"network-policy-one-selector-label-key-two":   "network-policy-one-selector-label-value-two",
											"network-policy-one-selector-label-key-three": "network-policy-one-selector-label-value-three",
										},
									},
									PodSelector: &metav1.LabelSelector{
										MatchExpressions: []metav1.LabelSelectorRequirement{
											{
												Key:      "network-policy-one-selector-label-key-four",
												Operator: metav1.LabelSelectorOpIn,
												Values: []string{
													"network-policy-one-selector-label-key-four-1",
													"network-policy-one-selector-label-key-four-2",
												},
											},
										},
									},
								},
								{
									IPBlock: &networkingv1.IPBlock{
										CIDR: "204.123.1.1/16",
										Except: []string{
											"204.123.214.147/24",
											"204.123.216.149/24",
										},
									},
								},
							},
						},
						{
							Ports: []networkingv1.NetworkPolicyPort{
								{
									Protocol: toPointer(corev1.ProtocolTCP),
									Port:     toPointer(intstr.FromInt(8000)),
									EndPort:  toPointer[int32](9000),
								},
							},
							From: []networkingv1.NetworkPolicyPeer{
								{
									NamespaceSelector: &metav1.LabelSelector{
										MatchLabels: map[string]string{
											"network-policy-one-selector-label-key-five": "network-policy-one-selector-label-value-five",
										},
									},
								},
							},
						},
					},
				},
			},
			sourceObjectUpdate: &networkingv1.NetworkPolicy{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"e2e-tests.replicator.io/test-label-key": "test-label-value",
					},
					Annotations: map[string]string{
						"e2e-tests.replicator.io/test-annotation-key": "test-annotation-value",
					},
				},
				Spec: networkingv1.NetworkPolicySpec{
					PodSelector: metav1.LabelSelector{
						MatchExpressions: []metav1.LabelSelectorRequirement{
							{
								Key:      "network-policy-two-selector-label-key-one",
								Operator: metav1.LabelSelectorOpIn,
								Values: []string{
									"network-policy-two-selector-label-value-one-1",
									"network-policy-two-selector-label-value-one-2",
								},
							},
						},
					},
					PolicyTypes: []networkingv1.PolicyType{
						networkingv1.PolicyTypeIngress,
						networkingv1.PolicyTypeEgress,
					},
					Ingress: []networkingv1.NetworkPolicyIngressRule{
						{
							Ports: []networkingv1.NetworkPolicyPort{
								{
									Protocol: toPointer(corev1.ProtocolUDP),
									Port:     toPointer(intstr.FromInt(9090)),
								},
							},
							From: []networkingv1.NetworkPolicyPeer{
								{
									NamespaceSelector: &metav1.LabelSelector{
										MatchLabels: map[string]string{
											"network-policy-two-selector-label-key-two": "network-policy-two-selector-label-value-two",
										},
									},
								},
							},
						},
					},
					Egress: []networkingv1.NetworkPolicyEgressRule{
						{
							Ports: []networkingv1.NetworkPolicyPort{
								{
									Protocol: toPointer(corev1.ProtocolSCTP),
									Port:     toPointer(intstr.FromInt(10098)),
								},
							},
							To: []networkingv1.NetworkPolicyPeer{
								{
									NamespaceSelector: &metav1.LabelSelector{
										MatchLabels: map[string]string{
											"network-policy-two-selector-label-key-three": "network-policy-two-selector-label-value-three",
											"network-policy-two-selector-label-key-four":  "network-policy-two-selector-label-value-four",
										},
									},
									PodSelector: &metav1.LabelSelector{
										MatchLabels: map[string]string{
											"network-policy-two-selector-label-key-five": "network-policy-two-selector-label-value-five",
											"network-policy-two-selector-label-key-six":  "network-policy-two-selector-label-value-six",
										},
									},
								},
							},
						},
					},
				},
			},
			matcher: func(sourceObject k8s.Object, replicaObject k8s.Object) bool {
				sourceNetworkPolicy := sourceObject.(*networkingv1.NetworkPolicy)
				replicaNetworkPolicy := replicaObject.(*networkingv1.NetworkPolicy)
				if !reflect.DeepEqual(sourceNetworkPolicy.Spec, replicaNetworkPolicy.Spec) {
					t.Errorf("network policy spec not equal; want %+v, got %+v",
						sourceNetworkPolicy.Spec, replicaNetworkPolicy.Spec)
				}
				return true
			},
		},
	}

	filteredResources := []resourceTestDatum{}
	for _, resource := range resources {
		if !testResourcesFilterRegex.MatchString(resource.name) {
			continue
		}
		resource.sourceObjectUpdate.SetName(resource.sourceObject.GetName())

		updateSourceObjectLabels := func(sourceObject k8s.Object) {
			labels := sourceObject.GetLabels()
			labels[replicator.ObjectTypeLabelKey] = replicator.ObjectTypeLabelValueSource
		}
		updateSourceObjectLabels(resource.sourceObject)
		updateSourceObjectLabels(resource.sourceObjectUpdate)

		filteredResources = append(filteredResources, resource)
	}
	return filteredResources
}

func toPointer[T interface{}](val T) *T {
	return &val
}
