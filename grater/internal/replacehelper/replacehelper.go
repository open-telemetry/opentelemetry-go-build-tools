// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

// Package replacehelper provides utilities for working with replacements in the .grater directory.
package replacehelper

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"go.opentelemetry.io/build-tools/grater/internal/module"
	"go.opentelemetry.io/build-tools/grater/internal/workspace"
)

// Replace adds replacements to the workspace.
func Replace(ws *workspace.Workspace, data []string) error {
	var replacements [][]module.Module

	for _, moduleData := range data {
		moduleOld, moduleNew, found := strings.Cut(moduleData, " ")
		if !found {
			return fmt.Errorf("invalid replacement format: %s", moduleData)
		}

		var replacement []module.Module

		moduleOldPath, moduleOldVersion, foundOld := strings.Cut(moduleOld, "@")
		if foundOld {
			replacement = append(replacement,
				*module.NewModule(moduleOldPath, moduleOldVersion))
		} else {
			replacement = append(replacement,
				*module.NewModule(moduleOld, ""))
		}

		moduleNewPath, moduleNewVersion, foundNew := strings.Cut(moduleNew, "@")
		if foundNew {
			replacement = append(replacement,
				*module.NewModule(moduleNewPath, moduleNewVersion))
		} else {
			replacement = append(replacement,
				*module.NewModule(moduleNew, ""))
		}

		replacements = append(replacements, replacement)
	}

	ws.AddReplacements(replacements)
	return ws.WriteReplacements()
}

// AddFromFile reads replacements from a .txt file and adds them to the workspace.
func AddFromFile(ws *workspace.Workspace, path string) error {
	cleanPath := filepath.Clean(path)

	data, err := os.ReadFile(cleanPath)
	if err != nil {
		return fmt.Errorf("failed to read replacements from file: %w", err)
	}

	lines := strings.Split(strings.TrimSpace(string(data)), "\n")

	return Replace(ws, lines)
}
