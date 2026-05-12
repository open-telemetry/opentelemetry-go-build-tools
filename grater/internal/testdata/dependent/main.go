// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

// Package dependent represents a dependent module.
package dependent

import (
	"module"
)

// Add method uses Add from the module imported.
func Add(a, b int) int {
	return module.Add(a, b)
}