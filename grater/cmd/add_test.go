// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const fileReadWrite = os.FileMode(0o644)
const addUsage = `Adds a new dependent to be tested.

Usage:
  grater add [flags]

Flags:
  -f, --file string   path to the dependents file`

func TestAdd(t *testing.T) {
	var err error
	var out string

	out, err = runCobra(t, "add", "--help")
	require.NoError(t, err)
	assert.Contains(t, out, addUsage)
}

func TestAddCmd(t *testing.T) {
	var err error
	var out string

	testDir := t.TempDir()
	t.Chdir(testDir)

	out, err = runCobra(t, "add", "foo/bar", "bar/foo")
	require.NoError(t, err)
	assert.Contains(t, out, "Successfully added dependents.")

	err = os.WriteFile("deps.txt", []byte("foo/bar/foo\nbar/foo/bar\n"), fileReadWrite)
	require.NoError(t, err)
	out, err = runCobra(t, "add", "--file", "deps.txt")
	require.NoError(t, err)
	assert.Contains(t, out, "Successfully added dependents.")

	out, err = runCobra(t, "add", "foo/bar", "--file", "deps.txt")
	require.NoError(t, err)
	assert.Contains(t, out, "Successfully added dependents.")
}
