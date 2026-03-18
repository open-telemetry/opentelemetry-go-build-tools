// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package workspace

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetWorkspace(t *testing.T) {
	testDir := t.TempDir()
	t.Chdir(testDir)

	ws, err := GetWorkspace()
	require.NoError(t, err)
	require.NotNil(t, ws)
}

func TestGetWorkspaceDirAlreadyExist(t *testing.T) {
	testDir := t.TempDir()
	t.Chdir(testDir)

	ws1, err := GetWorkspace()
	require.NoError(t, err)
	require.NotNil(t, ws1)

	ws2, err := GetWorkspace()
	require.NoError(t, err)
	require.NotNil(t, ws2)

	assert.Equal(t, ws1, ws2)
}

func TestGetWorkspaceFails(t *testing.T) {
	var err error

	testDir := t.TempDir()
	t.Chdir(testDir)

	f, err := os.Create(".grater")
	require.NoError(t, err)
	require.NoError(t, f.Close())

	_, err = GetWorkspace()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create .grater/ directory")
}

func TestAddDependent(t *testing.T) {
	testDir := t.TempDir()
	t.Chdir(testDir)

	ws, err := GetWorkspace()
	require.NoError(t, err)

	err = ws.AddDependent("foo/bar")
	require.NoError(t, err)

	content, err := os.ReadFile(ws.dependentsPath)
	require.NoError(t, err)

	assert.Contains(t, string(content), "foo/bar")
}
