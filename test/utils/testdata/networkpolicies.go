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
package testdata

import (
	"reflect"

	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
)

func generateNetworkPolicyTestDatum() Resource {
	return process(resourceData{
		Name: "NetworkPolicy",
		SourceObject: &networkingv1.NetworkPolicy{
			ObjectMeta: metav1.ObjectMeta{
				Name: envconf.RandomName("test-network-policy", 32),
				Labels: map[string]string{
					"e2e-tests.k8s-replicator.io/test-label-key": "test-label-value",
				},
				Annotations: map[string]string{
					"e2e-tests.k8s-replicator.io/test-annotation-key": "test-annotation-value",
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
								Protocol: ptr.To(corev1.ProtocolTCP),
								Port:     ptr.To(intstr.FromString("named-port-1")),
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
								Protocol: ptr.To(corev1.ProtocolTCP),
								Port:     ptr.To(intstr.FromInt(8000)),
								EndPort:  ptr.To[int32](9000),
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
		SourceObjectUpdate: &networkingv1.NetworkPolicy{
			ObjectMeta: metav1.ObjectMeta{
				Labels: map[string]string{
					"e2e-tests.k8s-replicator.io/test-label-key": "test-label-value",
				},
				Annotations: map[string]string{
					"e2e-tests.k8s-replicator.io/test-annotation-key": "test-annotation-value",
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
								Protocol: ptr.To(corev1.ProtocolUDP),
								Port:     ptr.To(intstr.FromInt(9090)),
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
								Protocol: ptr.To(corev1.ProtocolSCTP),
								Port:     ptr.To(intstr.FromInt(10098)),
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
		EmptyObject:     &networkingv1.NetworkPolicy{},
		EmptyObjectList: &networkingv1.NetworkPolicyList{},
		IsEqual: func(sourceObject client.Object, replicaObject client.Object) bool {
			sourceNetworkPolicy := sourceObject.(*networkingv1.NetworkPolicy)
			replicaNetworkPolicy := replicaObject.(*networkingv1.NetworkPolicy)
			return reflect.DeepEqual(sourceNetworkPolicy.Spec, replicaNetworkPolicy.Spec)
		},
	})
}
