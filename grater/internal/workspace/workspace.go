// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

// Package workspace provides utilities for working with the .grater directory.
package workspace

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"go.opentelemetry.io/build-tools/grater/internal/module"
	"go.opentelemetry.io/build-tools/grater/internal/report"
)

const (
	dirReadWrite  = os.FileMode(0o755)
	fileReadWrite = os.FileMode(0o644)
)

// Workspace represents a workspace directory.
type Workspace struct {
	dir                  string
	dependentsPath       string
	replacementsPath     string
	reportPath           string
	regressionReportPath string
	dependents           []module.Module
	replacements         [][]module.Module
}

// NewWorkspace creates a new Workspace instance.
func NewWorkspace() (*Workspace, error) {
	root, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("failed to get current working directory: %w", err)
	}
	w := &Workspace{
		dir:                  root,
		dependentsPath:       filepath.Join(root, ".grater", "dependents.json"),
		replacementsPath:     filepath.Join(root, ".grater", "replacements.json"),
		reportPath:           filepath.Join(root, ".grater", "report.json"),
		regressionReportPath: filepath.Join(root, ".grater", "regression_report.json"),
		dependents:           []module.Module{},
		replacements:         [][]module.Module{},
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

// AddDependents adds dependents to the internal list of dependents.
func (w *Workspace) AddDependents(dependents []module.Module) {
	w.dependents = append(w.dependents, dependents...)
}

// GetDependents returns the internal list of dependents.
func (w *Workspace) GetDependents() []module.Module {
	return w.dependents
}

// WriteDependents writes the list of dependents to the dependents.json file.
func (w *Workspace) WriteDependents() error {
	content, err := json.MarshalIndent(w.dependents, "", "  ")
	if err != nil {
		return err
	}
	return commitToFile(content, w.dependentsPath)
}

// AddReplacements adds replacements to the internal list of replacements.
func (w *Workspace) AddReplacements(replacements [][]module.Module) {
	w.replacements = append(w.replacements, replacements...)
}

// GetReplacements returns the internal list of replacements.
func (w *Workspace) GetReplacements() [][]module.Module {
	return w.replacements
}

// WriteReplacements writes the list of replacements to the replacements.json file.
func (w *Workspace) WriteReplacements() error {
	content, err := json.MarshalIndent(w.replacements, "", "  ")
	if err != nil {
		return err
	}
	return commitToFile(content, w.replacementsPath)
}

// WriteReport writes the test report to the report.json file.
func (w *Workspace) WriteReport(results []report.Result) error {
	content, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		return err
	}
	return commitToFile(content, w.reportPath)
}

// WriteRegressionReport writes the regression report to the regression_report.json file.
func (w *Workspace) WriteRegressionReport(results []report.Result) error {
	content, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		return err
	}
	return commitToFile(content, w.regressionReportPath)
}

func commitToFile(content []byte, path string) error {
	cleanPath := filepath.Clean(path)
	f, err := os.OpenFile(cleanPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, fileReadWrite)
	if err != nil {
		return fmt.Errorf("failed to write to %s: %w", cleanPath, err)
	}
	defer f.Close()
	_, err = f.Write(content)
	if err != nil {
		return fmt.Errorf("failed to write to %s: %w", cleanPath, err)
	}
	return nil
}
