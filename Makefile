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

all: build

.PHONY: clean
clean:
	rm -f ./out/replicator

.PHONY: build
build: clean
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o ./out/replicator ./cmd/replicator

.PHONY: docker
docker: build
	docker build -t ghcr.io/nadundesilva/k8s-replicator:test .

.PHONY: test
test: test.e2e

.PHONY: test.e2e
test.e2e: docker
	go test -v -race -timeout 20m -covermode=atomic -coverprofile=./coverage.txt ./test/e2e/...
