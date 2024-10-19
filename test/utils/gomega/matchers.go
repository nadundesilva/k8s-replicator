/*
 * Copyright (c) 2023, Nadun De Silva. All Rights Reserved.
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
package matchers

import (
	"github.com/nadundesilva/k8s-replicator/test/utils/validation"
	"github.com/onsi/gomega/matchers"
	"github.com/onsi/gomega/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type beEquivalentToResourceMatcher struct {
	matchers.BeEquivalentToMatcher
	matcher validation.ObjectMatcher
}

func (m *beEquivalentToResourceMatcher) Match(actual interface{}) (success bool, err error) {
	return m.matcher(actual.(client.Object), m.Expected.(client.Object)), nil
}

func BeEquivalentToResource(expected interface{}, matcher validation.ObjectMatcher) types.GomegaMatcher {
	return &beEquivalentToResourceMatcher{
		BeEquivalentToMatcher: matchers.BeEquivalentToMatcher{
			Expected: expected,
		},
		matcher: matcher,
	}
}
