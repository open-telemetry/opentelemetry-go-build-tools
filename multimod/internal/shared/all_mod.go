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

package shared

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"golang.org/x/mod/modfile"
)

func newAllModulePathMap(root string) (ModulePathMap, error) {
	modPathMap := make(ModulePathMap)

	findGoMod := func(filePath string, _ fs.FileInfo, err error) error {
		if err != nil {
			fmt.Printf("Warning: file could not be read during filepath.Walk(): %v", err)
			return nil
		}
		if filepath.Base(filePath) == "go.mod" {
			// read go.mod file into mod []byte
			mod, err := os.ReadFile(filepath.Clean(filePath))
			if err != nil {
				return err
			}

			// read path of module from go.mod file
			modPathString := modfile.ModulePath(mod)

			// convert modPath, filePath string to modulePath and moduleFilePath
			modPath := ModulePath(modPathString)
			modFilePath := ModuleFilePath(filePath)

			modPathMap[modPath] = modFilePath
		}
		return nil
	}

	if err := filepath.Walk(root, findGoMod); err != nil {
		return nil, err
	}

	return modPathMap, nil
}
