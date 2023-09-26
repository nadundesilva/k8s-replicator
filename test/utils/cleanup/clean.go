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
package cleanup

import (
	"context"
	"testing"
	"time"

	"github.com/nadundesilva/k8s-replicator/test/utils/common"
	"github.com/nadundesilva/k8s-replicator/test/utils/validation"
	"k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/e2e-framework/klient/k8s"
	"sigs.k8s.io/e2e-framework/klient/k8s/resources"
	"sigs.k8s.io/e2e-framework/klient/wait"
	"sigs.k8s.io/e2e-framework/klient/wait/conditions"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
)

func CleanTestObjects(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
	return CleanTestObjectsWithOptions(ctx, t, cfg)
}

func CleanTestObjectsWithOptions(ctx context.Context, t *testing.T, cfg *envconf.Config, options ...CleanupOption) context.Context {
	ctx = cleanObjects(ctx, t, cfg, testObjectsContextKey{}, options...)
	ctx = cleanObjects(ctx, t, cfg, testControllerObjectsContextKey{}, options...)
	common.GetControllerLogsWaitGroup(ctx).Wait()
	return ctx
}

func cleanObjects(ctx context.Context, t *testing.T, cfg *envconf.Config, contextKey any, options ...CleanupOption) context.Context {
	ctxValue := ctx.Value(contextKey)
	if ctxValue != nil {
		opts := &CleanupOptions{
			timeout: time.Minute,
		}
		for _, option := range options {
			option(opts)
		}

		deleteObjs := func(object k8s.Object, objectType string) {
			err := cfg.Client().Resources().Delete(ctx, object.DeepCopyObject().(k8s.Object),
				resources.WithDeletePropagation("Background"))
			if err != nil {
				if errors.IsNotFound(err) {
					t.Logf("test %s object %s already deleted", objectType, object.GetName())
				} else {
					t.Fatalf("failed to delete test object %s: %v", object.GetName(), err)
				}
			} else {
				t.Logf("deleted test %s object %s", objectType, object.GetName())
			}
		}
		waitForDeleteObjs := func(objList k8s.ObjectList, objectType string) {
			clonedObjList := objList.DeepCopyObject().(k8s.ObjectList)
			t.Logf("waiting for test %s objects to delete", objectType)
			err := wait.For(
				conditions.New(cfg.Client().Resources()).ResourcesDeleted(clonedObjList),
				wait.WithTimeout(opts.timeout),
				wait.WithImmediate(),
				wait.WithInterval(time.Second),
			)
			if err != nil {
				t.Fatalf("failed to wait for objects to delete: %v", err)
			}
			t.Logf("waiting for test %s objects to delete complete", objectType)
		}

		objects := ctxValue.(*testObjects)
		for _, obj := range objects.managedObjects {
			deleteObjs(obj, "object")
		}
		for _, obj := range objects.managedObjects {
			clonedObj := obj.DeepCopyObject().(k8s.Object)
			t.Logf("waiting for managed test object %s/%s to delete", clonedObj.GetNamespace(), clonedObj.GetName())
			err := wait.For(
				conditions.New(cfg.Client().Resources(clonedObj.GetNamespace())).ResourceDeleted(clonedObj),
				wait.WithTimeout(opts.timeout),
				wait.WithImmediate(),
				wait.WithInterval(time.Second),
			)
			if err != nil {
				t.Fatalf(
					"failed to wait for managed test object %s/%s with finalizers %v to delete: %v",
					clonedObj.GetNamespace(),
					clonedObj.GetName(),
					clonedObj.GetFinalizers(),
					err,
				)
			}
			t.Logf("waiting for managed test object %s/%s to delete complete", clonedObj.GetNamespace(), clonedObj.GetName())

			t.Logf("waiting for replicas of managed test object %s/%s to delete", clonedObj.GetNamespace(), clonedObj.GetName())
			validation.ValidateResourceDeletion(ctx, t, cfg, clonedObj, validation.WithDeletionPrintStateOnFail(false))
			t.Logf("waiting for replicas of managed test object %s/%s to delete complete", clonedObj.GetNamespace(), clonedObj.GetName())
		}

		for _, obj := range objects.namespaces.Items {
			deleteObjs(&obj, "namespace")
		}
		waitForDeleteObjs(&objects.namespaces, "namespace")

		for _, obj := range objects.clusterRoles.Items {
			deleteObjs(&obj, "cluster role")
		}
		for _, obj := range objects.clusterRoleBindings.Items {
			deleteObjs(&obj, "cluster role binding")
		}
		waitForDeleteObjs(&objects.clusterRoles, "cluster role")
		waitForDeleteObjs(&objects.clusterRoleBindings, "cluster role binding")
		ctx = context.WithValue(ctx, contextKey, nil)
	}
	return ctx
}
