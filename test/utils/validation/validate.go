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

	"github.com/nadundesilva/k8s-replicator/test/utils/common"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/e2e-framework/klient/k8s"
	"sigs.k8s.io/e2e-framework/klient/wait"
	"sigs.k8s.io/e2e-framework/klient/wait/conditions"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
)

type ObjectMatcher func(sourceObject client.Object, replicaObject client.Object) bool

func ValidateReplication(ctx context.Context, t *testing.T, cfg *envconf.Config,
	sourceObject k8s.Object, objectList k8s.ObjectList, options ...ReplicationOption) {
	opts := &ReplicationOptions{
		printState: true,
		timeout:    time.Minute,
	}
	for _, option := range options {
		option(opts)
	}

	nsList := &corev1.NamespaceList{}
	err := cfg.Client().Resources().List(ctx, nsList)
	if err != nil {
		t.Fatalf("failed to list namespaces: %v", err)
	}

	waitedResources := []string{}
	listItems := []runtime.Object{}
	for _, ns := range nsList.Items {
		var ignored bool
		for _, ignoredNs := range opts.ignoreedNamespaces {
			if ns.GetName() == ignoredNs {
				ignored = true
			}
		}
		if strings.HasPrefix(ns.GetName(), "kube-") || ns.GetName() == common.GetControllerNamespace(ctx) {
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
		waitedResources = append(waitedResources, fmt.Sprintf("%s/%s", clonedObj.GetNamespace(), clonedObj.GetName()))
	}
	err = meta.SetList(objectList, listItems)
	if err != nil {
		t.Fatalf("failed to create list of objects to wait for: %v", err)
	}

	t.Logf("waiting for replicas to be created: %s", waitedResources)
	err = wait.For(
		conditions.New(cfg.Client().Resources()).ResourcesMatch(
			objectList.DeepCopyObject().(k8s.ObjectList),
			func(object k8s.Object) bool {
				matchMap := func(sourceMap map[string]string, replicaMap map[string]string) error {
					for k, v := range sourceMap {
						if k == common.ObjectTypeLabelKey {
							continue
						}
						if value, ok := replicaMap[k]; ok {
							if value != v {
								return fmt.Errorf("source object %s/%s value %s for key %s does not exist in replica",
									common.GetSourceObjectNamespace(ctx).GetName(), sourceObject.GetName(), v, k)
							}
						} else {
							return fmt.Errorf("source object %s/%s key %s does not exist in replica",
								common.GetSourceObjectNamespace(ctx).GetName(), sourceObject.GetName(), k)
						}
					}
					return nil
				}
				err := matchMap(sourceObject.GetLabels(), object.GetLabels())
				if err != nil {
					t.Logf("object %s/%s labels are not matching: %v", object.GetNamespace(), object.GetName(), err)
					return false
				}
				err = matchMap(sourceObject.GetAnnotations(), object.GetAnnotations())
				if err != nil {
					t.Logf("object %s/%s annotations are not matching: %v", object.GetNamespace(), object.GetName(), err)
					return false
				}

				objType, objTypeOk := object.GetLabels()[common.ObjectTypeLabelKey]
				if !objTypeOk {
					t.Logf("object %s/%s does not contain label key %s", object.GetNamespace(), object.GetName(),
						common.ObjectTypeLabelKey)
					return false
				}
				if common.GetSourceObjectNamespace(ctx).GetName() == object.GetNamespace() {
					if objTypeOk && objType != common.ObjectTypeLabelValueReplicated {
						t.Fatalf("object %s/%s label %s does not contain the expected value; want %s, got %s",
							object.GetNamespace(), object.GetName(), common.ObjectTypeLabelKey,
							common.ObjectTypeLabelValueReplicated, objType)
						return false
					}
				} else {
					if objTypeOk && objType != common.ObjectTypeLabelValueReplica {
						t.Fatalf("object %s/%s label %s does not contain the expected value; want %s, got %s",
							object.GetNamespace(), object.GetName(), common.ObjectTypeLabelKey,
							common.ObjectTypeLabelValueReplica, objType)
						return false
					}

					sourceNs, sourceNsOk := object.GetAnnotations()[common.SourceNamespaceAnnotationKey]
					if sourceNsOk {
						if sourceNs != common.GetSourceObjectNamespace(ctx).GetName() {
							t.Logf("object %s/%s annotation %s does not contain the source namespace; want %s, got %s",
								object.GetNamespace(), object.GetName(), common.SourceNamespaceAnnotationKey,
								common.GetSourceObjectNamespace(ctx).GetName(), sourceNs)
							return false
						}
					} else {
						t.Logf("object %s/%s does not contain annotation key %s", object.GetNamespace(), object.GetName(),
							common.SourceNamespaceAnnotationKey)
						return false
					}
				}
				if opts.objectMatcher != nil {
					if !opts.objectMatcher(sourceObject, object) {
						t.Logf("failed matching objects: %v", err)
						return false
					}
				}
				return true
			},
		),
		wait.WithTimeout(opts.timeout),
		wait.WithImmediate(),
		wait.WithInterval(time.Second),
	)
	if err != nil {
		t.Errorf("failed to wait for replicated objects: %v", err)
		if opts.printState {
			err = printState(ctx, t, cfg, sourceObject)
			if err != nil {
				t.Fatalf("failed to print the cluster state after replication validation failure: %v", err)
			}
		}
		t.FailNow()
	}
	t.Log("waiting for replicas to be created complete")
}

func ValidateResourceDeletion(ctx context.Context, t *testing.T, cfg *envconf.Config, sourceObject k8s.Object,
	options ...DeletionOption) {
	opts := &DeletionOptions{
		printState: true,
	}
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

		t.Logf("waiting for object %s/%s to be deleted", clonedObj.GetNamespace(), clonedObj.GetName())
		err := wait.For(
			conditions.New(cfg.Client().Resources(namespace.GetName())).ResourceDeleted(clonedObj),
			wait.WithTimeout(time.Minute),
			wait.WithImmediate(),
			wait.WithInterval(time.Second),
		)
		if err != nil {
			t.Errorf(
				"failed to wait for replicated object %s/%s deletion: %v",
				clonedObj.GetNamespace(),
				clonedObj.GetName(),
				err,
			)
			if opts.printState {
				err = printState(ctx, t, cfg, sourceObject)
				if err != nil {
					t.Fatalf("failed to print the cluster state after deletion validation failure: %v", err)
				}
			}
			t.FailNow()
		}
	}
	t.Log("waiting for objects to be deleted complete")
}
