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
package benchmark

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
	report  Report
)

func TestMain(m *testing.M) {
	fmt.Printf("running benchmark tests using controller image: \"%s\"\n", controller.GetImage())
	report = Report{}

	cfg, err := envconf.NewFromFlags()
	if err != nil {
		fmt.Printf("failed to generate benchmark test config from flags: %v", err)
	}
	testenv = env.NewWithConfig(cfg)
	kindClusterName := envconf.RandomName("replicator-benchmark-tests-cluster", 32)

	testenv.Setup(
		envfuncs.CreateKindCluster(kindClusterName),
		envfuncs.LoadDockerImageToCluster(kindClusterName, controller.GetImage()),
	)

	testenv.Finish(
		envfuncs.DestroyKindCluster(kindClusterName),
	)

	exitCode := testenv.Run(m)
	err = report.export()
	if err != nil {
		fmt.Printf("failed to print benchmark results: %v", err)
	}
	os.Exit(exitCode)
}
