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
package testdata

import (
	"encoding/base64"
	"fmt"
	"reflect"

	"github.com/google/uuid"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func generateSecretTestDatum() Resource {
	return process(resourceData{
		Name: "Secret",
		SourceObject: &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name: fmt.Sprintf("test-secret-%s", uuid.New().String()),
				Labels: map[string]string{
					"e2e-tests.k8s-replicator.io/test-label-key": "test-label-value",
				},
				Annotations: map[string]string{
					"e2e-tests.k8s-replicator.io/test-annotation-key": "test-annotation-value",
				},
			},
			Type: corev1.SecretTypeOpaque,
			Data: map[string][]byte{
				"secret-data-item-one-key": []byte(base64.StdEncoding.EncodeToString([]byte("secret-data-item-one-value"))),
			},
		},
		SourceObjectUpdate: &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Labels: map[string]string{
					"e2e-tests.k8s-replicator.io/test-label-key": "test-label-value",
				},
				Annotations: map[string]string{
					"e2e-tests.k8s-replicator.io/test-annotation-key": "test-annotation-value",
				},
			},
			Type: corev1.SecretTypeOpaque,
			Data: map[string][]byte{
				"secret-data-item-two-key": []byte(base64.StdEncoding.EncodeToString([]byte("secret-data-item-two-value"))),
			},
		},
		EmptyObject:     &corev1.Secret{},
		EmptyObjectList: &corev1.SecretList{},
		IsEqual: func(sourceObject client.Object, replicaObject client.Object) bool {
			sourceSecret := sourceObject.(*corev1.Secret)
			replicaSecret := replicaObject.(*corev1.Secret)
			return reflect.DeepEqual(sourceSecret.Type, replicaSecret.Type) &&
				reflect.DeepEqual(sourceSecret.Immutable, replicaSecret.Immutable) &&
				reflect.DeepEqual(sourceSecret.Data, replicaSecret.Data) &&
				reflect.DeepEqual(sourceSecret.StringData, replicaSecret.StringData)
		},
	})
}
