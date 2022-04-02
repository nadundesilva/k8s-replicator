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

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/nadundesilva/k8s-replicator/pkg/replicator"
	"github.com/nadundesilva/k8s-replicator/test/utils/controller"
	"github.com/nadundesilva/k8s-replicator/test/utils/namespaces"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/e2e-framework/klient/k8s"
	"sigs.k8s.io/e2e-framework/klient/wait"
	"sigs.k8s.io/e2e-framework/klient/wait/conditions"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
)

type ObjectMatcher func(sourceObject k8s.Object, targetObject k8s.Object) bool

func ValidateReplication(ctx context.Context, t *testing.T, cfg *envconf.Config,
	sourceObject k8s.Object, objectList k8s.ObjectList, options ...ReplicationOption) {
	opts := &ReplicationOptions{}
	for _, option := range options {
		option(opts)
	}

	nsList := &corev1.NamespaceList{}
	err := cfg.Client().Resources().List(ctx, nsList)
	if err != nil {
		t.Fatalf("failed to list namespaces: %v", err)
	}

	listItems := []runtime.Object{}
	for _, ns := range nsList.Items {
		var ignored bool
		for _, ignoredNs := range opts.ignoreedNamespaces {
			if ns.GetName() == ignoredNs {
				ignored = true
			}
		}
		if strings.HasPrefix(ns.GetName(), "kube-") || ns.GetName() == controller.GetNamspace(ctx) {
			ignored = true
			for _, replicatedNs := range opts.replicatedNamespaces {
				if ns.GetName() == replicatedNs {
					ignored = false
				}
			}
		}
		if ignored {
			continue
		}

		clonedObj := sourceObject.DeepCopyObject().(k8s.Object)
		clonedObj.SetNamespace(ns.GetName())
		listItems = append(listItems, clonedObj)
	}
	err = meta.SetList(objectList, listItems)
	if err != nil {
		t.Fatalf("failed to create list of objects to wait for: %v", err)
	}

	err = wait.For(conditions.New(cfg.Client().Resources()).ResourcesMatch(
		objectList.DeepCopyObject().(k8s.ObjectList),
		func(object k8s.Object) bool {
			matchMap := func(sourceMap map[string]string, targetMap map[string]string) error {
				for k, v := range sourceMap {
					if k == replicator.ObjectTypeLabelKey {
						continue
					}
					if value, ok := targetMap[k]; ok {
						if value != v {
							return fmt.Errorf("source object %s/%s value %s for key %s does not exist in replica",
								namespaces.GetSource(ctx).GetName(), sourceObject.GetName(), v, k)
						}
					} else {
						return fmt.Errorf("source object %s/%s key %s does not exist in replica",
							namespaces.GetSource(ctx).GetName(), sourceObject.GetName(), k)
					}
				}
				return nil
			}
			err := matchMap(sourceObject.GetLabels(), object.GetLabels())
			if err != nil {
				t.Errorf("object %s/%s labels are not matching %v", object.GetNamespace(), object.GetName(), err)
			}
			err = matchMap(sourceObject.GetAnnotations(), object.GetAnnotations())
			if err != nil {
				t.Errorf("object %s/%s annotations are not matching %v", object.GetNamespace(), object.GetName(), err)
			}

			objType, objTypeOk := object.GetLabels()[replicator.ObjectTypeLabelKey]
			if !objTypeOk {
				t.Errorf("object %s/%s does not contain label key %s", object.GetNamespace(), object.GetName(),
					replicator.ObjectTypeLabelKey)
			}
			if namespaces.GetSource(ctx).GetName() == object.GetNamespace() {
				if objTypeOk && objType != replicator.ObjectTypeLabelValueSource {
					t.Errorf("object %s/%s label %s does not contain the expected value; want %s, got %s",
						object.GetNamespace(), object.GetName(), replicator.ObjectTypeLabelKey,
						replicator.ObjectTypeLabelValueSource, objType)
				}
			} else {
				if objTypeOk && objType != replicator.ObjectTypeLabelValueReplica {
					t.Errorf("object %s/%s label %s does not contain the expected value; want %s, got %s",
						object.GetNamespace(), object.GetName(), replicator.ObjectTypeLabelKey,
						replicator.ObjectTypeLabelValueReplica, objType)
				}

				sourceNs, sourceNsOk := object.GetLabels()[replicator.SourceNamespaceLabelKey]
				if sourceNsOk {
					if sourceNs != namespaces.GetSource(ctx).GetName() {
						t.Errorf("object %s/%s label %s does not contain the source namespace; want %s, got %s",
							object.GetNamespace(), object.GetName(), replicator.SourceNamespaceLabelKey,
							namespaces.GetSource(ctx).GetName(), sourceNs)
					}
				} else {
					t.Errorf("object %s/%s does not contain label key %s", object.GetNamespace(), object.GetName(),
						replicator.SourceNamespaceLabelKey)
				}
			}
			if opts.objectMatcher != nil {
				opts.objectMatcher(sourceObject, object)
			}
			return true
		}),
		wait.WithTimeout(time.Minute),
	)
	if err != nil {
		t.Fatalf("failed to wait for replicated objects: %v", err)
	}
}

func ValidateResourceDeletion(ctx context.Context, t *testing.T, cfg *envconf.Config, sourceObject k8s.Object,
	options ...DeletionOption) {
	opts := &DeletionOptions{}
	for _, option := range options {
		option(opts)
	}

	nsList := &corev1.NamespaceList{}
	err := cfg.Client().Resources().List(ctx, nsList)
	if err != nil {
		t.Fatalf("failed to list namespaces: %v", err)
	}
	for _, namespace := range nsList.Items {
		var ignored bool
		for _, ignoredNs := range opts.ignoreedNamespaces {
			if namespace.GetName() == ignoredNs {
				ignored = true
			}
		}
		if ignored {
			continue
		}

		clonedObj := sourceObject.DeepCopyObject().(k8s.Object)
		clonedObj.SetNamespace(namespace.GetName())
		err := wait.For(conditions.New(cfg.Client().Resources(namespace.GetName())).ResourceDeleted(clonedObj))
		if err != nil {
			t.Fatalf("failed to wait for replicated objects: %v", err)
		}
	}
}
