// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.opentelemetry.io/build-tools/grater/internal/workspace"
)

func TestInitCreatesGraterDir(t *testing.T) {
	var out string
	var err error
	testDir := setUpTestDir(t)

	out, err = runCobra(t, "init")
	assert.NoError(t, err)
	assert.Contains(t, out, "Initialized .grater/ workspace")

	// Check that the .grater directory was actually created.
	err = workspace.GraterExists(testDir)
	assert.NoError(t, err)
}

func TestInitGraterDirAlreadyExists(t *testing.T) {
	var err error
	setUpTestDir(t)

	_, err = runCobra(t, "init")
	require.NoError(t, err)

	// Re-run command to check it does nothing when .grater already exists.
	_, err = runCobra(t, "init")
	assert.NoError(t, err)
}

func TestInitFailsToCreateGraterDir(t *testing.T) {
	var err error
	testDir := setUpTestDir(t)

	// Change permission to ReadOnly to invoke failure to create .grater/ directory.
	require.NoError(t, os.Chmod(testDir, dirReadOnly))
	_, err = runCobra(t, "init")
	require.NoError(t, os.Chmod(testDir, dirReadWrite))

	assert.Contains(t, err.Error(), "failed to create .grater/ directory")
}
