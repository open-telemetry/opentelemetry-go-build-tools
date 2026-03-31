// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package internal

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const fileReadWrite = os.FileMode(0o644)

func TestAddDependents(t *testing.T) {
	testDir := t.TempDir()
	t.Chdir(testDir)

	err := AddDependents([]string{"foo/bar", "bar/foo"})
	require.NoError(t, err)

	content, err := os.ReadFile(".grater/dependents.txt")
	require.NoError(t, err)

	assert.Contains(t, string(content), "foo/bar")
	assert.Contains(t, string(content), "bar/foo")
}

func TestAddDependentsFromFile(t *testing.T) {
	testDir := t.TempDir()
	t.Chdir(testDir)

	err := os.WriteFile("deps.txt", []byte("foo/bar\nbar/foo\n"), fileReadWrite)
	require.NoError(t, err)

	err = AddDependentsFromFile("deps.txt")
	require.NoError(t, err)

	content, err := os.ReadFile(".grater/dependents.txt")
	require.NoError(t, err)

	assert.Contains(t, string(content), "foo/bar")
	assert.Contains(t, string(content), "bar/foo")
}

func TestAddDependentsFromFileInvalidFile(t *testing.T) {
	testDir := t.TempDir()
	t.Chdir(testDir)

	// Pass path to a file that does not exist to invoke failure.
	err := AddDependentsFromFile("non_existent.txt")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to read dependents file")
}
