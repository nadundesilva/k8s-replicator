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
	"fmt"
	"time"
)

type Target int64

const (
	Namespace Target = iota
	Resource
)

func (t Target) String() string {
	switch t {
	case Namespace:
		return "Namespace"
	case Resource:
		return "Resource"
	}
	return "Unknown"
}

type reportItem struct {
	Target       Target        `json:"target"`
	InitialCount int           `json:"initialCount"`
	TestCount    int           `json:"testCount"`
	Duration     time.Duration `json:"duration"`
}

type Report []reportItem

func (r Report) export() error {
	formattedJson, err := json.MarshalIndent(r, "", "\t")
	if err != nil {
		return fmt.Errorf("failed to format object data into json: %w", err)
	}
	fmt.Printf("Benchmark Results: %s", formattedJson)
	return nil
}
