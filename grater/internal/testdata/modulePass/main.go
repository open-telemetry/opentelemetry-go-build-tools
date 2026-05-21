// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

// Package module represents a package with passing implementations for dependents.
package module

// Add method provides an incorrect implementation of add for modulePass.
func Add(a, b int) int {
	return a + b
}