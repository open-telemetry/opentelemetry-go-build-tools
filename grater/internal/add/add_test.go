// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

// Package add provides utilities for working with dependents in the .grater directory.
package add

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.opentelemetry.io/build-tools/grater/internal/dependent"
	"go.opentelemetry.io/build-tools/grater/internal/workspace"
)

const fileReadWrite = 0644

func TestAdd(t *testing.T) {
	testDir := t.TempDir()
	t.Chdir(testDir)

	ws, err := workspace.NewWorkspace()
	require.NoError(t, err)

	Add(ws, []string{"foo/bar", "baz/qux"})

	dependents := ws.GetDependents()

	assert.ElementsMatch(t, dependents, []dependent.Dependent{
		{ModuleName: "foo/bar"},
		{ModuleName: "baz/qux"},
	})
}

func TestAddFromFile(t *testing.T) {
	testDir := t.TempDir()
	t.Chdir(testDir)

	ws, err := workspace.NewWorkspace()
	require.NoError(t, err)

	err = os.WriteFile("dependents.json", []byte("foo/bar\nbar/foo\n"), fileReadWrite)
	require.NoError(t, err)

	err = AddFromFile(ws, "dependents.json")
	require.NoError(t, err)

	dependents := ws.GetDependents()

	assert.ElementsMatch(t, dependents, []dependent.Dependent{
		{ModuleName: "foo/bar"},
		{ModuleName: "bar/foo"},
	})
}

func TestAddFromFileFails(t *testing.T) {
	testDir := t.TempDir()
	t.Chdir(testDir)

	ws, err := workspace.NewWorkspace()
	require.NoError(t, err)

	err = AddFromFile(ws, "non_existent.json")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to read dependents from file")
}
