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

ensure_dependencies() {
    missingRequirements=()
    declare -A requirements=(
        [helm]="sudo snap install helm --classic"
    )

    echo "🔎 Making sure all dependencies are installed"
    set +e
    for executable in "${!requirements[@]}"; do
        installation=$(which "${executable}")
        if [ -z "${installation}" ]; then
            echo "❓ ${executable} not found"
            missingRequirements+=("${executable}")
        else
            echo "✅ ${executable} installation detected: ${installation}"
        fi
    done
    set -e
    echo

    if [ "${#missingRequirements[@]}" != "0" ]; then
        echo -n "🤔 Missing dependencies found. Would you like to install them automatically ? (Y/n): "
        read -r shouldInstallRequirements
        shouldInstallRequirements=${shouldInstallRequirements:-"y"}
        shouldInstallRequirements="$(tr "[:upper:]" "[:lower:]" <<< "${shouldInstallRequirements}")"

        if [ "${shouldInstallRequirements}" = "y" ]; then
            echo
            echo "🚜 Installing dependencies"
            for requirement in "${missingRequirements[@]}"; do
                echo "🔨 Installing dependency \"${requirement}\""
                bash -c "${requirements[${requirement}]}"
            done
        elif [ "${shouldInstallRequirements}" == "n" ]; then
            echo "🛑 Exiting since there are missing dependencies. Please install them and retry again"
            exit 1
        else
            echo "💥 Unknown input (${shouldInstallRequirements}). Expected one of \"y\" or \"n\""
        fi
    else
        echo "✅ All dependencies are already available"
    fi
    echo
}

setup_cert_manager() {
    echo "🌟 Installing Cert Manager"
    helm repo add jetstack https://charts.jetstack.io
    helm repo update
    helm upgrade \
        kr-cert-manager jetstack/cert-manager \
        --install \
        --namespace kr-cert-manager \
        --create-namespace \
        --version v1.8.0 \
        --set installCRDs=true
    echo "✅ Installing Cert Manager Complete"
}

ensure_dependencies
setup_cert_manager

kubectl create ns kr-cert-issuer
kubectl apply -n kr-cert-issuer -k cert-issuer
