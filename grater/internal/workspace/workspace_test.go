// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package workspace

import (
	"os"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const dirReadOnly = os.FileMode(0o555)

func TestInit(t *testing.T) {
	testDir := t.TempDir()
	t.Chdir(testDir)

	ws, err := Init(testDir)
	require.NoError(t, err)
	require.NotNil(t, ws)
}

func TestInitDirAlreadyExist(t *testing.T) {
	testDir := t.TempDir()
	t.Chdir(testDir)

	ws1, err := Init(testDir)
	require.NoError(t, err)
	require.NotNil(t, ws1)

	ws2, err := Init(testDir)
	require.NoError(t, err)
	require.NotNil(t, ws2)

	assert.Equal(t, ws1, ws2)
}

func TestInitFails(t *testing.T) {
	var err error
	if runtime.GOOS == "windows" {
		t.Skip("Skipping test on Windows because chmod doesn't affect permissions on Windows, so this test won't work.")
	}

	testDir := t.TempDir()
	t.Chdir(testDir)

	// Set the directory to read-only to invoke failure to create .grater/
	require.NoError(t, os.Chmod(testDir, dirReadOnly))
	_, err = Init(testDir)
	require.NoError(t, os.Chmod(testDir, dirReadWrite))

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create .grater/ directory")
}

func TestAddDependent(t *testing.T) {
	testDir := t.TempDir()
	t.Chdir(testDir)

	ws, err := Init(testDir)
	require.NoError(t, err)

	err = ws.AddDependent("foo/bar")
	require.NoError(t, err)

	content, err := os.ReadFile(ws.dependentsPath)
	require.NoError(t, err)

	assert.Contains(t, string(content), "foo/bar")
}
