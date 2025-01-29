// Copyright The OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"testing"

	"github.com/joshdk/go-junit"
	"github.com/stretchr/testify/require"
)

func TestIngestArtifacts(t *testing.T) {
	rg := newReportGenerator()
	rg.ingestArtifacts("./testdata/junit")

	expectedTestResults := map[string]junit.Suite{
		"package1": {
			Name:       "package1",
			Package:    "",
			Properties: map[string]string{"go.version": "go1.23.1 darwin/arm64"},
			Tests: []junit.Test{
				{
					Name:      "TestFailure",
					Classname: "package1",
					Duration:  0,
					Status:    "failed",
					Message:   "Failed",
					Error: junit.Error{
						Message: "Failed",
						Type:    "",
						Body:    "=== RUN   TestFailure\n--- FAIL: TestFailure (0.00s)\n",
					},
					Properties: map[string]string{"classname": "package1", "name": "TestFailure", "time": "0.000000"},
					SystemOut:  "",
					SystemErr:  "",
				},
				{
					Name:       "TestSucess",
					Classname:  "package1",
					Duration:   0,
					Status:     "passed",
					Message:    "",
					Properties: map[string]string{"classname": "package1", "name": "TestSucess", "time": "0.000000"},
					SystemOut:  "",
					SystemErr:  ""},
			},
			SystemOut: "",
			SystemErr: "",
			Totals: junit.Totals{
				Tests:    2,
				Passed:   1,
				Skipped:  0,
				Failed:   1,
				Error:    0,
				Duration: 0,
			},
		}, "package2": {
			Name:       "package2",
			Package:    "",
			Properties: map[string]string{"go.version": "go1.23.1 darwin/arm64"},
			Tests: []junit.Test{
				{
					Name:      "TestFailure",
					Classname: "package2",
					Duration:  0,
					Status:    "failed",
					Message:   "Failed",
					Error: junit.Error{
						Message: "Failed",
						Type:    "",
						Body:    "=== RUN   TestFailure\n--- FAIL: TestFailure (0.00s)\n",
					}, Properties: map[string]string{"classname": "package2", "name": "TestFailure", "time": "0.000000"},
					SystemOut: "",
					SystemErr: ""},
				{
					Name:       "TestSucess",
					Classname:  "package2",
					Duration:   0,
					Status:     "passed",
					Message:    "",
					Properties: map[string]string{"classname": "package2", "name": "TestSucess", "time": "0.000000"},
					SystemOut:  "",
					SystemErr:  ""},
			},
			SystemOut: "",
			SystemErr: "",
			Totals: junit.Totals{
				Tests:    2,
				Passed:   1,
				Skipped:  0,
				Failed:   1,
				Error:    0,
				Duration: 0},
		},
	}
	require.Equal(t, expectedTestResults, rg.testSuites)

}
