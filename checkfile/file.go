// Copyright The OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"fmt"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
)

// FileExists returns an error if the given file is missing for at least one
// enabled component. "projectPath" is the absolute path to the root
// of the project to which the components belong. "defaultComponentsFilePath" is
// the path to the file that contains imports to all required components,
// "goModule" is the Go module to which the imports belong, "filename" is the name of the file
// to check existence. This method is intended to be used to verify documentation and metadata.yaml
// in Opentelemetry core and contrib repositories.
func FileExists(projectPath string, relativeComponentsPath string, projectGoModule string, filename string) error {
	defaultComponentsFilePath := filepath.Join(projectPath, relativeComponentsPath)
	_, err := os.Stat(defaultComponentsFilePath)
	if err != nil {
		return fmt.Errorf("failed to load file %s: %w", defaultComponentsFilePath, err)
	}

	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, defaultComponentsFilePath, nil, parser.ImportsOnly)
	if err != nil {
		return fmt.Errorf("failed to load imports: %w", err)
	}

	importPrefixesToCheck := getImportPrefixesToCheck(projectGoModule)

	for _, i := range f.Imports {
		importPath := strings.Trim(i.Path.Value, `"`)

		if isComponentImport(importPath, importPrefixesToCheck) {
			relativeComponentPath := strings.Replace(importPath, projectGoModule, "", 1)
			filePath := filepath.Join(projectPath, relativeComponentPath, filename)
			_, err := os.Stat(filePath)
			if err != nil {
				return fmt.Errorf("%s does not exist at %s, add one", filename, filePath)
			}
		}
	}
	return nil
}

var componentTypes = []string{"extension", "receiver", "processor", "exporter"}

// getImportPrefixesToCheck returns a slice of strings that are relevant import
// prefixes for components in the given module.
func getImportPrefixesToCheck(module string) []string {
	out := make([]string, len(componentTypes))
	for i, typ := range componentTypes {
		out[i] = strings.Join([]string{strings.TrimRight(module, "/"), typ}, "/")
	}
	return out
}

// isComponentImport returns true if the import corresponds to  a Otel component,
// i.e. an extension, exporter, processor or a receiver.
func isComponentImport(importStr string, importPrefixesToCheck []string) bool {
	for _, prefix := range importPrefixesToCheck {
		if strings.HasPrefix(importStr, prefix) {
			return true
		}
	}
	return false
}
