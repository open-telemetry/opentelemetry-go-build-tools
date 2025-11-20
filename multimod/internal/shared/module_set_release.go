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
	"path/filepath"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
)

// ModuleSetRelease contains info about a specific set of modules in the versioning file to be updated.
type ModuleSetRelease struct {
	ModuleVersioning
	ModSetName    string
	ModSet        ModuleSet
	TagNames      []ModuleTagName
	AllModPathMap ModulePathMap
}

// NewModuleSetRelease returns a ModuleSetRelease struct by specifying a specific set of modules to update.
func NewModuleSetRelease(versioningFilename, modSetToUpdate, repoRoot string) (ModuleSetRelease, error) {
	repoRoot, err := filepath.Abs(repoRoot)
	if err != nil {
		return ModuleSetRelease{}, fmt.Errorf("could not get absolute path of repo root: %w", err)
	}

	modVersioning, err := NewModuleVersioning(versioningFilename, repoRoot)
	if err != nil {
		return ModuleSetRelease{}, fmt.Errorf("unable to load baseVersionStruct: %w", err)
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
		return ModuleSetRelease{}, fmt.Errorf("could not retrieve tag names from module paths: %w", err)
	}

	// get all module paths
	allModPathMap, err := newAllModulePathMap(repoRoot)
	if err != nil {
		return ModuleSetRelease{}, fmt.Errorf("could not retrieve all module paths: %w", err)
	}

	return ModuleSetRelease{
		ModuleVersioning: modVersioning,
		ModSetName:       modSetToUpdate,
		ModSet:           modSet,
		TagNames:         tagNames,
		AllModPathMap:    allModPathMap,
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

// CheckGitTagsAlreadyExist checks if Git tags have already been created that match the specific module tag name
// and version number for the modules being updated. If the tag already exists, an error is returned.
func (modRelease ModuleSetRelease) CheckGitTagsAlreadyExist(repo *git.Repository) error {
	newTags := make(map[string]bool)

	modFullTags := modRelease.ModuleFullTagNames()

	for _, newFullTag := range modFullTags {
		newTags[newFullTag] = true
	}

	existingTags, err := repo.Tags()
	if err != nil {
		return fmt.Errorf("error getting repo tags: %w", err)
	}

	var existingGitTagNames []string

	err = existingTags.ForEach(func(ref *plumbing.Reference) error {
		existingTagName := ref.Name().Short()

		if _, exists := newTags[existingTagName]; exists {
			existingGitTagNames = append(existingGitTagNames, existingTagName)
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("could not check all git tags: %w", err)
	}

	switch len(existingGitTagNames) {
	case len(newTags):
		return ErrGitTagsAlreadyExist{
			tagNames: existingGitTagNames,
		}
	case 0:
		return nil
	default:
		return ErrInconsistentGitTagsExist{
			tagNames: existingGitTagNames,
		}
	}
}
