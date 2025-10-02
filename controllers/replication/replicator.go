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

// Package replication provides the core interfaces and implementations for replicating
// Kubernetes resources across namespaces. This package contains the extensible
// architecture that allows K8s Replicator to support any Kubernetes resource type
// through the Replicator interface.
//
// The package includes implementations for:
//   - Secrets: Secure credential and configuration replication
//   - ConfigMaps: Configuration data replication
//   - NetworkPolicies: Security policy replication
//
// To add support for a new resource type, implement the Replicator interface
// and register it in the NewReplicators() function.
package replication

import (
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Replicator defines the interface for replicating Kubernetes resources across namespaces.
// This interface provides an extensible architecture that allows easy addition of new resource types
// without modifying the core replication logic.
type Replicator interface {
	// GetKind returns the Kubernetes resource kind that this replicator handles.
	// Examples: "Secret", "ConfigMap", "NetworkPolicy", or any Kubernetes resource type.
	GetKind() string

	// AddToScheme registers the resource type with the Kubernetes scheme for proper
	// serialization and deserialization of the resource objects.
	AddToScheme(scheme *runtime.Scheme) error

	// EmptyObject returns an empty instance of the resource type for use in API operations
	// such as Get, List, Create, Update, and Delete operations.
	EmptyObject() client.Object

	// EmptyObjectList returns an empty list of the resource type for use in list operations
	// when querying multiple resources from the Kubernetes API server.
	EmptyObjectList() client.ObjectList

	// ObjectListToArray converts a client.ObjectList to a slice of client.Object for easier
	// processing and iteration over the list of resources.
	ObjectListToArray(client.ObjectList) []client.Object

	// Replicate copies all the relevant data from the source object to the target object.
	// This method performs ONLY in-memory data copying between Go objects and makes NO
	// Kubernetes API calls. The actual replication API calls (Create/Update/Delete) happen
	// AFTER this method completes, when the controller persists the modified target object
	// to the Kubernetes API server. This separation ensures clean testing, modularity,
	// and allows the replication logic to be independent of API operations.
	Replicate(sourceObject client.Object, targetObject client.Object)
}

// NewReplicators returns a slice of all available replicator implementations.
// This function is used to register all supported resource types with the controller.
// To add support for a new resource type, implement the Replicator interface and
// add the new replicator to this list.
func NewReplicators() []Replicator {
	return []Replicator{
		newSecretReplicator(),
		newConfigMapReplicator(),
		newNetworkPolicyReplicator(),
	}
}
