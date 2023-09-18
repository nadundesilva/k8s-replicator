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

package controllers

import (
	"fmt"
	"os"

	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

const (
	groupFqn = "replicator.nadundesilva.github.io"

	NamespaceTypeLabelKey          = groupFqn + "/namespace-type"
	NamespaceTypeLabelValueManaged = "managed"
	NamespaceTypeLabelValueIgnored = "ignored"

	ObjectTypeLabelKey             = groupFqn + "/object-type"
	ObjectTypeLabelValueReplicated = "replicated"
	ObjectTypeLabelValueReplica    = "replica"

	resourceFinalizer = groupFqn + "/finalizer"

	SourceNamespaceAnnotationKey = groupFqn + "/source-namespace"
)

var (
	namespaceSelector           labels.Selector
	replicatedResourcesSelector labels.Selector
	replicaResourcesSelector    labels.Selector
	managedResourcesPredicate   predicate.Predicate

	operatorNamespace = os.Getenv("OPERATOR_NAMESPACE")
)

func init() {
	namespaceSelectorReq, err := labels.NewRequirement(
		NamespaceTypeLabelKey,
		selection.NotEquals,
		[]string{NamespaceTypeLabelValueIgnored},
	)
	if err != nil {
		panic(fmt.Errorf("failed to initialize namespace selector %+w", err))
	}
	namespaceSelector = labels.NewSelector().Add(*namespaceSelectorReq)

	replicatedResourcesSelectorReq, err := labels.NewRequirement(
		ObjectTypeLabelKey,
		selection.Equals,
		[]string{ObjectTypeLabelValueReplicated},
	)
	if err != nil {
		panic(fmt.Errorf("failed to initialize replicated resources selector %+w", err))
	}
	replicatedResourcesSelector = labels.NewSelector().Add(*replicatedResourcesSelectorReq)

	replicaResourcesSelectorReq, err := labels.NewRequirement(
		ObjectTypeLabelKey,
		selection.Equals,
		[]string{ObjectTypeLabelValueReplica},
	)
	if err != nil {
		panic(fmt.Errorf("failed to initialize replica resources selector %+w", err))
	}
	replicaResourcesSelector = labels.NewSelector().Add(*replicaResourcesSelectorReq)

	managedResourcesPredicate = predicate.NewPredicateFuncs(func(object client.Object) bool {
		objectType, objectTypeOk := object.GetLabels()[ObjectTypeLabelKey]
		if objectTypeOk && (objectType == ObjectTypeLabelValueReplicated || objectType == ObjectTypeLabelValueReplica) {
			return true
		}
		return controllerutil.ContainsFinalizer(object, resourceFinalizer)
	})
}
