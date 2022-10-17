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

K8S_REPLICATOR_NAMESPACE="k8s-replicator-system"

operator-sdk cleanup --namespace "${K8S_REPLICATOR_NAMESPACE}" --delete-all k8s-replicator
kubectl delete ns "${K8S_REPLICATOR_NAMESPACE}"

if ! command -v operator-sdk &> /dev/null; then
    echo "😢 Unable to attempt to remove Operator Lifecycle Manager since operator-sdk is not installed. Please unintall if required."
else
    echo -n "🤔 Would you like to uninstall Operator Lifecycle Manager from your cluster (context: $(kubectl config current-context)) [y/N]? "
    read -r SHOULD_UNINSTALL_OLM
    SHOULD_UNINSTALL_OLM="$(tr "[:upper:]" "[:lower:]" <<< "${SHOULD_UNINSTALL_OLM}")"
    if [[ "${SHOULD_UNINSTALL_OLM}" == "n" || "${SHOULD_UNINSTALL_OLM}" == "" ]]; then
        echo "🌟 Operator Lifecycle Manager left intact"
    else
        operator-sdk olm uninstall
        echo "✅ Operator Lifecycle Manager uninstallation complete"
    fi
fi

echo "✋ Completed! K8s Replicator is removed from the cluster (context: $(kubectl config current-context))"
