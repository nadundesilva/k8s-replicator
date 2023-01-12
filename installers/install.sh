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

OLM_VERSION="v0.21.2"
K8S_REPLICATOR_NAMESPACE="k8s-replicator-system"

if ! command -v jq &> /dev/null
then
    echo "'jq' command could not be found. Please install and retry."
    exit
fi

if [ "${1}" == "" ]; then
    echo "✋ Expected version argument not provided"
    exit 1
else
    VERSION="${1}"
fi

if ! command -v operator-sdk &> /dev/null; then
    echo "🤷 operator-sdk could not be found. Please install and try again."
    exit 1
else
    echo "✅ operator-sdk found"
fi

set +e
TEMP_FILE=$(mktemp)
operator-sdk olm status 2> "${TEMP_FILE}"
OLM_INSTALLATION_STATUS_ERR=$(cat "${TEMP_FILE}")
rm "${TEMP_FILE}"
set -e
if [[ "${OLM_INSTALLATION_STATUS_ERR}" == *"no existing installation found"* ]]; then
    echo "🤷 Operator Lifecycle Manager installation not found"

    echo -n "🤔 Would you like to install Operator Lifecycle Manager into your cluster (context: $(kubectl config current-context)) [Y/n]? "
    read -r SHOULD_INSTALL_OLM
    SHOULD_INSTALL_OLM="$(tr "[:upper:]" "[:lower:]" <<< "${SHOULD_INSTALL_OLM}")"
    if [[ "${SHOULD_INSTALL_OLM}" == "y" || "${SHOULD_INSTALL_OLM}" == "" ]]; then
        operator-sdk olm install --version "${OLM_VERSION}"
        echo "✅ Operator Lifecycle Manager installation complete"
    else
        echo "✋ Operator Lifecycle Manager is required. Please install and try again"
        exit 1
    fi
else
    echo "✅ Operator Lifecycle Manager installation found"
fi

set +e
K8S_REPLICATOR_NAMESPACE_STATUS=$(kubectl get ns "${K8S_REPLICATOR_NAMESPACE}" -o json | jq .status.phase -r)
set -e
if [ "${K8S_REPLICATOR_NAMESPACE_STATUS}" == "Active" ]; then
    echo "✅ K8s Replicator namespace already exists"
elif [ "${K8S_REPLICATOR_NAMESPACE_STATUS}" == "" ]; then
    kubectl create ns "${K8S_REPLICATOR_NAMESPACE}"
    echo "✅ K8s Replicator namespace creation complete"
else
    echo "✋ K8s Replicator namespace in unexpected state: ${K8S_REPLICATOR_NAMESPACE_STATUS}"
    exit 1
fi

operator-sdk run bundle --index-image=quay.io/operator-framework/opm:v1.26.0 \
    --namespace "${K8S_REPLICATOR_NAMESPACE}" \
    "docker.io/nadunrds/k8s-replicator-bundle:${VERSION}"
echo "🏄 Completed! K8s Replicator is ready in the cluster (context: $(kubectl config current-context))"
