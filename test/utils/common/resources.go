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

package common

import (
	"context"

	corev1 "k8s.io/api/core/v1"
)

type sourceNamespaceContextKey struct{}

func WithSourceObjectNamespace(ctx context.Context, ns *corev1.Namespace) context.Context {
	return context.WithValue(ctx, sourceNamespaceContextKey{}, ns)
}

func GetSourceObjectNamespace(ctx context.Context) *corev1.Namespace {
	return ctx.Value(sourceNamespaceContextKey{}).(*corev1.Namespace)
}
