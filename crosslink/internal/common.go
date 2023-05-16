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

package crosslink

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"go.uber.org/zap"
	"golang.org/x/mod/modfile"
)

// Attempts to identify a go module at the root path. If no
// go.mod file is present an error is returned.
func identifyRootModule(rootPath string) (string, error) {
	rootModPath := filepath.Clean(filepath.Join(rootPath, "go.mod"))
	if _, err := os.Stat(rootModPath); err != nil {
		return "", fmt.Errorf("failed to identify go.mod file at root dir: %w", err)
	}

	// identify and read the root module
	rootModFile, err := os.ReadFile(rootModPath)
	if err != nil {
		return "", fmt.Errorf("failed to read go.mod file at root dir: %w", err)
	}
	return modfile.ModulePath(rootModFile), nil
}

func writeModule(module *moduleInfo) error {
	modContents := module.moduleContents
	//  now overwrite the existing gomod file
	gomodFile, err := modContents.Format()
	if err != nil {
		return fmt.Errorf("failed to format go.mod file: %w", err)
	}
	// write our updated go.mod file
	err = os.WriteFile(modContents.Syntax.Name, gomodFile, 0600)
	if err != nil {
		return fmt.Errorf("failed to write go.mod file: %w", err)
	}

	return nil
}

// forGoModules iterates the fn for all Go modules under rootPath.
// The path argument of fn function is the path to the go.mod file
// relative to rootPath (e.g. go.mod, submodule/go.mod).
func forGoModules(logger *zap.Logger, rootPath string, fn func(path string) error) error {
	return fs.WalkDir(os.DirFS(rootPath), ".", func(path string, dir fs.DirEntry, err error) error {
		if err != nil {
			logger.Warn("File could not be read during fs.WalkDir",
				zap.Error(err),
				zap.String("path", path))
			return nil
		}

		// find Go module
		if dir.Name() != "go.mod" {
			return nil
		}

		return fn(path)
	})
}
