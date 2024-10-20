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

package controller

import (
	"fmt"
	"os"

	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
)

const (
	groupFqn = "replicator.nadundesilva.github.io"

	namespaceTypeLabelKey          = groupFqn + "/namespace-type"
	namespaceTypeLabelValueManaged = "managed"
	namespaceTypeLabelValueIgnored = "ignored"

	objectTypeLabelKey             = groupFqn + "/object-type"
	objectTypeLabelValueReplicated = "replicated"
	objectTypeLabelValueReplica    = "replica"

	resourceFinalizer = groupFqn + "/finalizer"

	sourceNamespaceAnnotationKey = groupFqn + "/source-namespace"

	SourceObjectCreate = "SourceObjectCreate"
	SourceObjectUpdate = "SourceObjectUpdate"
	SourceObjectDelete = "SourceObjectDelete"
)

var (
	namespaceSelector        labels.Selector
	replicaResourcesSelector labels.Selector

	operatorNamespace = os.Getenv("OPERATOR_NAMESPACE")
)

func init() {
	namespaceSelectorReq, err := labels.NewRequirement(
		namespaceTypeLabelKey,
		selection.NotEquals,
		[]string{namespaceTypeLabelValueIgnored},
	)
	if err != nil {
		panic(fmt.Errorf("failed to initialize namespace selector %+w", err))
	}
	namespaceSelector = labels.NewSelector().Add(*namespaceSelectorReq)

	replicaResourcesSelectorReq, err := labels.NewRequirement(
		objectTypeLabelKey,
		selection.Equals,
		[]string{objectTypeLabelValueReplica},
	)
	if err != nil {
		panic(fmt.Errorf("failed to initialize replica resources selector %+w", err))
	}
	replicaResourcesSelector = labels.NewSelector().Add(*replicaResourcesSelectorReq)
}
