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

// Workspace represents a workspace directory.
type Workspace struct {
	dir string
}

// Init creates a new Workspace instance.
func Init(root string) (*Workspace, error) {
	w := &Workspace{
		dir: root,
	}
	err := w.create()
	return w, err
}

// Create initializes a .grater directory in the given path.
func (w *Workspace) create() error {
	var err error

	graterDir := filepath.Join(w.dir, ".grater")
	err = os.Mkdir(graterDir, dirReadWrite)

	if err != nil {
		return fmt.Errorf("failed to create .grater/ directory: %w", err)
	}

	return nil
}
