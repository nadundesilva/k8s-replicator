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

VERSION ?= $(GIT_REVISION)

ifeq ("$(CONTROLLER_IMAGE)", "")
	CONTROLLER_IMAGE=nadunrds/k8s-replicator:$(VERSION)
endif

GO_LDFLAGS := -w -s
GO_LDFLAGS += -X $(PROJECT_PKG)/pkg/version.buildVersion=$(VERSION)
GO_LDFLAGS += -X $(PROJECT_PKG)/pkg/version.buildGitRevision=$(GIT_REVISION)
GO_LDFLAGS += -X $(PROJECT_PKG)/pkg/version.buildTime=$(shell date +%Y-%m-%dT%H:%M:%S%z)

all: build

.PHONY: clean
clean:
	rm -f ./out/replicator

.PHONY: build
build: clean
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags "$(GO_LDFLAGS)" -o ./out/replicator ./cmd/replicator

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
	CONTROLLER_IMAGE=$(CONTROLLER_IMAGE) go test -v -ldflags "$(GO_LDFLAGS)" -race -timeout 1h ./test/e2e/...
