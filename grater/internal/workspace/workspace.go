// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

// Package workspace provides utilities for working with the .grater directory.
package workspace

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	dirReadWrite  = os.FileMode(0o755)
	fileReadWrite = os.FileMode(0o644)
)

// Workspace represents a workspace directory.
type Workspace struct {
	dir            string
	dependentsPath string
	dependents     []string
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
	graterDir := filepath.Join(w.dir, ".grater")
	err := os.MkdirAll(graterDir, dirReadWrite)

	if err != nil {
		return fmt.Errorf("failed to create .grater/ directory: %w", err)
	}

	return nil
}

// AddDependents adds dependents to the internal dependents list.
func (w *Workspace) AddDependents(dependents []string) {
	w.dependents = append(w.dependents, dependents...)
}

// GetDependents returns the list of dependents and also saves them to dependents.txt.
func (w *Workspace) GetDependents() ([]string, error) {
	err := w.commitToFile(w.dependents, w.dependentsPath)
	if err != nil {
		return nil, nil
	}

	return w.dependents, nil
}

func (w *Workspace) commitToFile(data []string, path string) error {
	content := strings.Join(data, "\n") + "\n"

	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, fileReadWrite)
	if err != nil {
		return fmt.Errorf("failed to write to %s: %w", path, err)
	}
	defer f.Close()

	_, err = f.WriteString(content)
	if err != nil {
		return fmt.Errorf("failed to write to %s: %w", path, err)
	}

	return nil
}