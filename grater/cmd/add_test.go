// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAddCmd(t *testing.T) {
	var err error
	var out string

	testDir := t.TempDir()
	t.Chdir(testDir)

	out, err = runCobra(t, "add", "foo/bar")
	require.NoError(t, err)
	assert.Contains(t, out, "Successfully added dependent: foo/bar")
}
