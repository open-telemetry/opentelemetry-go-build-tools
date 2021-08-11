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
	"path/filepath"
)

// ModuleSetRelease contains info about a specific set of modules in the versioning file to be updated.
type ModuleSetRelease struct {
	ModuleVersioning
	ModSetName string
	ModSet     ModuleSet
	TagNames   []ModuleTagName
}

// NewModuleSetRelease returns a ModuleSetRelease struct by specifying a specific set of modules to update.
func NewModuleSetRelease(versioningFilename, modSetToUpdate, repoRoot string) (ModuleSetRelease, error) {
	repoRoot, err := filepath.Abs(repoRoot)
	if err != nil {
		return ModuleSetRelease{}, fmt.Errorf("could not get absolute path of repo root: %v", err)
	}

	modVersioning, err := NewModuleVersioning(versioningFilename, repoRoot)
	if err != nil {
		return ModuleSetRelease{}, fmt.Errorf("unable to load baseVersionStruct: %v", err)
	}

	// get new version and mod tags to update
	modSet, exists := modVersioning.ModSetMap[modSetToUpdate]
	if !exists {
		return ModuleSetRelease{}, fmt.Errorf("could not find module set %v in versioning file", modSetToUpdate)
	}

	// get tag names of mods to update
	tagNames, err := ModulePathsToTagNames(
		modSet.Modules,
		modVersioning.ModPathMap,
		repoRoot,
	)
	if err != nil {
		return ModuleSetRelease{}, fmt.Errorf("could not retrieve tag names from module paths: %v", err)
	}

	return ModuleSetRelease{
		ModuleVersioning: modVersioning,
		ModSetName:       modSetToUpdate,
		ModSet:           modSet,
		TagNames:         tagNames,
	}, nil

}

// ModSetVersion gets the version of the module set to update.
func (modRelease ModuleSetRelease) ModSetVersion() string {
	return modRelease.ModSet.Version
}

// ModSetPaths gets the import paths of all modules in the module set to update.
func (modRelease ModuleSetRelease) ModSetPaths() []ModulePath {
	return modRelease.ModSet.Modules
}

// ModuleFullTagNames gets the full tag names (including the version) of all modules in the module set to update.
func (modRelease ModuleSetRelease) ModuleFullTagNames() []string {
	return combineModuleTagNamesAndVersion(modRelease.TagNames, modRelease.ModSetVersion())
}
