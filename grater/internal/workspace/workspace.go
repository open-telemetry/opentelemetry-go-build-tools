// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

// Package workspace provides utilities for working with the .grater directory.
package workspace

import (
	"fmt"
	"os"
	"path/filepath"
)

const (
	dirReadWrite  = os.FileMode(0o755)
	fileReadWrite = os.FileMode(0o644)
)

// Workspace represents a workspace directory.
type Workspace struct {
	dir            string
	dependentsPath string
}

// NewWorkspace creates a new Workspace instance.
func NewWorkspace() (*Workspace, error) {
	root, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("failed to get current working directory: %w", err)
	}

	w := &Workspace{
		dir:            root,
		dependentsPath: filepath.Join(root, ".grater", "dependents.txt"),
	}
	err = w.create()
	return w, err
}

// Create initializes a .grater directory in the given path.
func (w *Workspace) create() error {
	var err error

	graterDir := filepath.Join(w.dir, ".grater")
	err = os.MkdirAll(graterDir, dirReadWrite)

	if err != nil {
		return fmt.Errorf("failed to create .grater/ directory: %w", err)
	}

	return nil
}

// AddDependent adds a dependent to the dependents.txt file.
func (w *Workspace) AddDependent(dependent string) error {
	f, err := os.OpenFile(w.dependentsPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, fileReadWrite)
	if err != nil {
		return fmt.Errorf("failed to open dependents.txt: %w", err)
	}
	defer f.Close()

	_, err = fmt.Fprintln(f, dependent)
	return err
}
