// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

// Package add provides utilities for working with dependents in the .grater directory.
package add

import (
	"fmt"
	"os"
	"strings"

	"go.opentelemetry.io/build-tools/grater/internal/dependent"
	"go.opentelemetry.io/build-tools/grater/internal/workspace"
)

func Add(ws *workspace.Workspace, data []string) {
	var dependents []dependent.Dependent
	for _, moduleName := range data {
		dependents = append(dependents, dependent.Dependent{ModuleName: moduleName})
	}
	ws.AddDependents(dependents)
}

func AddFromFile(ws *workspace.Workspace, path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read dependents from file: %w", err)
	}

	var dependents []dependent.Dependent
	for _, line := range strings.Fields(string(data)) {
		dependents = append(dependents, dependent.Dependent{ModuleName: line})
	}

	ws.AddDependents(dependents)
	return nil
}
