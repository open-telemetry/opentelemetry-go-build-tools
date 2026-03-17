// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"go.opentelemetry.io/build-tools/grater/internal/workspace"
)

func TestInitUsage(t *testing.T) {
	var out string
	var err error

	out, err = runCobra(t, "init", "--help")
	assert.NoError(t, err)
	assert.Contains(t, out, "Initialize .grater/ workspace in the current working directory")
}

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
