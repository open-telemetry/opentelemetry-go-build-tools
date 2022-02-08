// Copyright The OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTransform(t *testing.T) {
	tests := []struct {
		testName   string
		inputSlice []string
	}{
		{
			testName: "with items",
			inputSlice: []string{
				"example.com/testA",
				"example.com/testB",
				"example.com/testC",
				"example.com/testD",
				"example.com/testE",
			},
		},
		{
			testName:   "with empty",
			inputSlice: []string{},
		},
	}

	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			actual := transformExclude(test.inputSlice)

			//len must match
			assert.Len(t, actual, len(test.inputSlice))

			//test for existence
			for _, val := range test.inputSlice {
				_, exists := actual[val]
				assert.True(t, exists)
			}
		})
	}
}
