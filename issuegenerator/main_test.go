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
	"os"
	"testing"

	"github.com/joshdk/go-junit"
	"github.com/stretchr/testify/assert"
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
					SystemErr:  "",
				},
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
		},
		"package2": {
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
					SystemErr: "",
				},
				{
					Name:       "TestSucess",
					Classname:  "package2",
					Duration:   0,
					Status:     "passed",
					Message:    "",
					Properties: map[string]string{"classname": "package2", "name": "TestSucess", "time": "0.000000"},
					SystemOut:  "",
					SystemErr:  "",
				},
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
		},
		"package3.0": {
			Name:       "package3.0",
			Package:    "",
			Properties: map[string]string{"go.version": "go1.23.1 darwin/arm64"},
			Tests: []junit.Test{
				{
					Name:       "TestSuccess",
					Classname:  "package3.0",
					Duration:   0,
					Status:     "passed",
					Message:    "",
					Properties: map[string]string{"classname": "package3.0", "name": "TestSuccess", "time": "0.000000"},
					SystemOut:  "",
					SystemErr:  "",
				},
			},
			SystemOut: "",
			SystemErr: "",
			Totals: junit.Totals{
				Tests:    1,
				Passed:   1,
				Skipped:  0,
				Failed:   0,
				Error:    0,
				Duration: 0,
			},
		},
		"package3.1": {
			Name:       "package3.1",
			Package:    "",
			Properties: map[string]string{"go.version": "go1.23.1 darwin/arm64"},
			Tests: []junit.Test{
				{
					Name:      "TestFailure",
					Classname: "package3.1",
					Duration:  0,
					Status:    "failed",
					Message:   "Failed",
					Error: junit.Error{
						Message: "Failed",
						Type:    "",
						Body:    "=== RUN   TestFailure\n--- FAIL: TestFailure (0.00s)\n",
					}, Properties: map[string]string{"classname": "package3.1", "name": "TestFailure", "time": "0.000000"},
					SystemOut: "",
					SystemErr: "",
				},
			},
			SystemOut: "",
			SystemErr: "",
			Totals: junit.Totals{
				Tests:    1,
				Passed:   0,
				Skipped:  0,
				Failed:   1,
				Error:    0,
				Duration: 0,
			},
		},
	}
	require.Equal(t, expectedTestResults, rg.testSuites)
}

func TestTrimPath(t *testing.T) {
	tests := []struct {
		name       string
		owner      string
		repo       string
		module     string
		wantModule string
	}{
		{
			name:       "Test contrib host metrics integration path",
			owner:      "open-telemetry",
			repo:       "opentelemetry-collector-contrib",
			module:     "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/hostmetricsreceiver/integration_test.go",
			wantModule: "receiver/hostmetricsreceiver/integration_test.go",
		},
		{
			name:       "Test core otlphttp exporter test path",
			owner:      "open-telemetry",
			repo:       "opentelemetry-collector",
			module:     "github.com/open-telemetry/opentelemetry-collector/exporter/otlphttpexporter/otlp_test.go",
			wantModule: "exporter/otlphttpexporter/otlp_test.go",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.wantModule, trimModule(tt.owner, tt.repo, tt.module), "owner: %s, repo: %s, module: %s, wantModule: %s", tt.owner, tt.repo, tt.module, tt.wantModule)
		})
	}
}

func TestTemplateExpansion(t *testing.T) {
	// Create a reportGenerator and ingest test data instead of hardcoding
	rg := newReportGenerator()
	rg.ingestArtifacts("./testdata/junit")
	rg.processTestResults()

	// Set up the environment variables
	rg.envVariables = map[string]string{
		githubWorkflow:   "test-ci",
		githubServerURL:  "https://github.com",
		githubOwner:      "test-org",
		githubRepository: "test-repo",
		githubRunID:      "555555",
	}

	// Hardcoding iterator to 0 to get the first report
	rg.reportIterator = 0

	tests := []struct {
		name     string
		template string
		expected string
	}{
		{
			name:     "issue body template",
			template: issueBodyTemplate,
			expected: `
Auto-generated report for ` + "`test-ci`" + ` job build.

Link to failed build: https://github.com/test-org/test-repo/actions/runs/555555

#### Test Failures
-  ` + "`TestFailure`" + `
` + "```" + `
=== RUN   TestFailure
--- FAIL: TestFailure (0.00s)

` + "```" + `


**Note**: Information about any subsequent build failures that happen while
this issue is open, will be added as comments with more information to this issue.
`,
		},
		{
			name:     "issue comment template",
			template: issueCommentTemplate,
			expected: `
Link to latest failed build: https://github.com/test-org/test-repo/actions/runs/555555

#### Test Failures
-  ` + "`TestFailure`" + `
` + "```" + `
=== RUN   TestFailure
--- FAIL: TestFailure (0.00s)

` + "```" + `

`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := os.Expand(tt.template, rg.templateHelper)
			assert.Equal(t, tt.expected, result)
		})
	}
}
