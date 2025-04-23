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

// Issuegenerator generates Github issues for failed tests.
package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/google/go-github/github"
	"github.com/joshdk/go-junit"
	"go.uber.org/zap"
	"golang.org/x/oauth2"
)

const (
	// Keys of required environment variables
	githubOwnerAndRepository = "GITHUB_REPOSITORY"
	githubWorkflow           = "GITHUB_ACTION"
	githubAPITokenKey        = "GITHUB_TOKEN" // #nosec G101

	// Variables used to build workflow URL.
	githubServerURL = "GITHUB_SERVER_URL"
	githubRunID     = "GITHUB_RUN_ID"

	issueTitleTemplate = `[${module}]: Report for failed tests on main`
	issueBodyTemplate  = `
Auto-generated report for ${jobName} job build.

Link to failed build: ${linkToBuild}

${failedTests}

**Note**: Information about any subsequent build failures that happen while
this issue is open, will be added as comments with more information to this issue.
`
	issueCommentTemplate = `
Link to latest failed build: ${linkToBuild}

${failedTests}
`
)

type reportGenerator struct {
	ctx          context.Context
	logger       *zap.Logger
	client       *github.Client
	envVariables map[string]string
	testSuites   map[string]junit.Suite

	reports        []report
	reportIterator int
}

type report struct {
	module      string
	failedTests []string
}

func newReportGenerator() *reportGenerator {
	logger, err := zap.NewDevelopment()
	if err != nil {
		fmt.Printf("Failed to set up logger: %v", err)
		os.Exit(1)
	}

	return &reportGenerator{
		ctx:        context.Background(),
		logger:     logger,
		testSuites: make(map[string]junit.Suite),
		reports:    make([]report, 0),
	}
}

func main() {
	pathToArtifacts := flag.String("path", "", "Path to the directory with test results")
	flag.Parse()
	if *pathToArtifacts == "" {
		fmt.Println("Path to the directory with test results is required")
		os.Exit(1)
	}

	rg := newReportGenerator()
	rg.ingestArtifacts(*pathToArtifacts)
	rg.processTestResults()
	rg.initializeGHClient()

	// Look for existing open GitHub Issue that resulted from previous
	// failures of this job.
	rg.logger.Info("Searching GitHub for existing Issues")
	for _, report := range rg.reports {
		rg.logger.Info(
			"Processing test results",
			zap.String("module", report.module),
			zap.Int("failed_tests", len(report.failedTests)),
		)

		existingIssue := rg.getExistingIssue(report.module)
		if existingIssue == nil {
			// If none exists, create a new GitHub Issue for the failure.
			rg.logger.Info("No existing Issues found, creating a new one.")
			createdIssue := rg.createIssue(report)
			rg.logger.Info("New GitHub Issue created", zap.String("html_url", *createdIssue.HTMLURL))
		} else {
			// Otherwise, add a comment to the existing Issue.
			rg.logger.Info(
				"Updating GitHub Issue with latest failure",
				zap.String("html_url", *existingIssue.HTMLURL),
			)
			createdIssueComment := rg.commentOnIssue(existingIssue)
			rg.logger.Info("GitHub Issue updated", zap.String("html_url", *createdIssueComment.HTMLURL))
		}
		rg.reportIterator++
	}
}

func (rg *reportGenerator) ingestArtifacts(pathToArtifacts string) {
	if pathToArtifacts != "" {
		files, err := os.ReadDir(pathToArtifacts)
		if err != nil {
			rg.logger.Warn(
				"Failed to read directory with test results",
				zap.Error(err),
				zap.String("path", pathToArtifacts),
			)
		}

		for _, file := range files {
			if file.IsDir() {
				continue
			}
			rg.logger.Info("Ingesting test reports", zap.String("path", file.Name()))
			suites, err := junit.IngestFile(path.Join(pathToArtifacts, file.Name()))
			if err != nil {
				rg.logger.Fatal(
					"Failed to ingest JUnit xml, omitting test results from report",
					zap.Error(err),
				)
			}

			// We only expect one suite per file.
			rg.testSuites[suites[0].Name] = suites[0]
		}
	}
}

// processTestResults iterates over the test results and matches the module
// with the code owners, creating a report.
func (rg *reportGenerator) processTestResults() {
	for module, suite := range rg.testSuites {
		if suite.Totals.Failed == 0 {
			continue
		}

		r := report{
			module:      module,
			failedTests: make([]string, 0, suite.Totals.Failed),
		}
		for _, t := range suite.Tests {
			if t.Status == junit.StatusFailed {
				r.failedTests = append(r.failedTests, t.Name)
			}
		}
		rg.reports = append(rg.reports, r)
	}
}

func (rg *reportGenerator) initializeGHClient() {
	rg.getRequiredEnv()
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: rg.envVariables[githubAPITokenKey]})
	tc := oauth2.NewClient(rg.ctx, ts)
	rg.client = github.NewClient(tc)
}

// getRequiredEnv loads required environment variables for the main method.
// Some of the environment variables are built-in in Github Actions, whereas others
// need to be configured. See https://docs.github.com/en/actions/writing-workflows/choosing-what-your-workflow-does/store-information-in-variables#default-environment-variables
// for a list of built-in environment variables.
func (rg *reportGenerator) getRequiredEnv() {
	env := map[string]string{}

	// As shown in the docs, the GITHUB_REPOSITORY environment variable is of the form
	// owner/repository.
	// See https://docs.github.com/en/actions/writing-workflows/choosing-what-your-workflow-does/store-information-in-variables#default-environment-variables:~:text=or%20tag.-,GITHUB_REPOSITORY,-The%20owner%20and
	ownerAndRepository := strings.Split(os.Getenv(githubOwnerAndRepository), "/")
	env["githubOwner"] = ownerAndRepository[0]
	env["githubRepository"] = ownerAndRepository[1]
	env[githubWorkflow] = os.Getenv(githubWorkflow)
	env[githubServerURL] = os.Getenv(githubServerURL)
	env[githubRunID] = os.Getenv(githubRunID)
	env[githubAPITokenKey] = os.Getenv(githubAPITokenKey)

	for k, v := range env {
		if v == "" {
			rg.logger.Fatal(
				"Required environment variable not set",
				zap.String("env_var", k),
			)
		}
	}

	rg.envVariables = env
}

func (rg *reportGenerator) templateHelper(param string) string {
	switch param {
	case "jobName":
		return "`" + rg.envVariables[githubWorkflow] + "`"
	case "linkToBuild":
		return fmt.Sprintf("%s/%s/%s/actions/runs/%s", rg.envVariables[githubServerURL], rg.envVariables["githubOwner"], rg.envVariables["githubRepository"], rg.envVariables[githubRunID])
	case "failedTests":
		return rg.reports[rg.reportIterator].getFailedTests()
	default:
		return ""
	}
}

// getExistingIssues gathers an existing GitHub Issue related to previous failures
// of the same module.
func (rg *reportGenerator) getExistingIssue(module string) *github.Issue {
	issues, response, err := rg.client.Issues.ListByRepo(
		rg.ctx,
		rg.envVariables["githubOwner"],
		rg.envVariables["githubRepository"],
		&github.IssueListByRepoOptions{
			State: "open",
		},
	)
	if err != nil {
		rg.logger.Fatal("Failed to search GitHub Issues", zap.Error(err))
	}

	if response.StatusCode != http.StatusOK {
		rg.handleBadResponses(response)
	}

	requiredTitle := strings.Replace(issueTitleTemplate, "${module}", module, 1)
	for _, issue := range issues {
		if *issue.Title == requiredTitle {
			return issue
		}
	}

	return nil
}

// commentOnIssue adds a new comment on an existing GitHub issue with
// information about the latest failure. This method is expected to be
// called only if there's an existing open Issue for the current job.
func (rg *reportGenerator) commentOnIssue(issue *github.Issue) *github.IssueComment {
	body := os.Expand(issueCommentTemplate, rg.templateHelper)

	issueComment, response, err := rg.client.Issues.CreateComment(
		rg.ctx,
		rg.envVariables["githubOwner"],
		rg.envVariables["githubRepository"],
		*issue.Number,
		&github.IssueComment{
			Body: &body,
		},
	)
	if err != nil {
		rg.logger.Fatal("Failed to search GitHub Issues", zap.Error(err))
	}

	if response.StatusCode != http.StatusCreated {
		rg.handleBadResponses(response)
	}

	return issueComment
}

// createIssue creates a new GitHub Issue corresponding to a build failure.
func (rg *reportGenerator) createIssue(r report) *github.Issue {
	title := strings.Replace(issueTitleTemplate, "${module}", r.module, 1)
	body := os.Expand(issueBodyTemplate, rg.templateHelper)

	issue, response, err := rg.client.Issues.Create(
		rg.ctx,
		rg.envVariables["githubOwner"],
		rg.envVariables["githubRepository"],
		&github.IssueRequest{
			Title: &title,
			Body:  &body,
			// TODO: Set Assignees and labels
		})
	if err != nil {
		rg.logger.Fatal("Failed to create GitHub Issue", zap.Error(err))
	}

	if response.StatusCode != http.StatusCreated {
		rg.handleBadResponses(response)
	}

	return issue
}

// getFailedTests returns information about failed tests if available, otherwise
// an empty string.
func (r *report) getFailedTests() string {
	if len(r.failedTests) == 0 {
		return ""
	}

	var sb strings.Builder
	sb.WriteString("#### Test Failures\n")

	for _, s := range r.failedTests {
		sb.WriteString("-  `" + s + "`\n")
	}

	return sb.String()
}

func (rg *reportGenerator) handleBadResponses(response *github.Response) {
	body, _ := io.ReadAll(response.Body)
	rg.logger.Fatal(
		"Unexpected response from GitHub",
		zap.Int("status_code", response.StatusCode),
		zap.String("response", string(body)),
		zap.String("url", response.Request.URL.String()),
	)
}
