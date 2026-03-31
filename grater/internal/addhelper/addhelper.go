// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

// Package addhelper provides utilities for working with dependents in the .grater directory.
package addhelper

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"go.opentelemetry.io/build-tools/grater/internal/dependent"
	"go.opentelemetry.io/build-tools/grater/internal/workspace"
)

// Add adds dependents to the workspace.
func Add(ws *workspace.Workspace, data []string) {
	var dependents []dependent.Dependent
	for _, moduleName := range data {
		dependents = append(dependents, dependent.Dependent{ModuleName: moduleName})
	}
	ws.AddDependents(dependents)
}

// AddFromFile reads dependents from a .txt file and adds them to the workspace.
func AddFromFile(ws *workspace.Workspace, path string) error {
	cleanPath := filepath.Clean(path)

	data, err := os.ReadFile(cleanPath)
	if err != nil {
		return fmt.Errorf("failed to read dependents from file: %w", err)
	}

	// TODO: Handle other file formats like json, csv using a switch case.

	var dependents []dependent.Dependent
	for _, line := range strings.Fields(string(data)) {
		dependents = append(dependents, dependent.Dependent{ModuleName: line})
	}

	ws.AddDependents(dependents)
	return nil
}
