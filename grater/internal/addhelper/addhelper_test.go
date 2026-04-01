// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

// Package addhelper provides utilities for working with dependents in the .grater directory.
package addhelper

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

	err = Add(ws, []string{"foo/bar", "baz/qux"})
	require.NoError(t, err)

	dependents := ws.GetDependents()

	assert.ElementsMatch(t, dependents, []dependent.Dependent{
		{ModuleName: "foo/bar"},
		{ModuleName: "baz/qux"},
	})

	content, err := os.ReadFile(".grater/dependents.json")
	require.NoError(t, err)
	assert.JSONEq(t, `[{"module_name":"foo/bar"},{"module_name":"baz/qux"}]`, string(content))
}

func TestAddFromFile(t *testing.T) {
	testDir := t.TempDir()
	t.Chdir(testDir)

	ws, err := workspace.NewWorkspace()
	require.NoError(t, err)

	err = os.WriteFile("dependents.txt", []byte("foo/bar\nbar/foo\n"), fileReadWrite)
	require.NoError(t, err)

	err = AddFromFile(ws, "dependents.txt")
	require.NoError(t, err)

	dependents := ws.GetDependents()

	assert.ElementsMatch(t, dependents, []dependent.Dependent{
		{ModuleName: "foo/bar"},
		{ModuleName: "bar/foo"},
	})

	content, err := os.ReadFile(".grater/dependents.json")
	require.NoError(t, err)
	assert.JSONEq(t, `[{"module_name":"foo/bar"},{"module_name":"bar/foo"}]`, string(content))
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
