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
	"os"
	"sync"
)

const (
	groupFqn = "replicator.nadundesilva.github.io"

	NamespaceTypeLabelKey          = groupFqn + "/namespace-type"
	NamespaceTypeLabelValueIgnored = "ignored"
	NamespaceTypeLabelValueManaged = "managed"

	ObjectTypeLabelKey             = groupFqn + "/object-type"
	ObjectTypeLabelValueReplicated = "replicated"
	ObjectTypeLabelValueReplica    = "replica"

	SourceNamespaceAnnotationKey = groupFqn + "/source-namespace"

	defaultControllerImage = "nadunrds/k8s-replicator:test"
)

var controllerImage = os.Getenv("CONTROLLER_IMAGE")

func GetControllerImage() string {
	if controllerImage == "" {
		controllerImage = defaultControllerImage
	}
	return controllerImage
}

type controllerNamespaceContextKey struct{}

func WithControllerNamespace(ctx context.Context, ns string) context.Context {
	return context.WithValue(ctx, controllerNamespaceContextKey{}, ns)
}

func GetControllerNamespace(ctx context.Context) string {
	return ctx.Value(controllerNamespaceContextKey{}).(string)
}

type controllerLogsWaitGroupContextKey struct{}

func WithControllerLogsWaitGroup(ctx context.Context, wg *sync.WaitGroup) context.Context {
	return context.WithValue(ctx, controllerLogsWaitGroupContextKey{}, wg)
}

func GetControllerLogsWaitGroup(ctx context.Context) *sync.WaitGroup {
	return ctx.Value(controllerLogsWaitGroupContextKey{}).(*sync.WaitGroup)
}
