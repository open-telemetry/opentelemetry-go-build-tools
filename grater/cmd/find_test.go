// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const findUsage = `Finds all dependents of a given module from pkg.go.dev and adds them to the workspace.

Usage:
  grater find [module@version] [flags]

Examples:

grater find go.opentelemetry.io/otel@v1.24.0


Flags:
  -h, --help   help for find`

func TestFind(t *testing.T) {
	out, err := runCobra(t, "find", "--help")
	require.NoError(t, err)
	assert.Contains(t, out, findUsage)
}

func TestFindCmd(t *testing.T) {
	testDir := t.TempDir()
	t.Chdir(testDir)

	out, err := runCobra(t, "find", "go.opentelemetry.io/otel@v1.24.0")
	require.NoError(t, err)
	assert.Contains(t, out, "Successfully found and added dependents.")
}

func TestFindCmdInvalidInput(t *testing.T) {
	testDir := t.TempDir()
	t.Chdir(testDir)

	_, err := runCobra(t, "find", "go.opentelemetry.io/otel")
	assert.ErrorContains(t, err, "module@version")
}

func TestFindCmdNoArgs(t *testing.T) {
	_, err := runCobra(t, "find")
	assert.Error(t, err)
}
