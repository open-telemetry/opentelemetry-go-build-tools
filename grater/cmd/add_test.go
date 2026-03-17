// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"testing"
	"os"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const dirReadWrite = os.FileMode(0o755)

func TestAddCmd(t *testing.T) {
	var err error
	var out string

	testDir := t.TempDir()
	t.Chdir(testDir)

	out, err = runCobra(t, "add", "foo/bar", "bar/foo")
	require.NoError(t, err)
	assert.Contains(t, out, "Successfully added dependents: [foo/bar bar/foo]\n")
}

func TestAddCmdFails(t *testing.T) {
	var err error

	testDir := t.TempDir()
	t.Chdir(testDir)

	err = os.MkdirAll(".grater", dirReadWrite)
	require.NoError(t, err)

    err = os.Mkdir(".grater/dependents.txt", dirReadWrite)
    require.NoError(t, err)

    _, err = runCobra(t, "add", "foo/bar", "bar/foo")
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "failed to add dependent \"foo/bar\"")
}