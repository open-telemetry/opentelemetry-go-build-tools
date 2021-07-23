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

package common

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
)

// MockModuleVersioning creates a ModuleVersioning struct for testing purposes.
func MockModuleVersioning(modSetMap ModuleSetMap, modPathMap ModulePathMap) (ModuleVersioning, error) {
	vCfg := versionConfig{
		ModuleSets:      modSetMap,
		ExcludedModules: []ModulePath{},
	}

	modInfoMap, err := vCfg.buildModuleMap()
	if err != nil {
		return ModuleVersioning{}, fmt.Errorf("error building module map: %v", err)
	}

	return ModuleVersioning{
		ModSetMap:  modSetMap,
		ModPathMap: modPathMap,
		ModInfoMap: modInfoMap,
	}, nil
}

// MockModuleSetRelease creates a ModuleSetRelease struct for testing purposes.
func MockModuleSetRelease(modSetMap ModuleSetMap, modPathMap ModulePathMap, modSetToUpdate string, repoRoot string) (ModuleSetRelease, error) {
	modVersioning, err := MockModuleVersioning(modSetMap, modPathMap)
	if err != nil {
		return ModuleSetRelease{}, fmt.Errorf("error getting MockModuleVersioning: %v", err)
	}

	modSet := modSetMap[modSetToUpdate]

	// get tag names of mods to update
	tagNames, err := modulePathsToTagNames(
		modSet.Modules,
		modPathMap,
		repoRoot,
	)

	return ModuleSetRelease{
		ModuleVersioning: modVersioning,
		ModSetName:       modSetToUpdate,
		ModSet:           modSet,
		TagNames:         tagNames,
	}, nil
}

// WriteGoModFiles is a helper function to dynamically write go.mod files used for testing.
func WriteGoModFiles(modFiles map[ModuleFilePath][]byte) error {
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
