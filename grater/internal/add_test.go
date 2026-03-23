// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package internal

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAddDependents(t *testing.T) {
	var err error

	testDir := t.TempDir()
	t.Chdir(testDir)

	err = AddDependents([]string{"foo/bar", "bar/foo"})
	require.NoError(t, err)

	content, err := os.ReadFile(".grater/dependents.txt")
	require.NoError(t, err)

	assert.Contains(t, string(content), "foo/bar")
	assert.Contains(t, string(content), "bar/foo")
}
