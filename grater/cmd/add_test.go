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
const addUsage = `Adds one or more dependents to be tested. The dependents can be specified as command line arguments or in a .txt file, or both.

Usage:
  grater add [dependents...] [flags]

Examples:

grater add foo/bar bar/foo --file dependents.txt
grater add foo/bar
grater add --file dependents.txt
grater add -f dependents.txt


Flags:
  -f, --file string   path to the dependents file
  -h, --help          help for add`

func TestAdd(t *testing.T) {
	out, err := runCobra(t, "add", "--help")
	require.NoError(t, err)
	assert.Contains(t, out, addUsage)
}

func TestAddCmd(t *testing.T) {
	testDir := t.TempDir()
	t.Chdir(testDir)

	out, err := runCobra(t, "add", "foo/bar", "bar/foo")
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
