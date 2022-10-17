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

if [ "${1}" == "" ]; then
    echo "✋ Expected version argument not provided"
    exit 1
else
    VERSION="${1}"
fi

TEMP_DIR=$(mktemp -d)
echo "🌟 Using temporary directory ${TEMP_DIR}"

cleanup() {
    rm -rf "${TEMP_DIR}"
}
trap cleanup EXIT

DOWNLOAD_DIR="${TEMP_DIR}/k8s-replicator"
echo "🚜 Downloading release ${VERSION} to dir ${DOWNLOAD_DIR}"
echo
curl -L -o "${DOWNLOAD_DIR}.zip" "https://github.com/nadundesilva/k8s-replicator/releases/download/v${VERSION}/k8s-replicator-v${VERSION}.zip"
echo
unzip "${DOWNLOAD_DIR}.zip" -d "${DOWNLOAD_DIR}"
echo

echo "🐳 Applying K8s Replicator to cluster (context: $(kubectl config current-context))"
kubectl apply -k "${DOWNLOAD_DIR}/kustomize"
kubectl annotate deployment/replicator -n k8s-replicator "installer.replicator.nadundesilva.github.io/release=${VERSION}"
echo

echo "🔎 Waiting for K8s Replicator to be ready"
kubectl wait deployment/replicator -n k8s-replicator --for condition=available
echo

echo "🏄 Completed! K8s Replicator is ready in the cluster"
