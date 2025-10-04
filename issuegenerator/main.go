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
	"os"

	"go.uber.org/zap"

	"go.opentelemetry.io/build-tools/issuegenerator/internal/github"
	"go.opentelemetry.io/build-tools/issuegenerator/internal/report"
)

func main() {
	// Parse CLI flags
	genCfg := report.GeneratorConfig{}
	flag.StringVar(&genCfg.ArtifactsPath, "path", "", "Path to the directory with test results")
	ghCfg := github.ClientConfig{}
	flag.TextVar(&ghCfg.Labels, "labels", github.CommaSeparatedList{}, "Comma-separated list of labels to add to the created issues")
	flag.Parse()

	// Initialize logger
	logger, err := zap.NewDevelopment()
	if err != nil {
		fmt.Printf("Failed to set up logger: %v", err)
		os.Exit(1)
	}

	if genCfg.ArtifactsPath == "" {
		logger.Fatal("Path to the directory with test results is required")
	}

	// Initialize clients
	rg := report.NewGenerator(logger, genCfg)
	reports := rg.ProcessTestResults()
	ctx := context.Background()
	ghClient, err := github.NewClient(ctx, logger, ghCfg)
	if err != nil {
		logger.Fatal("Failed to create Github client", zap.Error(err))
	}

	// Look for existing open GitHub Issue that resulted from previous
	// failures of this job.
	logger.Info("Searching GitHub for existing Issues")
	for _, report := range reports {
		logger.Info(
			"Processing test results",
			zap.String("module", report.Module),
			zap.Int("failed_tests", len(report.FailedTests)),
		)

		existingIssue := ghClient.GetExistingIssue(ctx, report.Module)
		// CreateIssue/CommentOnIssue will also comment on the related PR.
		if existingIssue == nil {
			// If none exists, create a new GitHub Issue for the failure.
			logger.Info("No existing Issues found, creating a new one.")
			createdIssue := ghClient.CreateIssue(ctx, report)
			logger.Info("New GitHub Issue created", zap.String("html_url", *createdIssue.HTMLURL))
		} else {
			// Otherwise, add a comment to the existing Issue.
			logger.Info(
				"Updating GitHub Issue with latest failure",
				zap.String("html_url", *existingIssue.HTMLURL),
			)
			createdIssueComment := ghClient.CommentOnIssue(ctx, report, existingIssue)
			logger.Info("GitHub Issue updated", zap.String("html_url", *createdIssueComment.HTMLURL))
		}

	}
}
