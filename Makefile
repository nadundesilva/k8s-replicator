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

PROJECT_PKG := github.com/nadundesilva/k8s-replicator
GIT_REVISION := $(shell git rev-parse --verify HEAD)

ifeq ("$(CONTROLLER_IMAGE)", "")
	CONTROLLER_IMAGE=nadunrds/k8s-replicator:$(GIT_REVISION)
endif

GO_LDFLAGS := -X $(PROJECT_PKG)/test/e2e.controllerDockerImage=$(CONTROLLER_IMAGE)

all: build

.PHONY: clean
clean:
	rm -f ./out/replicator

.PHONY: build
build: clean
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o ./out/replicator ./cmd/replicator

.PHONY: docker
docker: build
	docker build -t $(CONTROLLER_IMAGE) .

.PHONY: test
test: test.e2e

.PHONY: test.e2e
ifeq ("$(DISABLE_IMAGE_BUILD)", "true")
test.e2e:
else
test.e2e: docker
endif
	go test -v -ldflags "$(GO_LDFLAGS)" -race -timeout 30m ./test/e2e/...
