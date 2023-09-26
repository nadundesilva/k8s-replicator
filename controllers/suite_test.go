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

package controllers

import (
	"context"
	"testing"
	"time"

	"github.com/nadundesilva/k8s-replicator/controllers/replication"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	//+kubebuilder:scaffold:imports
)

var (
	cfg       *rest.Config
	k8sClient client.Client
	testEnv   *envtest.Environment
	cancel    context.CancelFunc
)

func TestAPIs(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecs(t, "Controller Suite")
}

var _ = BeforeSuite(func() {
	logf.SetLogger(zap.New(zap.WriteTo(GinkgoWriter), zap.UseDevMode(true)))

	replicators := replication.NewReplicators()
	scheme := runtime.NewScheme()
	for _, replicator := range replicators {
		Expect(replicator.AddToScheme(scheme)).NotTo(HaveOccurred())
	}

	By("bootstrapping test environment")
	testEnv = &envtest.Environment{
		Scheme:                scheme,
		ErrorIfCRDPathMissing: false,
	}

	var err error
	cfg, err = testEnv.Start()
	Expect(err).NotTo(HaveOccurred())
	Expect(cfg).NotTo(BeNil())

	//+kubebuilder:scaffold:scheme

	k8sClient, err = client.New(cfg, client.Options{Scheme: scheme})
	Expect(err).NotTo(HaveOccurred())
	Expect(k8sClient).NotTo(BeNil())

	mgr, err := ctrl.NewManager(cfg, ctrl.Options{
		Scheme: scheme,
		Cache: cache.Options{
			Scheme:                      scheme,
			ReaderFailOnMissingInformer: true,
		},
	})
	Expect(err).ToNot(HaveOccurred())

	for _, replicator := range replicators {
		err = (&ReplicationReconciler{
			Replicator: replicator,
		}).SetupWithManager(mgr)
		Expect(err).ToNot(HaveOccurred())
	}
	err = (&NamespaceReconciler{
		Replicators: replicators,
	}).SetupWithManager(mgr)
	Expect(err).ToNot(HaveOccurred())

	var ctx context.Context
	ctx, cancel = context.WithCancel(ctrl.SetupSignalHandler())
	go func() {
		defer GinkgoRecover()
		err = mgr.Start(ctx)
		Expect(err).ToNot(HaveOccurred(), "failed to run manager")
	}()
})

var _ = AfterSuite(func() {
	By("shutting down the controller")
	cancel()

	By("shutting down the envtest environment")
	gexec.KillAndWait(5 * time.Second)

	var err error
	stopAttemptStartTime := time.Now()
	for true {
		err = testEnv.Stop()
		if err != nil && time.Since(stopAttemptStartTime) < time.Minute {
			time.Sleep(5 * time.Second)
		} else {
			break
		}
	}

	err = testEnv.Stop()
	Expect(err).NotTo(HaveOccurred())
})
