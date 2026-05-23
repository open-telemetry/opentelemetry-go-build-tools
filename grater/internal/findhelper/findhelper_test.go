// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0
package findhelper

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.opentelemetry.io/build-tools/grater/internal/module"
	"go.opentelemetry.io/build-tools/grater/internal/workspace"
)

func TestFetchDependents(t *testing.T) {
	dependents, err := fetchDependents(*module.NewModule("go.opentelemetry.io/otel", "v1.24.0"))
	require.NoError(t, err)
	assert.NotEmpty(t, dependents)
}

func TestFindDependents(t *testing.T) {
	t.Chdir(t.TempDir())
	ws, err := workspace.NewWorkspace()
	require.NoError(t, err)

	err = FindDependents(ws, "go.opentelemetry.io/otel@v1.24.0")
	require.NoError(t, err)
	assert.NotEmpty(t, ws.GetDependents())
}
