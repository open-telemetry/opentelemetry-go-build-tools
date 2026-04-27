// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

// Package addhelper provides utilities for working with dependents in the .grater directory.
package addhelper

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"go.opentelemetry.io/build-tools/grater/internal/module"
	"go.opentelemetry.io/build-tools/grater/internal/workspace"
)

// Add adds dependents to the workspace.
func Add(ws *workspace.Workspace, data []string) error {
	var dependents []module.Module
	for _, modulePath := range data {
		dependents = append(dependents, *module.NewModule(modulePath))
	}
	ws.AddDependents(dependents)
	return ws.WriteDependents()
}

// AddFromFile reads dependents from a .txt file and adds them to the workspace.
func AddFromFile(ws *workspace.Workspace, path string) error {
	cleanPath := filepath.Clean(path)

	data, err := os.ReadFile(cleanPath)
	if err != nil {
		return fmt.Errorf("failed to read dependents from file: %w", err)
	}

	// TODO: Handle other file formats like json, csv using a switch case.

	var dependents []module.Module
	for _, modulePath := range strings.Fields(string(data)) {
		dependents = append(dependents, *module.NewModule(modulePath))
	}

	ws.AddDependents(dependents)
	return ws.WriteDependents()
}
