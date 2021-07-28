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

package verify

import (
	"fmt"
	"io/ioutil"
	"log"

	"golang.org/x/mod/modfile"
	"golang.org/x/mod/semver"

	tools "go.opentelemetry.io/build-tools"
	"go.opentelemetry.io/build-tools/releaser/internal/common"
)

func Run(versioningFile string) {

	repoRoot, err := tools.FindRepoRoot()
	if err != nil {
		log.Fatalf("unable to find repo root: %v", err)
	}

	v, err := newVerification(versioningFile, repoRoot)
	if err != nil {
		log.Fatalf("Error creating new verification struct: %v", err)
	}

	if err = v.verifyAllModulesInSet(); err != nil {
		log.Fatalf("verifyAllModulesInSet failed: %v", err)
	}

	if err = v.verifyVersions(); err != nil {
		log.Fatalf("verifyVersions failed: %v", err)
	}

	v.verifyDependencies()

	log.Println("PASS: Module sets successfully verified.")
}

type verification struct {
	common.ModuleVersioning
	dependencies dependencyMap
}

type dependencyMap map[common.ModulePath][]common.ModulePath

func newVerification(versioningFilename, repoRoot string) (verification, error) {
	modVersioning, err := common.NewModuleVersioning(versioningFilename, repoRoot)
	if err != nil {
		return verification{}, fmt.Errorf("call to NewModuleVersioning failed: %v\n", err)
	}

	dependencies, err := getDependencies(modVersioning)
	if err != nil {
		return verification{}, fmt.Errorf("could not get dependencies: %v\n", err)
	}

	return verification{
		ModuleVersioning: modVersioning,
		dependencies:     dependencies,
	}, nil
}

// getDependencies returns a map of each module's dependencies on other modules within the same repo.
func getDependencies(modVersioning common.ModuleVersioning) (dependencyMap, error) {
	dependencies := make(dependencyMap)

	// Dependencies are defined by the require section of go.mod files.
	for modPath, _ := range modVersioning.ModInfoMap {
		modFilePath := modVersioning.ModPathMap[modPath]
		modData, err := ioutil.ReadFile(string(modFilePath))

		modFile, err := modfile.Parse("teststring", modData, nil)
		if err != nil {
			return nil, fmt.Errorf("could not parse go.mod file at %v: %v", modFilePath, err)
		}

		// get dependencies as defined by the "require" section
		for _, dep := range modFile.Require {
			// check if dependency is in the same repo (i.e. if it exists in the module versioning file)
			if _, exists := modVersioning.ModInfoMap[common.ModulePath(dep.Mod.Path)]; exists {
				dependencies[modPath] = append(dependencies[modPath], common.ModulePath(dep.Mod.Path))
			}
		}
	}

	return dependencies, nil
}

// verifyAllModulesInSet checks that every module (as defined by a go.mod file) is contained in exactly
// one module set, unless it is excluded.
func (v verification) verifyAllModulesInSet() error {
	for modPath, modFilePath := range v.ModuleVersioning.ModPathMap {
		if _, exists := v.ModuleVersioning.ModInfoMap[modPath]; !exists {
			return &errModuleNotInSet{
				modPath:     modPath,
				modFilePath: modFilePath,
			}
		}
	}

	for modPath, modInfo := range v.ModuleVersioning.ModInfoMap {
		if _, exists := v.ModuleVersioning.ModPathMap[modPath]; !exists {
			return &errModuleNotInRepo{
				modPath:    modPath,
				modSetName: modInfo.ModuleSetName,
			}
		}
	}

	log.Println("PASS: All modules exist in exactly one set.")

	return nil
}

// verifyVersions checks that module set versions conform to versioning semantics.
func (v verification) verifyVersions() error {
	// setMajorVersions keeps track of all sets' major versions, used to check for multiple sets
	// with the same non-zero major version.
	setMajorVersions := make(map[string][]string)

	for modSetName, modSet := range v.ModuleVersioning.ModSetMap {
		// Check that module set versions conform to semver semantics
		if !semver.IsValid(modSet.Version) {
			return &errInvalidVersion{
				modSetName:    modSetName,
				modSetVersion: modSet.Version,
			}
		}

		if common.IsStableVersion(modSet.Version) {
			// Add all sets to major version map
			modSetMajorVersion := semver.Major(modSet.Version)
			setMajorVersions[modSetMajorVersion] = append(setMajorVersions[modSetMajorVersion], modSetName)
		}
	}

	// Check that no more than one module exists for any given non-zero major version
	for majorVersion, modSetNames := range setMajorVersions {
		if len(modSetNames) > 1 {
			return &errMultipleSetSameVersion{
				modSetNames:   modSetNames,
				modSetVersion: majorVersion,
			}
		}
	}

	log.Println("PASS: All module versions are valid, and no module sets have same non-zero major version.")

	return nil
}

// verifyDependencies checks that dependencies between modules conform to versioning semantics.
func (v verification) verifyDependencies() {
	for modPath, modDeps := range v.dependencies {
		// check if module is stable
		modVersion := v.ModuleVersioning.ModInfoMap[modPath].Version
		if common.IsStableVersion(modVersion) {

			for _, depPath := range modDeps {
				// check if dependency is on an unstable module
				depVersion := v.ModuleVersioning.ModInfoMap[depPath].Version
				if !common.IsStableVersion(depVersion) {
					log.Println(
						&errDependency{
							modPath:    modPath,
							modVersion: modVersion,
							depPath:    depPath,
							depVersion: depVersion,
						},
					)
				}
			}
		}
	}

	log.Println("Finished checking all stable modules' dependencies.")
}
