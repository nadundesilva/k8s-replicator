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

package replication

import (
	"testing"

	"github.com/nadundesilva/k8s-replicator/test/utils/testdata"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestAPIs(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecs(t, "Replicator Suite")
}

var _ = Describe("Object Replication Test Data", func() {
	Context("When creating test data", func() {
		It("Should generate testdata for all replicated resources", func(ctx SpecContext) {
			testResources := testdata.GenerateResourceTestData()
			testResourceNames := []string{}
			for _, r := range testResources {
				testResourceNames = append(testResourceNames, r.Name)
			}

			replicators := NewReplicators()
			replicatorNames := []string{} // interface{} type is used to match HaveExactElements function args
			for _, r := range replicators {
				replicatorNames = append(replicatorNames, r.GetKind())
			}

			Expect(testResourceNames).To(HaveExactElements(replicatorNames))
		})
	})

})
