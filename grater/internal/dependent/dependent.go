// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

// Package dependent provides utilities for working with dependents in the .grater directory.
package dependent

// Dependent represents a dependent module in the workspace.
type Dependent struct {
	ModuleName string `json:"module_name"`
}
