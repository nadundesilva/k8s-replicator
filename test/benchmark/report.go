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
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sort"
)

const (
	markdownReportPath = "report.md"
)

type ReportItem struct {
	InitialNamespaceCount int    `json:"initialNamespaceCount"`
	NewNamespaceCount     int    `json:"newNamespaceCount"`
	Duration              string `json:"duration"`
}

type ReportItems []ReportItem

func (r ReportItems) Len() int {
	return len(r)
}

func (r ReportItems) Less(i, j int) bool {
	if r[i].InitialNamespaceCount == r[j].InitialNamespaceCount {
		return r[i].NewNamespaceCount < r[j].NewNamespaceCount
	} else {
		return r[i].InitialNamespaceCount < r[j].InitialNamespaceCount
	}
}

func (r ReportItems) Swap(i, j int) {
	temp := r[i]
	r[i] = r[j]
	r[j] = temp
}

type Report struct {
	namespace ReportItems
	resource  ReportItems
}

func (r Report) export() error {
	sort.Sort(r.namespace)
	sort.Sort(r.resource)

	formattedJson, err := json.MarshalIndent(r, "", "\t")
	if err != nil {
		return fmt.Errorf("failed to format test report into json: %w", err)
	}
	fmt.Printf("Benchmark Results: %s", formattedJson)

	return r.generateMarkdownReport()
}

func (r Report) generateMarkdownReport() error {
	content := "# K8s Replicator - Benchmark Results\n\n" +
		"These benchmark tests are performed within GitHub Actions, with the tester as well as the Kind K8s cluster sharing " +
		"the same GitHub action resources. These are only meant as a measure of the relative performance of the Operator over time. " +
		"When you are running the Operator in a Kubernetes cluster with higher resources allocated to the Kube API server, " +
		"you can expect much better performance." +
		"## Namespace Creation\n\n" +
		"This is a benchmark on the duration taken to replicate resources to a set of new namespaces with varying initial and new " +
		"namespaces counts. The initial namespaces are created beforehand and only the time taken to create the new namespaces " +
		"and replicate to them are measured for the benchmark.\n\n" +
		"| Initial Namespace Count | New Namespace Count | Duration |\n" +
		"| -- | -- | -- |\n"
	for _, reportItem := range r.namespace {
		content += fmt.Sprintf("| %d | %d | %s |\n", reportItem.InitialNamespaceCount, reportItem.NewNamespaceCount, reportItem.Duration)
	}
	content += "\n" +
		"## Resource Creation\n\n" +
		"This is a benchmark on replicating a new resource to namespaces with varying namespaces counts. The namespaces are " +
		"created beforehand and only the time to replicate to the new namespaces are measured.\n\n" +
		"| Namespace Count | Duration |\n" +
		"| -- | -- |\n"
	for _, reportItem := range r.resource {
		content += fmt.Sprintf("| %d | %s |\n", reportItem.InitialNamespaceCount, reportItem.Duration)
	}

	err := r.writeToFile(markdownReportPath, []byte(content))
	if err != nil {
		return fmt.Errorf("failed to generate markdown report: %+w", err)
	}
	return nil
}

func (r Report) writeToFile(path string, content []byte) error {
	if _, err := os.Stat(path); !errors.Is(err, os.ErrNotExist) {
		err = os.Remove(path)
		if err != nil {
			return fmt.Errorf("failed to remove previous report: %s", path)
		}
	}

	err := os.WriteFile(path, content, 0644)
	if err != nil {
		return fmt.Errorf("failed to write test report %s into file: %+w", path, err)
	}
	return nil
}
