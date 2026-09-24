// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package workspace

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.opentelemetry.io/build-tools/grater/internal/module"
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

	dependent := module.NewModule("foo/bar", "")
	ws.AddDependents([]module.Module{*dependent})
	assert.Contains(t, ws.dependents, *dependent)
}

func TestGetDependents(t *testing.T) {
	testDir := t.TempDir()
	t.Chdir(testDir)

	ws, err := NewWorkspace()
	require.NoError(t, err)

	dependent := module.NewModule("foo/bar", "")
	ws.AddDependents([]module.Module{*dependent})

	dependents := ws.GetDependents()
	assert.Contains(t, dependents, *dependent)
}

func TestWriteDependents(t *testing.T) {
	testDir := t.TempDir()
	t.Chdir(testDir)

	ws, err := NewWorkspace()
	require.NoError(t, err)

	dependent := module.NewModule("foo/bar", "")
	ws.AddDependents([]module.Module{*dependent})

	err = ws.WriteDependents()
	require.NoError(t, err)

	content, err := os.ReadFile(ws.dependentsPath)
	require.NoError(t, err)
	assert.JSONEq(t, `[{"module_name":"bar","module_path":"foo/bar","module_version":""}]`, string(content))
}

func TestCommitToFile(t *testing.T) {
	testDir := t.TempDir()
	t.Chdir(testDir)

	ws, err := NewWorkspace()
	require.NoError(t, err)

	err = commitToFile([]byte(`foo/bar`), ws.dependentsPath)
	require.NoError(t, err)

	content, err := os.ReadFile(ws.dependentsPath)
	require.NoError(t, err)
	assert.Contains(t, string(content), "foo/bar")
}

func TestAddReplacements(t *testing.T) {
	testDir := t.TempDir()
	t.Chdir(testDir)

	ws, err := NewWorkspace()
	require.NoError(t, err)

	oldModule := module.NewModule("foo/old", "v1.0.0")
	newModule := module.NewModule("foo/new", "v1.1.0")

	replacement := []module.Module{*oldModule, *newModule}

	ws.AddReplacements([][]module.Module{replacement})

	assert.Contains(t, ws.replacements, replacement)
}

func TestGetReplacements(t *testing.T) {
	testDir := t.TempDir()
	t.Chdir(testDir)

	ws, err := NewWorkspace()
	require.NoError(t, err)

	oldModule := module.NewModule("foo/old", "v1.0.0")
	newModule := module.NewModule("foo/new", "v1.1.0")

	replacement := []module.Module{*oldModule, *newModule}

	ws.AddReplacements([][]module.Module{replacement})

	replacements := ws.GetReplacements()

	assert.Contains(t, replacements, replacement)
}

func TestWriteReplacements(t *testing.T) {
	testDir := t.TempDir()
	t.Chdir(testDir)

	ws, err := NewWorkspace()
	require.NoError(t, err)

	oldModule := module.NewModule("foo/old", "v1.0.0")
	newModule := module.NewModule("foo/new", "v1.1.0")

	replacement := []module.Module{*oldModule, *newModule}

	ws.AddReplacements([][]module.Module{replacement})

	err = ws.WriteReplacements()
	require.NoError(t, err)

	content, err := os.ReadFile(ws.replacementsPath)
	require.NoError(t, err)

	assert.JSONEq(t,
		`[
			[
				{
					"module_name":"old",
					"module_path":"foo/old",
					"module_version":"v1.0.0"
				},
				{
					"module_name":"new",
					"module_path":"foo/new",
					"module_version":"v1.1.0"
				}
			]
		]`,
		string(content),
	)
}

func TestCommitToFileFails(t *testing.T) {
	testDir := t.TempDir()
	t.Chdir(testDir)

	ws, err := NewWorkspace()
	require.NoError(t, err)

	// Create a directory for dependentsPath to fail file creation.
	err = os.MkdirAll(ws.dependentsPath, dirReadWrite)
	require.NoError(t, err)

	err = commitToFile([]byte(`foo/bar`), ws.dependentsPath)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to write")
}
