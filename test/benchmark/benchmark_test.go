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
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/nadundesilva/k8s-replicator/test/utils/common"
	"github.com/nadundesilva/k8s-replicator/test/utils/testdata"
	"k8s.io/klog/v2"
	"sigs.k8s.io/e2e-framework/pkg/env"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/envfuncs"
	"sigs.k8s.io/e2e-framework/support/kind"
)

const (
	clusterLogsDir         = "./cluster-logs"
	controllerLogVerbosity = 1
)

var (
	testenv env.Environment
	report  Report
)

func TestMain(m *testing.M) {
	klog.Infof(
		"running benchmark tests using resource filter: %s and controller image: %s",
		testdata.GetFilterRegex(),
		common.GetControllerImage(),
	)

	report = Report{}

	var err error
	testenv, err = env.NewFromFlags()
	if err != nil {
		klog.Infof("failed to generate benchmark test environment from flags: %+v", err)
	}
	kindClusterName := envconf.RandomName("replicator-benchmark-tests-cluster", 32)

	cleanUpLogs := func(ctx context.Context, c *envconf.Config) (context.Context, error) {
		err := os.RemoveAll(clusterLogsDir)
		if err != nil {
			return nil, err
		}
		return ctx, nil
	}

	clusterProvider := kind.NewProvider()
	k8sVersion := os.Getenv("KUBERNETES_VERSION")
	if len(k8sVersion) > 0 {
		clusterProvider = clusterProvider.WithVersion(k8sVersion)
	}

	testenv.Setup(
		envfuncs.CreateCluster(clusterProvider, kindClusterName),
		envfuncs.LoadDockerImageToCluster(kindClusterName, common.GetControllerImage()),
		cleanUpLogs,
	)

	testenv.Finish(
		envfuncs.ExportClusterLogs(kindClusterName, clusterLogsDir),
		envfuncs.DestroyCluster(kindClusterName),
	)

	exitCode := testenv.Run(m)
	err = report.export()
	if err != nil {
		fmt.Printf("failed to print benchmark results: %v", err)
	}
	os.Exit(exitCode)
}

func getBenchmarkTestData(t *testing.T) testdata.Resource {
	selectedIndex := -1
	testdata := testdata.GenerateResourceTestData()
	for i := range testdata {
		if testdata[i].Name == "Secret" { // Only secrets are tested as of now
			selectedIndex = i
		}
	}
	if selectedIndex < 0 {
		t.Fatalf("Secret test data not found")
	}
	return testdata[selectedIndex]
}
