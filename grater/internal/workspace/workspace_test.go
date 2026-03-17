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

func TestGraterExists(t *testing.T) {
	var err error
	testDir := setUpTestDir(t)

	err = GraterExists(testDir)
	assert.True(t, os.IsNotExist(err))

	err = GraterInit(testDir)
	require.NoError(t, err)

	err = GraterExists(testDir)
	assert.NoError(t, err)
}

func TestGraterInitCreatesWorkspace(t *testing.T) {
	var err error
	testDir := setUpTestDir(t)

	err = GraterInit(testDir)
	require.NoError(t, err)

	// Check that the .grater directory was created.
	err = GraterExists(testDir)
	assert.NoError(t, err)
}

func TestGraterInitDirAlreadyExists(t *testing.T) {
	var err error
	testDir := setUpTestDir(t)

	err = GraterInit(testDir)
	require.NoError(t, err)

	// Try to initialize again
	err = GraterInit(testDir)
	assert.NoError(t, err)
}

func TestGraterInitFailsToCreateDir(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Skipping test on Windows because chmod doesn't affect permissions on Windows, so this test won't work.")
	}

	var err error
	testDir := setUpTestDir(t)

	// Remove write permissions from the test directory
	require.NoError(t, os.Chmod(testDir, dirReadOnly))
	err = GraterInit(testDir)
	require.NoError(t, os.Chmod(testDir, dirReadWrite))

	assert.Error(t, err)
}

func setUpTestDir(t *testing.T) string {
	testDir := t.TempDir()
	t.Chdir(testDir)
	return testDir
}
