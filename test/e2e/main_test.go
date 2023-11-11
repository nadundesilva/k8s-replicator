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
	"context"
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

const clusterLogsDir = "./cluster-logs"

var testenv env.Environment

func TestMain(m *testing.M) {
	klog.Infof(
		"running E2E tests using resource filter: %s and controller image: %s",
		testdata.GetFilterRegex(),
		common.GetControllerImage(),
	)

	var err error
	testenv, err = env.NewFromFlags()
	if err != nil {
		klog.Infof("failed to generate e2e test environment from flags: %+v", err)
	}
	kindClusterName := envconf.RandomName("k8s-replicator-e2e-tests-cluster", 32)

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

	os.Exit(testenv.Run(m))
}
