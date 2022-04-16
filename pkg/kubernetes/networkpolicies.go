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

	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/labels"
	applyconfignetworkingv1 "k8s.io/client-go/applyconfigurations/networking/v1"
	"k8s.io/client-go/tools/cache"
)

const KindNetworkPolicy = "NetworkPolicy"

func (c *client) NetworkPolicyInformer() cache.SharedIndexInformer {
	return c.resourceInformerFactory.Networking().V1().NetworkPolicies().Informer()
}

func (c *client) ApplyNetworkPolicy(ctx context.Context, namespace string, netpol *networkingv1.NetworkPolicy) (*networkingv1.NetworkPolicy, error) {
	specApplyConfig := applyconfignetworkingv1.NetworkPolicySpec().
		WithPolicyTypes(netpol.Spec.PolicyTypes...).
		WithPodSelector(extractLabelSelectApplyConfig(&netpol.Spec.PodSelector))

	extractPeerApplyConfig := func(netpolPeers []networkingv1.NetworkPolicyPeer) []*applyconfignetworkingv1.NetworkPolicyPeerApplyConfiguration {
		netpolPeerApplyConfigs := []*applyconfignetworkingv1.NetworkPolicyPeerApplyConfiguration{}
		for _, peer := range netpolPeers {
			netpolPeerApplyConfig := applyconfignetworkingv1.NetworkPolicyPeer()
			if peer.NamespaceSelector != nil {
				netpolPeerApplyConfig = netpolPeerApplyConfig.
					WithNamespaceSelector(extractLabelSelectApplyConfig(peer.NamespaceSelector))
			}
			if peer.PodSelector != nil {
				netpolPeerApplyConfig = netpolPeerApplyConfig.
					WithPodSelector(extractLabelSelectApplyConfig(peer.PodSelector))
			}
			if peer.IPBlock != nil {
				ipBlockApplyConfig := applyconfignetworkingv1.IPBlock().
					WithCIDR(peer.IPBlock.CIDR).
					WithExcept(peer.IPBlock.Except...)
				netpolPeerApplyConfig = netpolPeerApplyConfig.WithIPBlock(ipBlockApplyConfig)
			}
			netpolPeerApplyConfigs = append(netpolPeerApplyConfigs, netpolPeerApplyConfig)
		}
		return netpolPeerApplyConfigs
	}

	extractPortsApplyConfig := func(ports []networkingv1.NetworkPolicyPort) []*applyconfignetworkingv1.NetworkPolicyPortApplyConfiguration {
		portApplyConfigs := []*applyconfignetworkingv1.NetworkPolicyPortApplyConfiguration{}
		for _, port := range ports {
			portApplyConfig := applyconfignetworkingv1.NetworkPolicyPort()
			if port.Protocol != nil {
				portApplyConfig = portApplyConfig.WithProtocol(*port.Protocol)
			}
			if port.Port != nil {
				portApplyConfig = portApplyConfig.WithPort(*port.Port)
			}
			if port.EndPort != nil {
				portApplyConfig = portApplyConfig.WithEndPort(*port.EndPort)
			}
			portApplyConfigs = append(portApplyConfigs, portApplyConfig)
		}
		return portApplyConfigs
	}

	for _, ingress := range netpol.Spec.Ingress {
		portApplyConfigs := extractPortsApplyConfig(ingress.Ports)
		netpolPeerApplyConfigs := extractPeerApplyConfig(ingress.From)

		ingressRulesApplyConfig := applyconfignetworkingv1.NetworkPolicyIngressRule().
			WithPorts(portApplyConfigs...).
			WithFrom(netpolPeerApplyConfigs...)
		specApplyConfig = specApplyConfig.WithIngress(ingressRulesApplyConfig)
	}

	for _, egress := range netpol.Spec.Egress {
		portApplyConfigs := extractPortsApplyConfig(egress.Ports)
		netpolPeerApplyConfigs := extractPeerApplyConfig(egress.To)

		egressRulesApplyConfig := applyconfignetworkingv1.NetworkPolicyEgressRule().
			WithPorts(portApplyConfigs...).
			WithTo(netpolPeerApplyConfigs...)
		specApplyConfig = specApplyConfig.WithEgress(egressRulesApplyConfig)
	}

	applyConfig := applyconfignetworkingv1.NetworkPolicy(netpol.GetName(), namespace).
		WithLabels(netpol.GetLabels()).
		WithAnnotations(netpol.GetAnnotations()).
		WithSpec(specApplyConfig)
	return c.clientset.NetworkingV1().NetworkPolicies(namespace).Apply(ctx, applyConfig, defaultApplyOptions)
}

func (c *client) ListNetworkPolicies(namespace string, selector labels.Selector) ([]*networkingv1.NetworkPolicy, error) {
	return c.resourceInformerFactory.Networking().V1().NetworkPolicies().Lister().NetworkPolicies(namespace).List(selector)
}

func (c *client) GetNetworkPolicy(ctx context.Context, namespace, name string) (*networkingv1.NetworkPolicy, error) {
	return c.clientset.NetworkingV1().NetworkPolicies(namespace).Get(ctx, name, defaultGetOptions)
}

func (c *client) DeleteNetworkPolicy(ctx context.Context, namespace, name string) error {
	return c.clientset.NetworkingV1().NetworkPolicies(namespace).Delete(ctx, name, defaultDeleteOptions)
}
