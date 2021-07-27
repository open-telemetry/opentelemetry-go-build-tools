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

package commontest

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"go.opentelemetry.io/build-tools/releaser/internal/common"
)

// MockModuleVersioning creates a ModuleVersioning struct for testing purposes.
func MockModuleVersioning(modSetMap common.ModuleSetMap, modPathMap common.ModulePathMap) (common.ModuleVersioning, error) {
	modInfoMap := make(common.ModuleInfoMap)

	for setName, moduleSet := range modSetMap {
		for _, modPath := range moduleSet.Modules {
			// Check if module has already been added to the map
			if _, exists := modInfoMap[modPath]; exists {
				return common.ModuleVersioning{}, fmt.Errorf("Module %v exists more than once. Exists in sets %v and %v.",
					modPath, modInfoMap[modPath].ModuleSetName, setName)
			}

			modInfoMap[modPath] = common.ModuleInfo{ModuleSetName: setName, Version: moduleSet.Version}
		}
	}

	return common.ModuleVersioning{
		ModSetMap:  modSetMap,
		ModPathMap: modPathMap,
		ModInfoMap: modInfoMap,
	}, nil
}

// MockModuleSetRelease creates a ModuleSetRelease struct for testing purposes.
func MockModuleSetRelease(modSetMap common.ModuleSetMap, modPathMap common.ModulePathMap, modSetToUpdate string, repoRoot string) (common.ModuleSetRelease, error) {
	modVersioning, err := MockModuleVersioning(modSetMap, modPathMap)
	if err != nil {
		return common.ModuleSetRelease{}, fmt.Errorf("error getting MockModuleVersioning: %v", err)
	}

	modSet := modSetMap[modSetToUpdate]

	// get tag names of mods to update
	tagNames, err := common.ModulePathsToTagNames(
		modSet.Modules,
		modPathMap,
		repoRoot,
	)

	return common.ModuleSetRelease{
		ModuleVersioning: modVersioning,
		ModSetName:       modSetToUpdate,
		ModSet:           modSet,
		TagNames:         tagNames,
	}, nil
}

// WriteGoModFiles is a helper function to dynamically write go.mod files used for testing.
func WriteGoModFiles(modFiles map[common.ModuleFilePath][]byte) error {
	perm := os.FileMode(0700)

	for modFilePath, file := range modFiles {
		path := filepath.Dir(string(modFilePath))
		os.MkdirAll(path, perm)

		if err := ioutil.WriteFile(string(modFilePath), file, perm); err != nil {
			return fmt.Errorf("could not write temporary mod file %v", err)
		}
	}

	return nil
}
