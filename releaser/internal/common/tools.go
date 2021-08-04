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

// RunMakeLint runs 'make lint' to automatically update go.sum files.
func RunMakeLint() error {
	log.Println("Updating go.sum with 'make lint'")

	cmd := exec.Command("make", "lint")
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("'make lint' failed: %v (%v)", string(output), err)
	}

	return nil
}

// RunMakeCI runs 'make ci' to build tools needed for CI as defined by the Makefile.
func RunMakeCI() error {
	log.Println("Running 'make ci'")

	cmd := exec.Command("make", "ci")
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("'make ci' failed: %v (%v)", string(output), err)
	}

	return nil
}


// UpdateGoModVersions updates one go.mod file, given by modFilePath, by updating all modules listed in
// newModPaths to use the newVersion given.
func UpdateGoModVersions(modFilePath ModuleFilePath, newModPaths []ModulePath, newVersion string) error {
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

// UpdateAllGoModFiles updates the go.mod files in modFilePaths by updating all modules listed in
// newModPaths to use the newVersion given.
func UpdateAllGoModFiles(modFilePaths []ModuleFilePath, newModPaths []ModulePath, newVersion string) error {
	log.Println("Updating all module versions in go.mod files...")
	for _, modFilePath := range modFilePaths {
		if err := UpdateGoModVersions(
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
