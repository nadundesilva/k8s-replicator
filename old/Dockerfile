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
FROM alpine:3.15.4 AS builder

ARG GROUP=controller
ARG USER=replicator
ARG GID=10500
ARG UID=10500

# hadolint ignore=DL3018
RUN apk update && \
    apk add --no-cache git ca-certificates && \
    update-ca-certificates && \
    addgroup \
        --gid "${GID}" \
        "${GROUP}" && \
    adduser \
        --ingroup "${GROUP}" \
        --disabled-password \
        --gecos "" \
        --home "/nonexistent" \
        --shell "/sbin/nologin" \
        --no-create-home \
        --uid "${UID}" \
        "${USER}"

FROM scratch

ARG GID
ARG UID

WORKDIR /controller
USER ${UID}:${GID}

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /etc/passwd /etc/passwd
COPY --from=builder /etc/group /etc/group

COPY out/replicator /controller/replicator

ENTRYPOINT ["/controller/replicator"]
