// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package internal

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"go.opentelemetry.io/build-tools/grater/internal/workspace"
	"go.opentelemetry.io/build-tools/grater/internal/dependent"
)

// AddDependents adds the list of dependents to the dependents.txt file.
func AddDependents(dependents []string) error {
	ws, err := workspace.NewWorkspace()
	if err != nil {
		return err
	}

	var deps []dependent.Dependent
	for _, dep := range dependents {
		deps = append(deps, dependent.Dependent{Dependent: dep})
	}

	ws.AddDependents(deps)

	return nil
}

// AddDependentsFromFile reads the list of dependents from a file and adds them.
func AddDependentsFromFile(path string) error {
	cleanPath := filepath.Clean(path)
	content, err := os.ReadFile(cleanPath)
	if err != nil {
		return fmt.Errorf("failed to read dependents file: %w", err)
	}

	lines := strings.Split(strings.TrimSpace(string(content)), "\n")
	return AddDependents(lines)
}
