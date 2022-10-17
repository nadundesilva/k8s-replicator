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

VERSION="$(kubectl get deployment/replicator -n k8s-replicator -o jsonpath='{.metadata.annotations.installer\.replicator\.nadundesilva\.github\.io/release}')"
echo "✨ Detected K8s Replicator ${VERSION}"

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

echo "🛑 Removing K8s Replicator to cluster (context: $(kubectl config current-context))"
kubectl delete -k "${DOWNLOAD_DIR}/kustomize"
echo

echo "🔎 Waiting for K8s Replicator to be removed"
kubectl wait deployment/replicator -n k8s-replicator --for delete
echo

echo "✋ Completed! K8s Replicator is removed from the cluster"
