# Copyright (c) 2022, Deep Net. All Rights Reserved.
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#   http://www.apache.org/licenses/LICENSE-2.0
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
FROM golang:1.18 as builder

WORKDIR /repo

COPY ./ /repo/
RUN GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o /repo/out/replicator /repo/cmd/replicator

FROM alpine:3.15.1

WORKDIR /controller
ENV HOME=/controller

COPY --from=builder /repo/out/replicator /controller/replicator

ENTRYPOINT ["/controller/replicator"]
