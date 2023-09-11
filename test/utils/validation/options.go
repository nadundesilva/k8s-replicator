/*
 * Copyright (c) 2022, Nadun De Silva. All Rights Reserved.
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *   http://www.apache.org/licenses/LICENSE-2.0
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package validation

import "time"

type ReplicationOptions struct {
	objectMatcher        ObjectMatcher
	ignoreedNamespaces   []string
	replicatedNamespaces []string
	timeout              time.Duration
}

type ReplicationOption func(*ReplicationOptions)

func WithObjectMatcher(objectMatcher ObjectMatcher) ReplicationOption {
	return func(options *ReplicationOptions) {
		options.objectMatcher = objectMatcher
	}
}

func WithReplicationIgnoredNamespaces(namespaces ...string) ReplicationOption {
	return func(options *ReplicationOptions) {
		options.ignoreedNamespaces = append(options.ignoreedNamespaces, namespaces...)
	}
}

func WithReplicatedNamespaces(namespaces ...string) ReplicationOption {
	return func(options *ReplicationOptions) {
		options.replicatedNamespaces = append(options.replicatedNamespaces, namespaces...)
	}
}

func WithReplicationTimeout(timeout time.Duration) ReplicationOption {
	return func(options *ReplicationOptions) {
		options.timeout = timeout
	}
}

type DeletionOptions struct {
	ignoreedNamespaces []string
}

type DeletionOption func(*DeletionOptions)

func WithDeletionIgnoredNamespaces(namespaces ...string) DeletionOption {
	return func(options *DeletionOptions) {
		options.ignoreedNamespaces = append(options.ignoreedNamespaces, namespaces...)
	}
}
