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

package github

import (
	"os"
	"path/filepath"
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"

	"go.opentelemetry.io/build-tools/issuegenerator/internal/report"
)

func TestTemplateExpansion(t *testing.T) {
	// Create a reportGenerator and ingest test data instead of hardcoding
	rg := report.NewGenerator(zaptest.NewLogger(t), report.GeneratorConfig{ArtifactsPath: filepath.Join("..", "..", "testdata", "junit")})
	reports := rg.ProcessTestResults()

	// Set up the environment variables
	envVariables := map[string]string{
		githubWorkflow:   "test-ci",
		githubServerURL:  "https://github.com",
		githubOwner:      "test-org",
		githubRepository: "test-repo",
		githubRunID:      "555555",
		githubSHAKey:     "abcde12345",
	}

	// Sort the reports by module name to ensure deterministic order
	sort.Slice(reports, func(i, j int) bool {
		return reports[i].Module < reports[j].Module
	})
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
Commit: abcde12

### Component(s)
` + "package1" + `

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
Commit: abcde12

#### Test Failures
-  ` + "`TestFailure`" + `
` + "```" + `
=== RUN   TestFailure
--- FAIL: TestFailure (0.00s)

` + "```" + `

`,
		},
	}

	require.GreaterOrEqual(t, len(reports), len(tests))
	for i, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := os.Expand(tt.template, templateHelper(envVariables, reports[i]))
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestTrimPathAndComponentName(t *testing.T) {
	tests := []struct {
		name          string
		owner         string
		repo          string
		module        string
		wantModule    string
		wantComponent string
	}{
		{
			name:          "Test contrib host metrics integration path",
			owner:         "open-telemetry",
			repo:          "opentelemetry-collector-contrib",
			module:        "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/hostmetricsreceiver/integration_test.go",
			wantModule:    "receiver/hostmetricsreceiver/integration_test.go",
			wantComponent: "receiver/hostmetrics",
		},
		{
			name:          "Test core otlphttp exporter test path",
			owner:         "open-telemetry",
			repo:          "opentelemetry-collector",
			module:        "github.com/open-telemetry/opentelemetry-collector/exporter/otlphttpexporter/otlp_test.go",
			wantModule:    "exporter/otlphttpexporter/otlp_test.go",
			wantComponent: "exporter/otlphttp",
		},
		{
			name:          "No trimming when component name equals type only",
			owner:         "open-telemetry",
			repo:          "opentelemetry-collector",
			module:        "github.com/open-telemetry/opentelemetry-collector/receiver/receiver/test.go",
			wantModule:    "receiver/receiver/test.go",
			wantComponent: "receiver/receiver",
		},
		{
			name:          "Processor trimming",
			owner:         "open-telemetry",
			repo:          "opentelemetry-collector",
			module:        "github.com/open-telemetry/opentelemetry-collector/processor/filterprocessor/test.go",
			wantModule:    "processor/filterprocessor/test.go",
			wantComponent: "processor/filter",
		},
		{
			name:          "extension trimming",
			owner:         "open-telemetry",
			repo:          "opentelemetry-collector",
			module:        "github.com/open-telemetry/opentelemetry-collector/extension/ackextension/test.go",
			wantModule:    "extension/ackextension/test.go",
			wantComponent: "extension/ack",
		},
		{
			name:          "connector trimming",
			owner:         "open-telemetry",
			repo:          "opentelemetry-collector",
			module:        "github.com/open-telemetry/opentelemetry-collector/connector/kafkaconnector/test.go",
			wantModule:    "connector/kafkaconnector/test.go",
			wantComponent: "connector/kafka",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actualModule := trimModule(tt.owner, tt.repo, tt.module)
			actualComponent := getComponent(actualModule)
			assert.Equal(t, tt.wantModule, actualModule, "owner: %s, repo: %s, module: %s, wantModule: %s", tt.owner, tt.repo, tt.module, tt.wantModule)
			assert.Equal(t, tt.wantComponent, actualComponent, "owner: %s, repo: %s, module: %s, wantComponent: %s", tt.owner, tt.repo, tt.module, tt.wantComponent)
		})
	}
}
