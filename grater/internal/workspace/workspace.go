// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

// Package workspace provides utilities for working with the .grater directory.
package workspace

import (
	"os"
	"path/filepath"
)

// GraterExists checks whether a .grater directory exists in the given path.
func GraterExists(currentDir string) error {
	_, err := os.Stat(filepath.Join(currentDir, ".grater"))
	return err
}
