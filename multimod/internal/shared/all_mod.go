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

package shared // nolint:revive // keeping generic package name until a proper refactoring is done

import (
	"fmt"
	"os"
	"path/filepath"

	"golang.org/x/mod/modfile"
)

func newAllModulePathMap(root string) (ModulePathMap, error) {
	modPathMap := make(ModulePathMap)

	var walk func(string) error
	walk = func(dir string) error {
		entries, err := os.ReadDir(dir)
		if err != nil {
			return fmt.Errorf("failed to read directory %q: %w", dir, err)
		}

		for _, entry := range entries {
			entryPath := filepath.Join(dir, entry.Name())

			if entry.IsDir() {
				if err := walk(entryPath); err != nil {
					return err
				}
				continue
			}

			if entry.Name() == "go.mod" {
				mod, err := os.ReadFile(filepath.Clean(entryPath))
				if err != nil {
					return err
				}

				modPathString := modfile.ModulePath(mod)
				modPath := ModulePath(modPathString)
				modFilePath := ModuleFilePath(entryPath)

				modPathMap[modPath] = modFilePath
			}
		}

		return nil
	}

	if err := walk(root); err != nil {
		return nil, err
	}

	return modPathMap, nil
}
