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
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

const (
	NamespaceTypeLabelKey          = "replicator.nadundesilva.github.io/namespace-type"
	NamespaceTypeLabelValueManaged = "managed"
	NamespaceTypeLabelValueIgnored = "ignored"

	ObjectTypeLabelKey          = "replicator.nadundesilva.github.io/object-type"
	ObjectTypeLabelValueSource  = "source"
	ObjectTypeLabelValueReplica = "replica"

	SourceNamespaceAnnotationKey = "replicator.nadundesilva.github.io/source-namespace"
)

var (
	namespaceSelector        labels.Selector
	sourceResourcesPredicate predicate.Predicate

	operatorNamespace = os.Getenv("OPERATOR_NAMESPACE")
)

func init() {
	namespaceSelectorReq, err := labels.NewRequirement(
		NamespaceTypeLabelKey,
		selection.NotEquals,
		[]string{
			NamespaceTypeLabelValueIgnored,
		},
	)
	if err != nil {
		panic(fmt.Errorf("failed to initialize namespace selector %+w", err))
	}
	namespaceSelector = labels.NewSelector().Add(*namespaceSelectorReq)

	sourceResourcesPredicate = predicate.NewPredicateFuncs(func(object client.Object) bool {
		val, ok := object.GetAnnotations()[ObjectTypeLabelKey]
		return ok && val == ObjectTypeLabelValueSource
	})
}
