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
package e2e

import (
	"fmt"
	"os"
	"testing"

	"github.com/nadundesilva/k8s-replicator/test/utils/controller"
	"sigs.k8s.io/e2e-framework/pkg/env"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/envfuncs"
)

var (
	testenv env.Environment
)

func TestMain(m *testing.M) {
	fmt.Printf("running E2E tests on controller image: %s\n", controller.GetImage())

	cfg, err := envconf.NewFromFlags()
	if err != nil {
		fmt.Printf("failed to generate e2e test config from flags: %v", err)
	}
	testenv = env.NewWithConfig(cfg)
	kindClusterName := envconf.RandomName("replicator-e2e-tests-cluster", 32)

	testenv.Setup(
		envfuncs.CreateKindCluster(kindClusterName),
		envfuncs.LoadDockerImageToCluster(kindClusterName, controller.GetImage()),
	)

	testenv.Finish(
		envfuncs.DestroyKindCluster(kindClusterName),
	)

	os.Exit(testenv.Run(m))
}
