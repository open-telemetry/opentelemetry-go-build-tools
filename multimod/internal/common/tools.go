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
	"log"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"golang.org/x/mod/semver"
)

// IsStableVersion returns true if modSet.Version is stable (i.e. version major greater than
// or equal to v1), else false.
func IsStableVersion(v string) bool {
	return semver.Compare(semver.Major(v), "v1") >= 0
}

// GetAllModuleSetNames returns the name of all module sets given in a versioningFile.
func GetAllModuleSetNames(versioningFile string, repoRoot string) ([]string, error) {
	modVersioning, err := NewModuleVersioning(versioningFile, repoRoot)
	if err != nil {
		return nil, fmt.Errorf("call failed to NewModuleVersioning: %v", err)
	}

	var modSetNames []string

	for modSetName := range modVersioning.ModSetMap {
		modSetNames = append(modSetNames, modSetName)
	}

	return modSetNames, nil
}

func GetModuleSet(modSetName, versioningFilename string) (ModuleSet, error) {
	vCfg, err := readVersioningFile(versioningFilename)
	if err != nil {
		return ModuleSet{}, fmt.Errorf("error reading versioning file %v: %v", versioningFilename, err)
	}

	modSetMap, err := vCfg.buildModuleSetsMap()
	if err != nil {
		return ModuleSet{}, fmt.Errorf("error building module set map: %v", err)
	}
	return modSetMap[modSetName], nil
}

// updateGoModVersions updates one go.mod file, given by modFilePath, by updating all modules listed in
// newModPaths to use the newVersion given.
func updateGoModVersions(modFilePath ModuleFilePath, newModPaths []ModulePath, newVersion string) error {
	if !strings.HasSuffix(string(modFilePath), "go.mod") {
		return fmt.Errorf("cannot update file passed that does not end with go.mod")
	}

	newGoModFile, err := ioutil.ReadFile(string(modFilePath))
	if err != nil {
		panic(err)
	}

	for _, modPath := range newModPaths {
		newGoModFile, err = replaceModVersion(modPath, newVersion, newGoModFile)
		if err != nil {
			return err
		}
	}

	// once all module versions have been updated, overwrite the go.mod file
	if err := ioutil.WriteFile(string(modFilePath), newGoModFile, 0644); err != nil {
		return fmt.Errorf("error overwriting go.mod file: %v", err)
	}

	return nil
}

func replaceModVersion(modPath ModulePath, version string, newGoModFile []byte) ([]byte, error) {
	oldVersionRegex := `(?m:` + filePathToRegex(string(modPath)) + `\s+` + SemverRegex + `(\s*\/\/\s*indirect\s*?)?$)`
	r, err := regexp.Compile(oldVersionRegex)
	if err != nil {
		return nil, fmt.Errorf("error compiling regex: %v", err)
	}

	newModVersionString := string(modPath) + " " + version

	// ${6} is the capture group that has " // indirect" if it was present in the original
	newGoModFile = r.ReplaceAll(newGoModFile, []byte(newModVersionString+"${6}"))
	return newGoModFile, nil
}

// UpdateGoModFiles updates the go.mod files in modFilePaths by updating all modules listed in
// newModPaths to use the newVersion given.
func UpdateGoModFiles(modFilePaths []ModuleFilePath, newModPaths []ModulePath, newVersion string) error {
	log.Println("Updating all module versions in go.mod files...")
	for _, modFilePath := range modFilePaths {
		if err := updateGoModVersions(
			modFilePath,
			newModPaths,
			newVersion,
		); err != nil {
			return fmt.Errorf("could not update module versions in file %v: %v", modFilePath, err)
		}
	}
	return nil
}

func filePathToRegex(fpath string) string {
	quotedMeta := regexp.QuoteMeta(fpath)
	replacedSlashes := strings.Replace(quotedMeta, string(filepath.Separator), `\/`, -1)
	return replacedSlashes
}

// RunGoModTidy takes a ModulePathMap and runs "go mod tidy" at each module file path.
func RunGoModTidy(modPathMap ModulePathMap) error {
	for _, modFilePath := range modPathMap {
		cmd := exec.Command("go", "mod", "tidy", "-compat=1.17")
		cmd.Dir = filepath.Dir(string(modFilePath))

		if out, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("go mod tidy failed: %v\n%v", string(out), err)
		}
	}

	return nil
}
