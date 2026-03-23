// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

const rootUsage = `Grater is a tool to detect regressions introduced in our downstream dependents by our changes.

Usage:
  grater [command]

Available Commands:
  add         Adds a new dependent to be tested.`

func TestRoot(t *testing.T) {
	var out string
	var err error

	out, err = runCobra(t)
	assert.Contains(t, out, rootUsage)
	assert.Empty(t, err)
}
