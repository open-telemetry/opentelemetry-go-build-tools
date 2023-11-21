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
	"io/fs"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
	"golang.org/x/mod/modfile"
)

const (
	SemverRegexNumberOnly = `(0|[1-9]\d*)\.(0|[1-9]\d*)\.(0|[1-9]\d*)(?:-((?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*)(?:\.(?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*))*))?(?:\+([0-9a-zA-Z-]+(?:\.[0-9a-zA-Z-]+)*))?`
	SemverRegex           = `v` + SemverRegexNumberOnly
)

// versionConfig is needed to parse the versions.yaml file with viper.
type versionConfig struct {
	ModuleSets      ModuleSetMap `mapstructure:"module-sets"`
	ExcludedModules []ModulePath `mapstructure:"excluded-modules"`
	ignoreExcluded  bool
}

// excludedModules functions as a set containing all module paths that are excluded
// from versioning.
type excludedModulesSet map[ModulePath]struct{}

// ModuleSetMap maps the name of a module set to a ModuleSet struct.
type ModuleSetMap map[string]ModuleSet

// ModuleSet holds the version that the specified modules within the set will have.
type ModuleSet struct {
	Version string       `mapstructure:"version"`
	Modules []ModulePath `mapstructure:"modules"`
}

// ModulePath holds the module import path, such as "go.opentelemetry.io/otel".
type ModulePath string

// ModuleInfoMap is a mapping from a module's import path to its ModuleInfo struct.
type ModuleInfoMap map[ModulePath]ModuleInfo

// ModuleInfo is a reverse of the ModuleSetMap, to allow for quick lookup from module
// path to its set and version.
type ModuleInfo struct {
	ModuleSetName string
	Version       string
}

// ModuleFilePath holds the file path to the go.mod file within the repo,
// including the base file name ("go.mod").
type ModuleFilePath string

// ModulePathMap is a mapping from a module's import path to its file path.
type ModulePathMap map[ModulePath]ModuleFilePath

// ModuleTagName is the simple file path to the directory of a go.mod file used for Git tagging.
// For example, the opentelemetry-go/sdk/metric/go.mod file will have a ModuleTagName "sdk/metric".
type ModuleTagName string

// readVersioningFile reads in a versioning file (typically given as versions.yaml) and returns
// a versionConfig struct.
func readVersioningFile(versioningFilename string) (versionConfig, error) {
	viper.SetConfigFile(versioningFilename)

	var versionCfg versionConfig

	if err := viper.ReadInConfig(); err != nil {
		return versionConfig{}, fmt.Errorf("error reading versionsConfig file: %w", err)
	}

	if err := viper.Unmarshal(&versionCfg); err != nil {
		return versionConfig{}, fmt.Errorf("unable to unmarshal versionsConfig: %w", err)
	}

	if viper.ConfigFileUsed() != versioningFilename {
		return versionConfig{}, fmt.Errorf(
			"config file used (%v) does not match input file (%v)",
			viper.ConfigFileUsed(),
			versioningFilename,
		)
	}

	return versionCfg, nil
}

// buildModuleSetsMap creates a map with module set names as keys and ModuleSet structs as values.
func (versionCfg versionConfig) buildModuleSetsMap() ModuleSetMap {
	return versionCfg.ModuleSets
}

// BuildModuleMap creates a map with module paths as keys and their moduleInfo as values
// by creating and "reversing" a ModuleSetsMap.
func (versionCfg versionConfig) buildModuleMap() (ModuleInfoMap, error) {
	modMap := make(ModuleInfoMap)

	for setName, moduleSet := range versionCfg.ModuleSets {
		for _, modPath := range moduleSet.Modules {
			// Check if module has already been added to the map
			if _, exists := modMap[modPath]; exists {
				return nil, fmt.Errorf("module %v exists more than once (exists in sets %v and %v)",
					modPath, modMap[modPath].ModuleSetName, setName)
			}

			// Check if module is in excluded modules section
			if !versionCfg.ignoreExcluded && versionCfg.shouldExcludeModule(modPath) {
				return nil, fmt.Errorf("module %v is an excluded module and should not be versioned", modPath)
			}
			modMap[modPath] = ModuleInfo{setName, moduleSet.Version}
		}
	}

	return modMap, nil
}

// getExcludedModules returns if a given module path is listed in the excluded modules section of a versioning file.
func (versionCfg versionConfig) shouldExcludeModule(modPath ModulePath) bool {
	if versionCfg.ignoreExcluded {
		return false
	}

	excludedModules := versionCfg.getExcludedModules()
	_, exists := excludedModules[modPath]

	return exists
}

// getExcludedModules returns a map structure containing all excluded module paths as keys and empty values.
func (versionCfg versionConfig) getExcludedModules() excludedModulesSet {
	excludedModules := make(excludedModulesSet)
	if versionCfg.ignoreExcluded {
		return excludedModules
	}
	// add all excluded modules to the excludedModulesSet
	for _, mod := range versionCfg.ExcludedModules {
		excludedModules[mod] = struct{}{}
	}

	return excludedModules
}

// BuildModulePathMap creates a map with module paths as keys and go.mod file paths as values.
func (versionCfg versionConfig) BuildModulePathMap(root string) (ModulePathMap, error) {
	modPathMap := make(ModulePathMap)

	findGoMod := func(filePath string, info fs.FileInfo, err error) error {
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

			excludedModules := versionCfg.getExcludedModules()
			if _, shouldExclude := excludedModules[modPath]; !shouldExclude {
				modPathMap[modPath] = modFilePath
			}
		}
		return nil
	}

	if err := filepath.Walk(root, findGoMod); err != nil {
		return nil, err
	}

	return modPathMap, nil
}
