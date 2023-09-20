/*
 * Copyright (c) 2023, Nadun De Silva. All Rights Reserved.
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

package replication

import "sigs.k8s.io/controller-runtime/pkg/client"

type Replicator interface {
	GetKind() string
	EmptyObject() client.Object
	EmptyObjectList() client.ObjectList
	ObjectListToArray(client.ObjectList) []client.Object
	Replicate(sourceObject client.Object, targetObject client.Object)
}

func NewReplicators() []Replicator {
	return []Replicator{
		newSecretReplicator(),
		newConfigMapReplicator(),
		newNetworkPolicyReplicator(),
	}
}
