// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package workspace

import (
	"os"
	"testing"
	"encoding/json"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.opentelemetry.io/build-tools/grater/internal/dependent"
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

	_, err = NewWorkspace()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create .grater/ directory")
}

func TestAddDependents(t *testing.T) {
	testDir := t.TempDir()
	t.Chdir(testDir)

	ws, err := NewWorkspace()
	require.NoError(t, err)

	ws.AddDependents([]dependent.Dependent{{Dependent: "foo/bar"}})
	assert.Contains(t, ws.dependents, dependent.Dependent{Dependent: "foo/bar"})
}

func TestGetDependents(t *testing.T) {
	testDir := t.TempDir()
	t.Chdir(testDir)

	ws, err := NewWorkspace()
	require.NoError(t, err)

	ws.AddDependents([]dependent.Dependent{{Dependent: "foo/bar"}})

	dependents, err := ws.GetDependents()
	require.NoError(t, err)

	assert.Contains(t, dependents, dependent.Dependent{Dependent: "foo/bar"})

	content, err := os.ReadFile(ws.dependentsPath)
	require.NoError(t, err)
	assert.Contains(t, string(content), "foo/bar")
}

func TestCommitToFile(t *testing.T) {
	testDir := t.TempDir()
	t.Chdir(testDir)

	ws, err := NewWorkspace()
	require.NoError(t, err)

	ws.AddDependents([]dependent.Dependent{{Dependent: "foo/bar"}})

	dependentsJSON, err := json.Marshal(ws.dependents)
	require.NoError(t, err)

	err = ws.commitToFile(dependentsJSON, ws.dependentsPath)
	require.NoError(t, err)

	content, err := os.ReadFile(ws.dependentsPath)
	require.NoError(t, err)
	assert.Contains(t, string(content), "foo/bar")
}

func TestCommitToFileFails(t *testing.T) {
	testDir := t.TempDir()
	t.Chdir(testDir)

	ws, err := NewWorkspace()
	require.NoError(t, err)

	ws.AddDependents([]dependent.Dependent{{Dependent: "foo/bar"}})

	dependentsJSON, err := json.Marshal(ws.dependents)
	require.NoError(t, err)

	// Create a directory for dependentsPath to fail file creation.
	err = os.MkdirAll(ws.dependentsPath, dirReadWrite)
	require.NoError(t, err)

	err = ws.commitToFile(dependentsJSON, ws.dependentsPath)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to write")
}
