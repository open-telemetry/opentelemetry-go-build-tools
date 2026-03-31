// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package workspace

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewWorkspace(t *testing.T) {
	testDir := t.TempDir()
	t.Chdir(testDir)

	ws, err := NewWorkspace()
	require.NoError(t, err)
	require.NotNil(t, ws)
}

func TestNewWorkspaceDirAlreadyExist(t *testing.T) {
	testDir := t.TempDir()
	t.Chdir(testDir)

	ws1, err := NewWorkspace()
	require.NoError(t, err)
	require.NotNil(t, ws1)

	ws2, err := NewWorkspace()
	require.NoError(t, err)
	require.NotNil(t, ws2)

	assert.Equal(t, ws1, ws2)
}

func TestNewWorkspaceFails(t *testing.T) {
	testDir := t.TempDir()
	t.Chdir(testDir)

	f, err := os.Create(".grater")
	require.NoError(t, err)
	require.NoError(t, f.Close())

	_, err := NewWorkspace()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create .grater/ directory")
}

func TestAddDependent(t *testing.T) {
	testDir := t.TempDir()
	t.Chdir(testDir)

	ws, err := NewWorkspace()
	require.NoError(t, err)

	err = ws.AddDependent("foo/bar")
	require.NoError(t, err)

	content, err := os.ReadFile(ws.dependentsPath)
	require.NoError(t, err)

	assert.Contains(t, string(content), "foo/bar")
}

func TestAddDependentFails(t *testing.T) {
	testDir := t.TempDir()
	t.Chdir(testDir)

	ws, err := NewWorkspace()
	require.NoError(t, err)

	// Create a directory for dependentsPath to fail file creation.
	err = os.MkdirAll(ws.dependentsPath, dirReadWrite)
	require.NoError(t, err)

	err = ws.AddDependent("foo/bar")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to open dependents.txt")
}

func TestGetDependents(t *testing.T) {
	testDir := t.TempDir()
	t.Chdir(testDir)

	ws, err := NewWorkspace()
	require.NoError(t, err)

	err = ws.AddDependent("foo/bar")
	require.NoError(t, err)

	dependents, err := ws.GetDependents()
	require.NoError(t, err)

	assert.Contains(t, dependents, "foo/bar")
}

func TestGetDependentsFails(t *testing.T) {
	testDir := t.TempDir()
	t.Chdir(testDir)

	ws, err := NewWorkspace()
	require.NoError(t, err)

	// Create a directory for dependentsPath to fail file creation.
	err = os.MkdirAll(ws.dependentsPath, dirReadWrite)
	require.NoError(t, err)

	dependents, err := ws.GetDependents()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to read dependents.txt")
	assert.Nil(t, dependents)
}
