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
	"os"
	"path/filepath"
	"sort"
	"strings"

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

func buildUses(rootModulePath string, graph map[string]*moduleInfo, rc RunConfig) ([]string, error) {
	var uses []string
	for module := range graph {
		// skip excluded
		if _, exists := rc.ExcludedPaths[module]; exists {
			rc.Logger.Debug("Excluded Module, ignoring use",
				zap.String("module", module))
			continue
		}

		localPath, err := filepath.Rel(rootModulePath, module)
		if err != nil {
			return nil, fmt.Errorf("failed to retrieve relative path: %w", err)
		}
		if localPath == "." || localPath == ".." {
			localPath += "/"
		} else if !strings.HasPrefix(localPath, "..") {
			localPath = "./" + localPath
		}
		uses = append(uses, localPath)
	}
	sort.Strings(uses)
	return uses, nil
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

func openGoWork(rc RunConfig) (*modfile.WorkFile, error) {
	goWorkPath := filepath.Join(rc.RootPath, "go.work")
	content, err := os.ReadFile(filepath.Clean(goWorkPath))
	if err != nil {
		return nil, err
	}
	return modfile.ParseWork(goWorkPath, content, nil)
}

func writeGoWork(goWork *modfile.WorkFile, rc RunConfig) error {
	goWorkPath := filepath.Join(rc.RootPath, "go.work")
	content := modfile.Format(goWork.Syntax)
	return os.WriteFile(goWorkPath, content, 0600)
}
