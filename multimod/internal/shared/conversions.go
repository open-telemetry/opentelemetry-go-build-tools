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
	"path/filepath"
	"strings"
)

// RepoRootTag is a special tag name that indicates the root of the repository.
const RepoRootTag = ModuleTagName("REPOROOTTAG")

// combineModuleTagNamesAndVersion combines a slice of ModuleTagNames with the version number and returns
// the new full module tags.
func combineModuleTagNamesAndVersion(modTagNames []ModuleTagName, version string) []string {
	var modFullTags []string
	for _, modTagName := range modTagNames {
		var newFullTag string
		if modTagName == RepoRootTag {
			newFullTag = version
		} else {
			// Handle vN subdirectory modules - tag as <module_path>/vN.x.x not <module_path>/vN/vN.x.x
			modTagStr := string(modTagName)

			// Case 1: Root-level vN module (e.g., "v5" -> "v5.0.0", "v10" -> "v10.0.0")
			if isRootLevelVNModule(modTagStr, version) {
				newFullTag = version
			} else if vIdx := strings.LastIndex(modTagStr, "/v"); vIdx != -1 {
				// Case 2: Subdirectory vN module (e.g., "detectors/aws/ec2/v2" -> "detectors/aws/ec2/v2.0.0")
				vPart := modTagStr[vIdx+2:] // Extract the part after "/v"
				if isValidVNPart(vPart) && versionMatchesVN(vPart, version) {
					// e.g., detectors/aws/ec2/v2 + v2.0.0 => detectors/aws/ec2/v2.0.0
					newFullTag = modTagStr[:vIdx] + "/" + version
				} else {
					newFullTag = modTagStr + "/" + version
				}
			} else {
				// Case 3: Regular module (e.g., "foo/bar" -> "foo/bar/v1.2.3")
				newFullTag = modTagStr + "/" + version
			}
		}
		modFullTags = append(modFullTags, newFullTag)
	}
	return modFullTags
}

// ModulePathsToTagNames returns a list of tag names from a list of module's import paths.
func ModulePathsToTagNames(modPaths []ModulePath, modPathMap ModulePathMap, repoRoot string) ([]ModuleTagName, error) {
	modFilePaths, err := modulePathsToFilePaths(modPaths, modPathMap)
	if err != nil {
		return nil, fmt.Errorf("could not convert module paths to file paths: %w", err)
	}

	modTagNames, err := moduleFilePathsToTagNames(modFilePaths, repoRoot)
	if err != nil {
		return nil, fmt.Errorf("could not convert module file paths to tag names: %w", err)
	}

	return modTagNames, nil
}

// modulePathsToFilePaths returns a list of absolute file paths from a list of module's import paths.
func modulePathsToFilePaths(modPaths []ModulePath, modPathMap ModulePathMap) ([]ModuleFilePath, error) {
	var modFilePaths []ModuleFilePath

	for _, modPath := range modPaths {
		if _, exists := modPathMap[modPath]; !exists {
			return []ModuleFilePath{}, fmt.Errorf("could not find module path %v in path map", modPath)
		}
		modFilePaths = append(modFilePaths, modPathMap[modPath])
	}

	return modFilePaths, nil
}

// moduleFilePathToTagName returns the module tag names of an input ModuleFilePath
// by removing the repoRoot prefix from the ModuleFilePath.
func moduleFilePathToTagName(modFilePath ModuleFilePath, repoRoot string) (ModuleTagName, error) {
	// convert to slash to make sure the prefix and suffix checks work on Windows
	modFilePathSlash := filepath.ToSlash(string(modFilePath))
	repoRootSlash := filepath.ToSlash(repoRoot)

	if !strings.HasPrefix(modFilePathSlash, repoRootSlash+"/") {
		return "", fmt.Errorf("modFilePath %v not contained in repo with root %v", modFilePath, repoRoot)
	}
	if !strings.HasSuffix(modFilePathSlash, "go.mod") {
		return "", fmt.Errorf("modFilePath %v does not end with 'go.mod'", modFilePath)
	}

	modTagNameWithGoMod := strings.TrimPrefix(modFilePathSlash, repoRootSlash+"/")
	modTagName := strings.TrimSuffix(modTagNameWithGoMod, "/go.mod")

	// if the modTagName is equal to go.mod, it is the root repo
	if modTagName == "go.mod" {
		return RepoRootTag, nil
	}

	// ModuleTagName is always forward slash separated, even on Windows
	return ModuleTagName(filepath.ToSlash(modTagName)), nil
}

// moduleFilePathsToTagNames returns a list of module tag names from the input full module file paths
// by removing the repoRoot prefix from each ModuleFilePath.
func moduleFilePathsToTagNames(modFilePaths []ModuleFilePath, repoRoot string) ([]ModuleTagName, error) {
	var modNames []ModuleTagName

	for _, modFilePath := range modFilePaths {
		modTagName, err := moduleFilePathToTagName(modFilePath, repoRoot)
		if err != nil {
			return nil, fmt.Errorf("could not convert module file path to tag name: %w", err)
		}
		modNames = append(modNames, modTagName)
	}

	return modNames, nil
}

// isRootLevelVNModule checks if a module tag is a root-level vN module (e.g., "v2", "v10", "v123")
// and if the version matches the module's major version.
func isRootLevelVNModule(modTagStr, version string) bool {
	if len(modTagStr) < 2 || modTagStr[0] != 'v' {
		return false
	}

	// Extract version number from module tag (e.g., "v10" -> "10")
	versionPart := modTagStr[1:]
	if !isValidVNPart(versionPart) {
		return false
	}

	return versionMatchesVN(versionPart, version)
}

// isValidVNPart checks if a string is a valid version number (digits only, >= 2)
func isValidVNPart(vPart string) bool {
	if len(vPart) == 0 {
		return false
	}

	// Must be all digits
	for _, char := range vPart {
		if char < '0' || char > '9' {
			return false
		}
	}

	// Must be >= 2 (v0 and v1 don't use subdirectories)
	if len(vPart) == 1 && vPart[0] < '2' {
		return false
	}

	return true
}

// versionMatchesVN checks if a version string (e.g., "v10.0.0") matches the expected vN part (e.g., "10")
func versionMatchesVN(vPart, version string) bool {
	if len(version) < 2 || version[0] != 'v' {
		return false
	}

	// Extract major version from version string (e.g., "v10.0.0" -> "10")
	versionRest := version[1:]

	// Find the first dot or end of string
	majorVersionEnd := len(versionRest)
	for i, char := range versionRest {
		if char == '.' || char == '-' || char == '+' {
			majorVersionEnd = i
			break
		}
	}

	majorVersion := versionRest[:majorVersionEnd]
	return majorVersion == vPart
}
