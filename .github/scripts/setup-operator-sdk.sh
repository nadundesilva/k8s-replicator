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

ARCH=$(case $(uname -m) in x86_64) echo -n amd64 ;; aarch64) echo -n arm64 ;; *) echo -n "$(uname -m)" ;; esac)
OS=$(uname | awk '{print tolower($0)}')

OPERATOR_SDK_DL_URL=https://github.com/operator-framework/operator-sdk/releases/download/v1.24.0
curl -LO "${OPERATOR_SDK_DL_URL}/operator-sdk_${OS}_${ARCH}"

chmod +x "operator-sdk_${OS}_${ARCH}" && sudo mv "operator-sdk_${OS}_${ARCH}" /usr/local/bin/operator-sdk
