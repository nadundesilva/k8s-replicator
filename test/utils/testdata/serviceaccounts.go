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
package testdata

import (
	"fmt"
	"reflect"

	"github.com/google/uuid"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func generateServiceAccountTestDatum() Resource {
	automountServiceAccountToken := true
	automountServiceAccountTokenUpdate := false

	return process(resourceData{
		Name: "ServiceAccount",
		SourceObject: &corev1.ServiceAccount{
			ObjectMeta: metav1.ObjectMeta{
				Name: fmt.Sprintf("test-serviceaccount-%s", uuid.New().String()),
				Labels: map[string]string{
					"e2e-tests.replicator.io/test-label-key": "test-label-value",
				},
				Annotations: map[string]string{
					"e2e-tests.replicator.io/test-annotation-key": "test-annotation-value",
				},
			},
			Secrets: []corev1.ObjectReference{
				{
					Kind:            "Secret",
					Namespace:       "test-namespace",
					Name:            "test-secret-1",
					UID:             "test-secret-1-uid",
					APIVersion:      "v1",
					ResourceVersion: "test-secret-1-resource-version",
					FieldPath:       "test-secret-1-field-path",
				},
			},
			ImagePullSecrets: []corev1.LocalObjectReference{
				{
					Name: "test-image-pull-secret",
				},
			},
			AutomountServiceAccountToken: &automountServiceAccountToken,
		},
		SourceObjectUpdate: &corev1.ServiceAccount{
			ObjectMeta: metav1.ObjectMeta{
				Labels: map[string]string{
					"e2e-tests.replicator.io/test-label-key": "test-label-value",
				},
				Annotations: map[string]string{
					"e2e-tests.replicator.io/test-annotation-key": "test-annotation-value",
				},
			},
			Secrets: []corev1.ObjectReference{
				{
					Kind:            "Secret",
					Namespace:       "test-namespace",
					Name:            "test-secret-2",
					UID:             "test-secret-2-uid",
					APIVersion:      "v1",
					ResourceVersion: "test-secret-2-resource-version",
					FieldPath:       "test-secret-2-field-path",
				},
			},
			ImagePullSecrets: []corev1.LocalObjectReference{
				{
					Name: "test-image-pull-secret-updated",
				},
			},
			AutomountServiceAccountToken: &automountServiceAccountTokenUpdate,
		},
		EmptyObject:     &corev1.ServiceAccount{},
		EmptyObjectList: &corev1.ServiceAccountList{},
		IsEqual: func(sourceObject client.Object, replicaObject client.Object) bool {
			sourceServiceAccount := sourceObject.(*corev1.ServiceAccount)
			replicaServiceAccount := replicaObject.(*corev1.ServiceAccount)

			// Compare only the name field in ObjectReference since Kubernetes API server
			// automatically strips other fields (kind, namespace, uid, apiVersion, resourceVersion, fieldPath)
			// from ServiceAccount secrets ObjectReference when storing them
			sourceSecretNames := make([]string, len(sourceServiceAccount.Secrets))
			for i, secret := range sourceServiceAccount.Secrets {
				sourceSecretNames[i] = secret.Name
			}
			replicaSecretNames := make([]string, len(replicaServiceAccount.Secrets))
			for i, secret := range replicaServiceAccount.Secrets {
				replicaSecretNames[i] = secret.Name
			}

			return reflect.DeepEqual(sourceSecretNames, replicaSecretNames) &&
				reflect.DeepEqual(sourceServiceAccount.ImagePullSecrets, replicaServiceAccount.ImagePullSecrets) &&
				reflect.DeepEqual(sourceServiceAccount.AutomountServiceAccountToken, replicaServiceAccount.AutomountServiceAccountToken)
		},
	})
}
