// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

// Package module represents a package with failing implementations for dependents.
package module

// Add method provides an incorrect implementation of add for moduleFail.
func Add(a, b int) int {
	return a * b
}