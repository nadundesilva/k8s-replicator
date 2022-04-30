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
)

type Target string

const (
	Namespace Target = "Namespace"
	Resource  Target = "Resource"

	reportPath = "report.json"
)

type reportItem struct {
	Target       Target `json:"target"`
	InitialCount int    `json:"initialCount"`
	FinalCount   int    `json:"finalCount"`
	Duration     string `json:"duration"`
}

type Report []reportItem

func (r Report) export() error {
	formattedJson, err := json.MarshalIndent(r, "", "\t")
	if err != nil {
		return fmt.Errorf("failed to format test report into json: %w", err)
	}
	fmt.Printf("Benchmark Results: %s", formattedJson)

	if _, err := os.Stat(reportPath); !errors.Is(err, os.ErrNotExist) {
		err = os.Remove(reportPath)
		if err != nil {
			return fmt.Errorf("failed to remove previous report: %s", reportPath)
		}
	}

	err = os.WriteFile(reportPath, formattedJson, 0644)
	if err != nil {
		return fmt.Errorf("failed to write test report into file: %w", err)
	}
	return nil
}
