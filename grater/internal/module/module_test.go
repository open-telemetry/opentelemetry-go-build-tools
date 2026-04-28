// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

// Package module represents a Go module.
package module

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsRemotePath(t *testing.T) {
	module := NewModule("github.com/foo/bar")
	assert.True(t, module.IsRemotePath())

	module = NewModule("foo/bar")
	assert.False(t, module.IsRemotePath())
}
