// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package report

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/build-tools/grater/internal/container"
	"go.opentelemetry.io/build-tools/grater/internal/module"
	"go.opentelemetry.io/build-tools/grater/internal/workspace"
)

func TestClassifyResult(t *testing.T) {
	assert.Equal(t, "pass", classifyResult(container.ExecuteCommandResponse{ExitCode: 0}, container.ExecuteCommandResponse{ExitCode: 0}))
	assert.Equal(t, "broken", classifyResult(container.ExecuteCommandResponse{ExitCode: 1}, container.ExecuteCommandResponse{ExitCode: 1}))
	assert.Equal(t, "regression", classifyResult(container.ExecuteCommandResponse{ExitCode: 0}, container.ExecuteCommandResponse{ExitCode: 1}))
	assert.Equal(t, "broken", classifyResult(container.ExecuteCommandResponse{ExitCode: 1}, container.ExecuteCommandResponse{ExitCode: 0}))
}

func TestGetReport(t *testing.T) {
	testDir := t.TempDir()
	t.Chdir(testDir)

	ws, err := workspace.NewWorkspace()
	require.NoError(t, err)

	err = GetReport(ws,
		[]module.Module{*module.NewModule("moduleA", ""), *module.NewModule("moduleB", ""), *module.NewModule("moduleC", "")},
		[][]container.ExecuteCommandResponse{
			{{ExitCode: 0, Output: "ok"}, {ExitCode: 0, Output: "ok"}},
			{{ExitCode: 1, Output: "fail"}, {ExitCode: 1, Output: "fail"}},
			{{ExitCode: 0, Output: "ok"}, {ExitCode: 1, Output: "fail"}},
		},
	)
	require.NoError(t, err)

	content, err := os.ReadFile(".grater/report.json")
	require.NoError(t, err)
	assert.JSONEq(t, `[
		{"dependent":"moduleA","status":"pass","base_output":"ok","head_output":"ok"},
		{"dependent":"moduleB","status":"broken","base_output":"fail","head_output":"fail"},
		{"dependent":"moduleC","status":"regression","base_output":"ok","head_output":"fail"}
	]`, string(content))
}

func TestGetRegressionReport(t *testing.T) {
	testDir := t.TempDir()
	t.Chdir(testDir)

	ws, err := workspace.NewWorkspace()
	require.NoError(t, err)

	err = GetRegressionReport(ws,
		[]module.Module{*module.NewModule("moduleA", ""), *module.NewModule("moduleB", ""), *module.NewModule("moduleC", "")},
		[][]container.ExecuteCommandResponse{
			{{ExitCode: 0, Output: "ok"}, {ExitCode: 0, Output: "ok"}},
			{{ExitCode: 1, Output: "fail"}, {ExitCode: 1, Output: "fail"}},
			{{ExitCode: 0, Output: "ok"}, {ExitCode: 1, Output: "fail"}},
		},
	)
	require.NoError(t, err)

	content, err := os.ReadFile(".grater/regression_report.json")
	require.NoError(t, err)
	assert.JSONEq(t, `[
		{"dependent":"moduleC","status":"regression","base_output":"ok","head_output":"fail"}
	]`, string(content))
}