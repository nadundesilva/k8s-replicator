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

	"sigs.k8s.io/e2e-framework/klient/k8s"
	"sigs.k8s.io/e2e-framework/klient/k8s/resources"
	"sigs.k8s.io/e2e-framework/klient/wait"
	"sigs.k8s.io/e2e-framework/klient/wait/conditions"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
)

func CleanTestObjects(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
	ctxValue := ctx.Value(testObjectsContextKey)
	if ctxValue != nil {
		deleteObjs := func(object k8s.Object) {
			err := cfg.Client().Resources().Delete(ctx, object.DeepCopyObject().(k8s.Object),
				resources.WithDeletePropagation("Background"))
			if err != nil {
				t.Errorf("failed to delete test object %s: %v", object.GetName(), err)
			}
		}
		waitForDeleteObjs := func(objList k8s.ObjectList) {
			clonedObjList := objList.DeepCopyObject().(k8s.ObjectList)
			err := wait.For(conditions.New(cfg.Client().Resources()).ResourcesDeleted(clonedObjList))
			if err != nil {
				t.Fatalf("failed to wait for objects to delete: %v", err)
			}
		}

		objects := ctxValue.(*testObjects)
		for _, obj := range objects.namespaces.Items {
			deleteObjs(&obj)
		}
		for _, obj := range objects.clusterRoles.Items {
			deleteObjs(&obj)
		}
		for _, obj := range objects.clusterRoleBindings.Items {
			deleteObjs(&obj)
		}
		waitForDeleteObjs(&objects.namespaces)
		waitForDeleteObjs(&objects.clusterRoles)
		waitForDeleteObjs(&objects.clusterRoleBindings)
		ctx = context.WithValue(ctx, testObjectsContextKey, nil)
	}
	return ctx
}