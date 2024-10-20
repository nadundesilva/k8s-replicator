#!/usr/bin/env bash
# Copyright (c) 2022, Nadun De Silva. All Rights Reserved.
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#   http://www.apache.org/licenses/LICENSE-2.0
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
set -e

validate_namespace() {
	NAMESPACE="${1}"
	HOSTNAME="${2}"

	echo "🔎 Validating ${NAMESPACE} namespace"
	NAMESPACE_NAME="$(kubectl get ns "${NAMESPACE}" -o=jsonpath='{.metadata.name}')"
	if [ "${NAMESPACE_NAME}" != "${NAMESPACE}" ]; then
		echo "${NAMESPACE_NAME}"
		echo "✋ Expected namespace ${NAMESPACE} not found"
		exit 1
	fi

	echo "🔎 Validating vscode pod in ${NAMESPACE} namespace"
	kubectl wait --namespace "${NAMESPACE}" \
		--for=condition=ready pod \
		--selector=example=cert-manager \
		--timeout=60s

	echo "🔎 Validating vscode-wildcard-tls secret in ${NAMESPACE} namespace"
	SECRET_NAME="$(kubectl get secret vscode-wildcard-tls -n "${NAMESPACE}" -o=jsonpath='{.metadata.name}')"
	if [ "${SECRET_NAME}" != "vscode-wildcard-tls" ]; then
		echo "${SECRET_NAME}"
		echo "✋ Expected secret vscode-wildcard-tls not found in namepsace ${NAMESPACE}"
		exit 1
	fi

	echo "🔎 Validating vscode ingress in ${NAMESPACE} namespace"
	INGRESS_NAME="$(kubectl get ingress vscode -n "${NAMESPACE}" -o=jsonpath='{.metadata.name}')"
	if [ "${INGRESS_NAME}" != "vscode" ]; then
		echo "${INGRESS_NAME}"
		echo "✋ Expected ingress vscode not found in namepsace ${NAMESPACE}"
		exit 1
	fi

	if [ "${K8S_CLUSTER_IP}" != "" ]; then
		echo "🔎 Validating accessing vscode editor served under ${HOSTNAME}"
		sudo echo "${K8S_CLUSTER_IP} ${HOSTNAME}" | sudo tee -a /etc/hosts
		kubectl get secret -n kr-cert-issuer vscode-wildcard-tls -o jsonpath='{.data.ca\.crt}' | base64 --decode >ca.crt
		curl --cacert ca.crt "https://${HOSTNAME}/"
	else
		echo "🤔 Skipping accessing the editor certificate"
	fi
}

echo "🌟 Validating Cert Manager example"

echo
./setup.sh

echo
validate_namespace kr-vscode-01 editor-01.vscode.local
echo
validate_namespace kr-vscode-02 editor-02.vscode.local
echo
validate_namespace kr-vscode-03 editor-03.vscode.local

echo
./clean.sh
