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
	InitialCount int    `json:"initialCount"`
	FinalCount   int    `json:"finalCount"`
	Duration     string `json:"duration"`
}

type ReportItems []ReportItem

func (r ReportItems) Len() int {
	return len(r)
}

func (r ReportItems) Less(i, j int) bool {
	if r[i].InitialCount == r[j].InitialCount {
		return r[i].FinalCount < r[j].FinalCount
	} else {
		return r[i].InitialCount < r[j].InitialCount
	}
}

func (r ReportItems) Swap(i, j int) {
	temp := r[i]
	r[i] = r[j]
	r[j] = temp
}

type Report struct {
	namespace ReportItems
}

func (r Report) export() error {
	sort.Sort(r.namespace)

	formattedJson, err := json.MarshalIndent(r, "", "\t")
	if err != nil {
		return fmt.Errorf("failed to format test report into json: %w", err)
	}
	fmt.Printf("Benchmark Results: %s", formattedJson)

	return r.generateMarkdownReport()
}

func (r Report) generateMarkdownReport() error {
	content := "## K8s Replicator - Benchmark Results\n\n" +
		"### Namespace Creation\n\n" +
		"| Initial Namespace Count | Final Namespace Count | Duration |\n" +
		"| -- | -- | -- |\n"
	for _, reportItem := range r.namespace {
		content += fmt.Sprintf("| %d | %d | %s |\n", reportItem.InitialCount, reportItem.FinalCount, reportItem.Duration)
	}
	content += "\n"

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
