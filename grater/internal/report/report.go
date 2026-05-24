// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

// Package report holds utilities to generate reports for tests.
package report

import (
	"go.opentelemetry.io/build-tools/grater/internal/container"
	"go.opentelemetry.io/build-tools/grater/internal/module"
	"go.opentelemetry.io/build-tools/grater/internal/workspace"
)

// Result struct holds an instance of a result of a test run.
type Result struct {
	Dependent  string `json:"dependent"`
	Status     string `json:"status"`
	BaseOutput string `json:"base_output"`
	HeadOutput string `json:"head_output"`
}

// NewResult creates a new result.
func NewResult(dependent module.Module, result []container.ExecuteCommandResponse) Result {
	return Result{
		Dependent:  dependent.ModuleName,
		Status:     classifyResult(result[0], result[1]),
		BaseOutput: result[0].Output,
		HeadOutput: result[1].Output,
	}
}

func classifyResult(base, head container.ExecuteCommandResponse) string {
	switch {
	case base.ExitCode == 0 && head.ExitCode == 0:
		return "pass"
	case base.ExitCode != 0:
		return "broken"
	case base.ExitCode == 0 && head.ExitCode != 0:
		return "regression"
	}
	return ""
}

// GetReport generates and writes a report for all test results.
func GetReport(ws *workspace.Workspace, dependents []module.Module, results [][]container.ExecuteCommandResponse) error {
	report := []Result{}
	for i, result := range results {
		report = append(report, NewResult(dependents[i], result))
	}
	return ws.WriteReport(report)
}

// GetRegressionReport generates and writes a report containing only regressions.
func GetRegressionReport(ws *workspace.Workspace, dependents []module.Module, results [][]container.ExecuteCommandResponse) error {
	report := []Result{}
	for i, result := range results {
		if classifyResult(result[0], result[1]) == "regression" {
			report = append(report, NewResult(dependents[i], result))
		}
	}
	return ws.WriteRegressionReport(report)
}