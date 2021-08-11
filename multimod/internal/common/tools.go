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
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"golang.org/x/mod/semver"

	tools "go.opentelemetry.io/build-tools"
)

// IsStableVersion returns true if modSet.Version is stable (i.e. version major greater than
// or equal to v1), else false.
func IsStableVersion(v string) bool {
	return semver.Compare(semver.Major(v), "v1") >= 0
}

// ChangeToRepoRoot changes to the root of the Git repository the script is called from and returns it as a string.
func ChangeToRepoRoot() (string, error) {
	repoRoot, err := tools.FindRepoRoot()
	if err != nil {
		return "", fmt.Errorf("unable to find repo root: %v", err)
	}

	log.Println("Changing to root directory...")
	err = os.Chdir(repoRoot)
	if err != nil {
		return "", fmt.Errorf("unable to change to repo root: %v", err)
	}

	return repoRoot, nil
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
		oldVersionRegex := filePathToRegex(string(modPath)) + `\s+` + SemverRegex
		r, err := regexp.Compile(oldVersionRegex)
		if err != nil {
			return fmt.Errorf("error compiling regex: %v", err)
		}

		newModVersionString := string(modPath) + " " + newVersion

		newGoModFile = r.ReplaceAll(newGoModFile, []byte(newModVersionString))
	}

	// once all module versions have been updated, overwrite the go.mod file
	if err := ioutil.WriteFile(string(modFilePath), newGoModFile, 0644); err != nil {
		return fmt.Errorf("error overwriting go.mod file: %v", err)
	}

	return nil
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
		cmd := exec.Command("go", "mod", "tidy")
		cmd.Dir = filepath.Dir(string(modFilePath))

		if out, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("go mod tidy failed: %v\n%v", string(out), err)
		}
	}

	return nil
}
