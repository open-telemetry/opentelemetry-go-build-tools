// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package module

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAdd(t *testing.T) {
	out := Add(2, 3)
	assert.Equal(t, out, 5)
}