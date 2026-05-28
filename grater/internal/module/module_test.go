// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

// Package module represents a Go module.
package module

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsRemotePath(t *testing.T) {
	module := NewModule("go.opentelemetry.io/otel", "v1.0.01")
	assert.True(t, module.IsRemotePath())

	module = NewModule("github.com/open-telemetry/opentelemetry-go", "v1.0.01")
	assert.True(t, module.IsRemotePath())

	module = NewModule("foo/bar", "v1.0.01")
	assert.False(t, module.IsRemotePath())
}
