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
	"strings"
)

// combineModuleTagNamesAndVersion combines a slice of ModuleTagNames with the version number and returns
// the new full module tags.
func combineModuleTagNamesAndVersion(modTagNames []ModuleTagName, version string) []string {
	var modFullTags []string
	for _, modTagName := range modTagNames {
		var newFullTag string
		if modTagName == repoRootTag {
			newFullTag = version
		} else {
			newFullTag = string(modTagName) + "/" + version
		}
		modFullTags = append(modFullTags, newFullTag)
	}

	return modFullTags
}

// ModulePathsToTagNames returns a list of tag names from a list of module's import paths.
func ModulePathsToTagNames(modPaths []ModulePath, modPathMap ModulePathMap, repoRoot string) ([]ModuleTagName, error) {
	modFilePaths, err := modulePathsToFilePaths(modPaths, modPathMap)
	if err != nil {
		return nil, fmt.Errorf("could not convert module paths to file paths: %v", err)
	}

	modTagNames, err := moduleFilePathsToTagNames(modFilePaths, repoRoot)
	if err != nil {
		return nil, fmt.Errorf("could not convert module file paths to tag names: %v", err)
	}

	return modTagNames, nil
}

// modulePathsToFilePaths returns a list of absolute file paths from a list of module's import paths.
func modulePathsToFilePaths(modPaths []ModulePath, modPathMap ModulePathMap) ([]ModuleFilePath, error) {
	var modFilePaths []ModuleFilePath

	for _, modPath := range modPaths {
		if _, exists := modPathMap[modPath]; !exists {
			return []ModuleFilePath{}, fmt.Errorf("could not find module path %v in path map.", modPath)
		}
		modFilePaths = append(modFilePaths, modPathMap[modPath])
	}

	return modFilePaths, nil
}

// moduleFilePathToTagName returns the module tag names of an input ModuleFilePath
// by removing the repoRoot prefix from the ModuleFilePath.
func moduleFilePathToTagName(modFilePath ModuleFilePath, repoRoot string) (ModuleTagName, error) {
	if !strings.HasPrefix(string(modFilePath), repoRoot+"/") {
		return "", fmt.Errorf("modFilePath %v not contained in repo with root %v", modFilePath, repoRoot)
	}
	if !strings.HasSuffix(string(modFilePath), "go.mod") {
		return "", fmt.Errorf("modFilePath %v does not end with 'go.mod'", modFilePath)
	}

	modTagNameWithGoMod := strings.TrimPrefix(string(modFilePath), repoRoot+"/")
	modTagName := strings.TrimSuffix(modTagNameWithGoMod, "/go.mod")

	// if the modTagName is equal to go.mod, it is the root repo
	if modTagName == "go.mod" {
		return repoRootTag, nil
	}

	return ModuleTagName(modTagName), nil
}

// moduleFilePathsToTagNames returns a list of module tag names from the input full module file paths
// by removing the repoRoot prefix from each ModuleFilePath.
func moduleFilePathsToTagNames(modFilePaths []ModuleFilePath, repoRoot string) ([]ModuleTagName, error) {
	var modNames []ModuleTagName

	for _, modFilePath := range modFilePaths {
		modTagName, err := moduleFilePathToTagName(modFilePath, repoRoot)
		if err != nil {
			return nil, fmt.Errorf("could not convert module file path to tag name: %v", err)
		}
		modNames = append(modNames, modTagName)
	}

	return modNames, nil
}
