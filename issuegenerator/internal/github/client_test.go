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
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"testing"

	"github.com/google/go-github/v75/github"
	"github.com/migueleliasweb/go-github-mock/src/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"

	"go.opentelemetry.io/build-tools/issuegenerator/internal/report"
)

const (
	testOwner      = "open-telemetry"
	testRepo       = "opentelemetry-collector-contrib"
	testRepoFull   = "open-telemetry/opentelemetry-collector-contrib"
	testIssueURL   = "https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/123"
	testCommentURL = "https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/123#issuecomment-789"
	testModule     = "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/hostmetricsreceiver"
	testComponent  = "receiver/hostmetricsreceiver"
)

func newTestReport() report.Report {
	return report.Report{
		Module:      testModule,
		FailedTests: map[string]string{"TestFailure": "test failed with error"},
	}
}

func newTestIssue() *github.Issue {
	return &github.Issue{ID: github.Ptr[int64](123), Number: github.Ptr(123), HTMLURL: github.Ptr(testIssueURL)}
}

func newTestClient(t *testing.T, httpClient *http.Client) *Client {
	return &Client{
		logger: zaptest.NewLogger(t),
		client: github.NewClient(httpClient),
		envVariables: map[string]string{
			githubOwner:      testOwner,
			githubRepository: testRepo,
			githubWorkflow:   "test-workflow",
			githubServerURL:  "https://github.com",
			githubRunID:      "12345",
			githubSHAKey:     "abcdef123456",
		},
		cfg: ClientConfig{Labels: []string{"test-label"}},
	}
}

func newMockHTTPClient(t *testing.T, endpoint mock.EndpointPattern, response any, statusCode int) *http.Client {
	if statusCode == 0 {
		return mock.NewMockedHTTPClient(
			mock.WithRequestMatch(endpoint, response),
		)
	}
	return mock.NewMockedHTTPClient(
		mock.WithRequestMatchHandler(
			endpoint,
			http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(statusCode)
				w.Header().Set("Content-Type", "application/json")
				require.NoError(t, json.NewEncoder(w).Encode(response))
			}),
		),
	)
}

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

func TestGetExistingIssue(t *testing.T) {
	tests := []struct {
		name          string
		mockResponse  []*github.Issue
		expectedIssue *github.Issue
	}{
		{
			name:          "no existing issues",
			mockResponse:  []*github.Issue{},
			expectedIssue: nil,
		},
		{
			name:          "single existing issue",
			mockResponse:  []*github.Issue{newTestIssue()},
			expectedIssue: newTestIssue(),
		},
		{
			name: "multiple existing issues",
			mockResponse: []*github.Issue{
				{
					ID:      github.Ptr[int64](1),
					Number:  github.Ptr(123),
					HTMLURL: github.Ptr(testIssueURL),
				},
				newTestIssue(),
			},
			expectedIssue: &github.Issue{
				ID:      github.Ptr[int64](1),
				Number:  github.Ptr(123),
				HTMLURL: github.Ptr(testIssueURL),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockedHTTPClient := newMockHTTPClient(t, mock.GetReposIssuesByOwnerByRepo, tt.mockResponse, 0)
			client := newTestClient(t, mockedHTTPClient)
			result := client.GetExistingIssue(t.Context(), testComponent)
			assert.Equal(t, tt.expectedIssue, result)
		})
	}
}

func TestCreateIssue(t *testing.T) {
	testReport := newTestReport()
	expectedIssue := newTestIssue()
	mockedHTTPClient := newMockHTTPClient(t, mock.PostReposIssuesByOwnerByRepo, expectedIssue, http.StatusCreated)
	client := newTestClient(t, mockedHTTPClient)
	result := client.CreateIssue(t.Context(), testReport)
	require.NotNil(t, result)
	assert.Equal(t, *expectedIssue.Number, *result.Number)
	assert.Equal(t, *expectedIssue.HTMLURL, *result.HTMLURL)
}

func TestCommentOnIssue(t *testing.T) {
	testReport := newTestReport()
	existingIssue := newTestIssue()
	expectedComment := &github.IssueComment{
		ID:      github.Ptr[int64](789),
		HTMLURL: github.Ptr(testCommentURL),
	}
	mockedHTTPClient := newMockHTTPClient(t, mock.PostReposIssuesCommentsByOwnerByRepoByIssueNumber, expectedComment, http.StatusCreated)
	client := newTestClient(t, mockedHTTPClient)
	result := client.CommentOnIssue(t.Context(), testReport, existingIssue)
	assert.Equal(t, expectedComment, result)
}

func TestNewClient(t *testing.T) {
	tests := []struct {
		name        string
		envVars     map[string]string
		expectError bool
	}{
		{
			name: "valid environment variables",
			envVars: map[string]string{
				"GITHUB_REPOSITORY": testRepoFull,
				"GITHUB_ACTION":     "test-workflow",
				"GITHUB_SERVER_URL": "https://github.com",
				"GITHUB_RUN_ID":     "12345",
				"GITHUB_TOKEN":      "test-token",
				"GITHUB_SHA":        "abcdef123456",
			},
		},
		{
			name: "missing required environment variable",
			envVars: map[string]string{
				"GITHUB_REPOSITORY": testRepoFull,
				"GITHUB_ACTION":     "test-workflow",
				// Missing other required vars
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for key, value := range tt.envVars {
				t.Setenv(key, value)
			}

			client, err := NewClient(t.Context(), zaptest.NewLogger(t), ClientConfig{})
			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, client)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, client)
			}
		})
	}
}
