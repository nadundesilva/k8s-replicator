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
package controller

import (
	"bufio"
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/nadundesilva/k8s-replicator/test/utils/cleanup"
	"github.com/nadundesilva/k8s-replicator/test/utils/common"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/e2e-framework/klient/k8s"
	"sigs.k8s.io/e2e-framework/klient/wait"
	"sigs.k8s.io/e2e-framework/klient/wait/conditions"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
)

const (
	defaulControllerNamespace  = "k8s-replicator-system"
	defaultTestNamespacePrefix = "replicator-e2e"
	logLevelArgKey             = "--zap-log-level"
)

var (
	kustomizeDirName = filepath.Join("..", "..", "config", "default")
)

func SetupReplicator(ctx context.Context, t *testing.T, cfg *envconf.Config, options ...Option) context.Context {
	opts := &Options{
		labels:       map[string]string{},
		logVerbosity: 100,
	}
	for _, option := range options {
		option(opts)
	}

	controllerNamespace := envconf.RandomName(defaultTestNamespacePrefix, 32)
	ctx = common.WithControllerNamespace(ctx, controllerNamespace)

	kustomizeDir, err := filepath.Abs(kustomizeDirName)
	if err != nil {
		t.Fatalf("failed to resolve kustomize dir %s: %v", kustomizeDirName, err)
	}
	t.Logf("creating controller artifacts from kustomize dir: %s", kustomizeDir)

	resources, err := buildKustomizeResources(t, kustomizeDir)
	if err != nil {
		t.Fatalf("failed build kustomization: %v", err)
	}

	var controllerDeployment *appsv1.Deployment
	for _, r := range resources {
		resource := r.obj.(k8s.Object)

		updateNamespace := func() {
			if resource.GetNamespace() == defaulControllerNamespace || len(resource.GetNamespace()) == 0 {
				resource.SetNamespace(controllerNamespace)
			}
		}

		updateRoleBindingSubjects := func(original []rbacv1.Subject) []rbacv1.Subject {
			newSubjs := []rbacv1.Subject{}
			for _, subject := range original {
				if subject.Namespace == defaulControllerNamespace {
					subject.Namespace = controllerNamespace
				}
				newSubjs = append(newSubjs, subject)
			}
			return newSubjs
		}

		if deployment, ok := resource.(*appsv1.Deployment); ok {
			if controllerDeployment != nil {
				t.Fatalf("More than one deployment found in the kustomize directory")
			}
			updateNamespace()
			controllerDeployment = deployment

			containerIndex := -1
			for i, c := range deployment.Spec.Template.Spec.Containers {
				if c.Name == "manager" {
					containerIndex = i
				}
			}
			if containerIndex < 0 {
				t.Fatalf("Manager container not found")
			}
			container := deployment.Spec.Template.Spec.Containers[containerIndex]

			container.Image = common.GetControllerImage()
			container.ImagePullPolicy = corev1.PullNever

			foundLogLevelArg := false
			logLevelArg := fmt.Sprintf("%s=%d", logLevelArgKey, opts.logVerbosity)
			for i, arg := range container.Args {
				if strings.HasPrefix(arg, logLevelArgKey) {
					container.Args[i] = logLevelArg
					foundLogLevelArg = true
				}
			}
			if !foundLogLevelArg {
				container.Args = append(container.Args, logLevelArg)
			}

			deployment.Spec.Template.Spec.Containers[containerIndex] = container

			t.Logf("creating controller deployment %s/%s", deployment.GetNamespace(), deployment.GetName())
		} else if svc, ok := resource.(*corev1.Service); ok {
			updateNamespace()
			t.Logf("creating controller service %s/%s", svc.GetNamespace(), svc.GetName())
		} else if cm, ok := resource.(*corev1.ConfigMap); ok {
			updateNamespace()
			t.Logf("creating controller config map %s/%s", cm.GetNamespace(), cm.GetName())
		} else if sa, ok := resource.(*corev1.ServiceAccount); ok {
			updateNamespace()
			t.Logf("creating controller service account %s/%s", sa.GetNamespace(), sa.GetName())
		} else if ns, ok := resource.(*corev1.Namespace); ok {
			t.Logf("creating controller namespace %s", ns.GetName())
			if ns.GetName() == defaulControllerNamespace {
				ns.SetName(controllerNamespace)
				for k, v := range opts.labels {
					ns.GetLabels()[k] = v
				}
			}
			ctx = cleanup.AddControllerObjectToContext(ctx, t, resource)
		} else if role, ok := resource.(*rbacv1.Role); ok {
			updateNamespace()
			t.Logf("creating role %s/%s", role.GetNamespace(), role.GetName())
		} else if rolebinding, ok := resource.(*rbacv1.RoleBinding); ok {
			updateNamespace()
			rolebinding.Subjects = updateRoleBindingSubjects(rolebinding.Subjects)
			t.Logf("creating role binding %s/%s", rolebinding.GetNamespace(), rolebinding.GetName())
		} else if clusterrole, ok := resource.(*rbacv1.ClusterRole); ok {
			ctx = cleanup.AddControllerObjectToContext(ctx, t, resource)
			t.Logf("creating controller cluster role %s", clusterrole.GetName())
		} else if clusterrolebinding, ok := resource.(*rbacv1.ClusterRoleBinding); ok {
			clusterrolebinding.Subjects = updateRoleBindingSubjects(clusterrolebinding.Subjects)
			ctx = cleanup.AddControllerObjectToContext(ctx, t, resource)
			t.Logf("creating controller cluster role binding %s", clusterrolebinding.GetName())
		} else {
			t.Fatalf("unknown resource type found in controller kustomization files: %s", r.kind)
		}

		labels := resource.GetLabels()
		labels["app.kubernetes.io/managed-by"] = "kustomize"

		err = cfg.Client().Resources().Create(ctx, resource)
		if err != nil {
			t.Fatalf("failed to create controller resource of kind %s: %v", r.kind, err)
		}
	}
	t.Logf("created controller in namespace %s", controllerNamespace)
	if controllerDeployment == nil {
		t.Fatal("controller deployment not found in controller kustomize files")
	}

	t.Log("waiting for controller to startup")
	err = wait.For(
		conditions.New(cfg.Client().Resources()).ResourceMatch(controllerDeployment, func(object k8s.Object) bool {
			d := object.(*appsv1.Deployment)
			return d.Status.AvailableReplicas > 0 && d.Status.ReadyReplicas > 0
		}),
		wait.WithTimeout(time.Minute),
		wait.WithImmediate(),
		wait.WithInterval(time.Second*5),
	)
	if err != nil {
		t.Fatalf("failed to wait for controller deployment to be ready: %v", err)
		startStreamingLogs(ctx, t, cfg, controllerDeployment, "manager")
		t.FailNow()
	}
	t.Log("waiting for controller to startup complete")

	controllerLogsWg := startStreamingLogs(ctx, t, cfg, controllerDeployment, "manager")
	ctx = common.WithControllerLogsWaitGroup(ctx, controllerLogsWg)
	return ctx
}

func startStreamingLogs(ctx context.Context, t *testing.T, cfg *envconf.Config, deployment *appsv1.Deployment, container string) *sync.WaitGroup {
	k8sClient, err := kubernetes.NewForConfig(cfg.Client().RESTConfig())
	if err != nil {
		t.Fatalf("failed to create a client-go k8s client using e2e-framework rest config: %v", err)
	}

	labelSelector := labels.FormatLabels(deployment.Spec.Selector.MatchLabels)
	podList, err := k8sClient.CoreV1().Pods(deployment.GetNamespace()).List(ctx, metav1.ListOptions{
		LabelSelector: labelSelector,
	})
	if err != nil {
		t.Fatalf("failed to list pods for deployment %s/%s: %v", deployment.GetNamespace(), deployment.GetName(), err)
	}
	logsWg := &sync.WaitGroup{}
	for _, pod := range podList.Items {
		podName := pod.GetName()

		readLogs := func(logsT *testing.T, previous bool, wg *sync.WaitGroup) {
			req := k8sClient.CoreV1().Pods(deployment.GetNamespace()).GetLogs(pod.GetName(), &corev1.PodLogOptions{
				Follow:    true,
				Previous:  previous,
				Container: container,
			})
			logReader, err := req.Stream(ctx)
			if err != nil {
				logsT.Logf("failed to stream logs from replicator (previous: %t, container: %s): %v", previous, container, err)
				return
			}

			logScanner := bufio.NewScanner(logReader)
			defer func() {
				err = logReader.Close()
				if err != nil {
					logsT.Logf("failed to close logs stream: %v", err)
					return
				}
			}()

			for logScanner.Scan() {
				logsT.Logf("[%s] %s", podName, logScanner.Text())
			}
			err = logScanner.Err()
			if err != nil {
				logsT.Logf("error occurred while reading logs: %v", err)
			}
			if wg != nil {
				wg.Done()
			}
		}
		readLogs(t, true, nil)
		logsWg.Add(1)
		go readLogs(t, false, logsWg)
	}
	return logsWg
}
