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
package replicator

import "github.com/nadundesilva/k8s-replicator/pkg/replicator/resources"

type ResourceEventHandler struct {
	replicator resources.ResourceReplicator
}

func NewResourcesEventHandler(replicator resources.ResourceReplicator) *ResourceEventHandler {
	return &ResourceEventHandler{
		replicator: replicator,
	}
}

func (h *ResourceEventHandler) OnAdd(obj interface{}) {
	// Handle Add
}

func (h *ResourceEventHandler) OnUpdate(oldObj, newObj interface{}) {
	// Handle Update
}

func (h *ResourceEventHandler) OnDelete(obj interface{}) {
	// Handle Delete
}
