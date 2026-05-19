// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const replaceUsage = `Adds one or more replacements to be tested. The replacements can be specified as command line arguments or in a .txt file, or both.

Usage:
  grater replace [replacements...] [flags]

Examples:

grater replace github.com/foo/bar github.com/foo/bar@v1.0.0
grater replace github.com/foo/bar@v1.0.0 ../local/module
grater replace old/module1 new/module1 old/module2 new/module2
grater replace --file replacements.txt
grater replace -f replacements.txt


Flags:
  -f, --file string   path to the replacements file
  -h, --help          help for replace`

func TestReplace(t *testing.T) {
	out, err := runCobra(t, "replace", "--help")
	require.NoError(t, err)

	assert.Contains(t, out, replaceUsage)
}

func TestReplaceCmd(t *testing.T) {
	testDir := t.TempDir()
	t.Chdir(testDir)

	out, err := runCobra(
		t,
		"replace",
		"foo/bar@v1.0.0",
		"baz/qux@v2.0.0",
		"abc/def",
		"xyz/pqr@v3.0.0",
	)
	require.NoError(t, err)

	assert.Contains(t, out, "Successfully added replacements.")

	err = os.WriteFile(
		"replacements.txt",
		[]byte("foo/bar baz/qux@v1.0.0\nabc/def@v2.0.0 xyz/pqr\n"),
		fileReadWrite,
	)
	require.NoError(t, err)

	out, err = runCobra(t, "replace", "--file", "replacements.txt")
	require.NoError(t, err)

	assert.Contains(t, out, "Successfully added replacements.")

	out, err = runCobra(
		t,
		"replace",
		"foo/bar",
		"baz/qux",
		"abc/def@v1.0.0",
		"xyz/pqr",
		"--file",
		"replacements.txt",
	)
	require.NoError(t, err)

	assert.Contains(t, out, "Successfully added replacements.")
}

func TestReplaceCmdFailsOddArgs(t *testing.T) {
	testDir := t.TempDir()
	t.Chdir(testDir)

	_, err := runCobra(
		t,
		"replace",
		"foo/bar",
		"baz/qux",
		"abc/def",
	)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "replacements must be provided in pairs")
}
