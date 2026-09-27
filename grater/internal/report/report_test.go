// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package report

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"go.opentelemetry.io/build-tools/grater/internal/container"
	"go.opentelemetry.io/build-tools/grater/internal/module"
)

func TestClassifyResult(t *testing.T) {
	assert.Equal(t, "pass", classifyResult(container.ExecuteCommandResponse{ExitCode: 0}, container.ExecuteCommandResponse{ExitCode: 0}))
	assert.Equal(t, "broken", classifyResult(container.ExecuteCommandResponse{ExitCode: 1}, container.ExecuteCommandResponse{ExitCode: 1}))
	assert.Equal(t, "regression", classifyResult(container.ExecuteCommandResponse{ExitCode: 0}, container.ExecuteCommandResponse{ExitCode: 1}))
	assert.Equal(t, "broken", classifyResult(container.ExecuteCommandResponse{ExitCode: 1}, container.ExecuteCommandResponse{ExitCode: 0}))
}

func TestGetReport(t *testing.T) {
	report := GetReport(
		[]module.Module{*module.NewModule("moduleA", ""), *module.NewModule("moduleB", ""), *module.NewModule("moduleC", "")},
		[][]container.ExecuteCommandResponse{
			{{ExitCode: 0, Output: "ok"}, {ExitCode: 0, Output: "ok"}},
			{{ExitCode: 1, Output: "fail"}, {ExitCode: 1, Output: "fail"}},
			{{ExitCode: 0, Output: "ok"}, {ExitCode: 1, Output: "fail"}},
		},
	)
	assert.Equal(t, Result{Dependent: "moduleA", Status: "pass", BaseOutput: "ok", HeadOutput: "ok"}, report[0])
	assert.Equal(t, Result{Dependent: "moduleB", Status: "broken", BaseOutput: "fail", HeadOutput: "fail"}, report[1])
	assert.Equal(t, Result{Dependent: "moduleC", Status: "regression", BaseOutput: "ok", HeadOutput: "fail"}, report[2])
}

func TestGetRegressionReport(t *testing.T) {
	report := GetRegressionReport(
		[]module.Module{*module.NewModule("moduleA", ""), *module.NewModule("moduleB", ""), *module.NewModule("moduleC", "")},
		[][]container.ExecuteCommandResponse{
			{{ExitCode: 0, Output: "ok"}, {ExitCode: 0, Output: "ok"}},
			{{ExitCode: 1, Output: "fail"}, {ExitCode: 1, Output: "fail"}},
			{{ExitCode: 0, Output: "ok"}, {ExitCode: 1, Output: "fail"}},
		},
	)
	assert.Len(t, report, 1)
	assert.Equal(t, Result{Dependent: "moduleC", Status: "regression", BaseOutput: "ok", HeadOutput: "fail"}, report[0])
}
