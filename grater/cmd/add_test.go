// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const addUsage = `Adds a new dependent to be tested.`

func TestAdd(t *testing.T) {
	var err error
	var out string

	out, err = runCobra(t, "add", "--help")
	require.NoError(t, err)
	assert.Contains(t, out, addUsage)
}

func TestAddCmd(t *testing.T) {
	var err error
	var out string

	testDir := t.TempDir()
	t.Chdir(testDir)

	out, err = runCobra(t, "add", "foo/bar", "bar/foo")
	require.NoError(t, err)
	assert.Contains(t, out, "Successfully added dependents: [foo/bar bar/foo]\n")
}
