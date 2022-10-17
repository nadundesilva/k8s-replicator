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
package testdata

import (
	"fmt"
	"log"
	"os"
	"regexp"
)

var (
	testResourcesFilterRegexString string
	testResourcesFilterRegex       *regexp.Regexp
)

func init() {
	testResourcesFilterRegexString = os.Getenv("TEST_RESOURCES_FILTER_REGEX")
	if testResourcesFilterRegexString == "" {
		testResourcesFilterRegexString = ".*"
	}
	testResourcesFilterRegexString = fmt.Sprintf("^%s$", testResourcesFilterRegexString)

	var err error
	testResourcesFilterRegex, err = regexp.Compile(testResourcesFilterRegexString)
	if err != nil {
		log.Fatalf("failed to initialize test resources filter: %v", err)
	}
}

func GetFilterRegex() string {
	return testResourcesFilterRegexString
}

func isTested(resource Resource) bool {
	return testResourcesFilterRegex.MatchString(resource.Name)
}
