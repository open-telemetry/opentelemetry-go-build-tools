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

// ModuleVersioning holds info about modules listed in a versioning file.
type ModuleVersioning struct {
	ModSetMap  ModuleSetMap
	ModPathMap ModulePathMap
	ModInfoMap ModuleInfoMap
}

// NewModuleVersioningWithIgnoreExcluded returns a ModuleVersioning struct from a versioning file and repo root and supports
// ignoring the excluded-modules configuration.
func NewModuleVersioningWithIgnoreExcluded(versioningFilename string, repoRoot string, ignoreExcluded bool) (ModuleVersioning, error) {
	repoRoot, err := filepath.Abs(repoRoot)
	if err != nil {
		return ModuleVersioning{}, fmt.Errorf("could not get absolute path of repo root: %w", err)
	}

	vCfg, err := readVersioningFile(versioningFilename)
	vCfg.ignoreExcluded = ignoreExcluded

	if err != nil {
		return ModuleVersioning{}, fmt.Errorf("error reading versioning file %v: %w", versioningFilename, err)
	}

	modSetMap := vCfg.buildModuleSetsMap()

	modInfoMap, err := vCfg.buildModuleMap()
	if err != nil {
		return ModuleVersioning{}, fmt.Errorf("error building module info map for NewModuleVersioning: %w", err)
	}

	modPathMap, err := vCfg.BuildModulePathMap(repoRoot)
	if err != nil {
		return ModuleVersioning{}, fmt.Errorf("error building module path map for NewModuleVersioning: %w", err)
	}

	return ModuleVersioning{
		ModSetMap:  modSetMap,
		ModPathMap: modPathMap,
		ModInfoMap: modInfoMap,
	}, nil
}

// NewModuleVersioning returns a ModuleVersioning struct from a versioning file and repo root.
func NewModuleVersioning(versioningFilename string, repoRoot string) (ModuleVersioning, error) {
	return NewModuleVersioningWithIgnoreExcluded(versioningFilename, repoRoot, false)
}
