// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

// Package addhelper provides utilities for working with dependents in the .grater directory.
package addhelper

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.opentelemetry.io/build-tools/grater/internal/module"
	"go.opentelemetry.io/build-tools/grater/internal/workspace"
)

const fileReadWrite = 0644

func TestAdd(t *testing.T) {
	testDir := t.TempDir()
	t.Chdir(testDir)

	ws, err := workspace.NewWorkspace()
	require.NoError(t, err)

	err = Add(ws, []string{"foo/bar@v1.0.0", "baz/qux"})
	require.NoError(t, err)

	dependents := ws.GetDependents()

	assert.ElementsMatch(t, dependents, []module.Module{
		*module.NewModule("foo/bar", "v1.0.0"),
		*module.NewModule("baz/qux", ""),
	})

	content, err := os.ReadFile(".grater/dependents.json")
	require.NoError(t, err)
	assert.JSONEq(t, `[{"module_name":"bar","module_path":"foo/bar","module_version":"v1.0.0"},{"module_name":"qux","module_path":"baz/qux","module_version":""}]`, string(content))
}

func TestAddFromFile(t *testing.T) {
	testDir := t.TempDir()
	t.Chdir(testDir)

	ws, err := workspace.NewWorkspace()
	require.NoError(t, err)

	err = os.WriteFile("dependents.txt", []byte("foo/bar\nbar/foo@v1.0.0\n"), fileReadWrite)
	require.NoError(t, err)

	err = AddFromFile(ws, "dependents.txt")
	require.NoError(t, err)

	dependents := ws.GetDependents()

	assert.ElementsMatch(t, dependents, []module.Module{
		*module.NewModule("foo/bar", ""),
		*module.NewModule("bar/foo", "v1.0.0"),
	})

	content, err := os.ReadFile(".grater/dependents.json")
	require.NoError(t, err)
	assert.JSONEq(t, `[{"module_name":"bar","module_path":"foo/bar","module_version":""},{"module_name":"foo","module_path":"bar/foo","module_version":"v1.0.0"}]`, string(content))
}

func TestAddFromFileFails(t *testing.T) {
	testDir := t.TempDir()
	t.Chdir(testDir)

	ws, err := workspace.NewWorkspace()
	require.NoError(t, err)

	err = AddFromFile(ws, "non_existent.txt")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to read dependents from file")
}
