// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

// Package workspace provides utilities for working with the .grater directory.
package workspace

import (
	"fmt"
	"os"
	"path/filepath"
)

const dirReadWrite = os.FileMode(0o755)

// GraterExists checks whether a .grater directory exists in the given path.
func GraterExists(currentDir string) error {
	_, err := os.Stat(filepath.Join(currentDir, ".grater"))
	return err
}

// GraterInit initializes a .grater directory in the given path.
func GraterInit(currentDir string) error {
	var err error

	if GraterExists(currentDir) == nil {
		return nil // If .grater already exists, skip.
	}

	graterDir := filepath.Join(currentDir, ".grater")
	err = os.Mkdir(graterDir, dirReadWrite)

	if err != nil {
		return fmt.Errorf("failed to create .grater/ directory: %w", err)
	}

	return nil
}
