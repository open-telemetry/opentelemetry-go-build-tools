// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package workspace

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetWorkspace(t *testing.T) {
	var err error
	testDir := t.TempDir()
	t.Chdir(testDir)

	// Creates a new Workspace
	ws, err := GetWorkspace()
	assert.NoError(t, err)
	assert.NotNil(t, ws)

	// Hits cache to retrieve Workspace on same path.
	newWs, err := GetWorkspace()
	assert.NoError(t, err)
	assert.NotNil(t, newWs)
	assert.Equal(t, ws, newWs)
}
