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

echo "🧹 Removing Editors"
kubectl delete -k ./editors/editor-03
kubectl delete -k ./editors/editor-02
kubectl delete -k ./editors/editor-01

echo
echo "🧹 Removing Cert Issuer"
kubectl delete -k ./cert-issuer

echo
echo "🧹 Removing Kubernetes Replicator"
kubectl delete -k ../../config/default

echo
echo "🧹 Removing Cert Manager"
helm delete kr-cert-manager -n kr-cert-manager
kubectl delete ns kr-cert-manager

echo
echo "✋ Completed! Cert Manager example is removed from the cluster"
