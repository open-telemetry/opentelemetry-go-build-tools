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

package prerelease

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"go.opentelemetry.io/build-tools/multimod/internal/common"
)

func Run(versioningFile, moduleSetName, fromExistingBranch string, skipGoModTidy bool) {

	repoRoot, err := common.ChangeToRepoRoot()
	if err != nil {
		log.Fatalf("unable to change to repo root: %v", err)
	}

	p, err := newPrerelease(versioningFile, moduleSetName, repoRoot)
	if err != nil {
		log.Fatalf("Error creating new prerelease struct: %v", err)
	}

	if err = p.verifyGitTagsDoNotAlreadyExist(); err != nil {
		log.Fatalf("verifyGitTagsDoNotAlreadyExist failed: %v", err)
	}

	if err = p.verifyWorkingTreeClean(); err != nil {
		log.Fatalf("verifyWorkingTreeClean failed: %v", err)
	}

	if err = p.createPrereleaseBranch(fromExistingBranch); err != nil {
		log.Fatalf("createPrereleaseBranch failed: %v", err)
	}

	if err = p.updateAllVersionGo(); err != nil {
		log.Fatal("updateAllVersionGo failed:", err)
	}

	if err = p.updateAllGoModFiles(); err != nil {
		log.Fatalf("updateAllGoModFiles failed: %v", err)
	}

	if skipGoModTidy {
		fmt.Println("Skipping 'go mod tidy'...")
	} else {
		if err = common.RunGoModTidy(p.ModuleSetRelease.ModuleVersioning.ModPathMap); err != nil {
			log.Fatalf("runGoModTidy failed: %v", err)
		}
	}

	if err = p.commitChanges(); err != nil {
		log.Fatalf("commitChanges failed: %v", err)
	}

	fmt.Println("\nPrerelease finished successfully. Now run the following to verify the changes:")
	fmt.Println("\ngit diff main")
	fmt.Println("\nThen, push the changes to upstream.")
}

type prerelease struct {
	common.ModuleSetRelease
}

func newPrerelease(versioningFilename, modSetToUpdate, repoRoot string) (prerelease, error) {
	modRelease, err := common.NewModuleSetRelease(versioningFilename, modSetToUpdate, repoRoot)
	if err != nil {
		log.Fatalf("Error creating new prerelease struct: %v", err)
	}

	return prerelease{
		ModuleSetRelease: modRelease,
	}, nil
}

// verifyGitTagsDoNotAlreadyExist checks if Git tags have already been created that match the specific module tag name
// and version number for the modules being updated. If the tag already exists, an error is returned.
func (p prerelease) verifyGitTagsDoNotAlreadyExist() error {
	modFullTags := p.ModuleSetRelease.ModuleFullTagNames()

	cmd := exec.Command("git", "tag")
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("could not execute git tag: %v", err)
	}

	existingTags := map[string]bool{}

	for _, tag := range strings.Split(string(output), "\n") {
		existingTags[strings.TrimSpace(tag)] = true
	}

	for _, newFullTag := range modFullTags {
		if existingTags[newFullTag] {
			return fmt.Errorf("git tag already exists for %v", newFullTag)
		}
	}

	return nil
}

// verifyWorkingTreeClean checks if the working tree is clean (i.e. running 'git diff --exit-code' gives exit code 0).
// If the working tree is not clean, the git diff output is printed, and an error is returned.
func (p prerelease) verifyWorkingTreeClean() error {
	cmd := exec.Command("git", "diff", "--exit-code")
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("working tree is not clean, can't proceed with the release process:\n\n%v",
			string(output),
		)
	}

	return nil
}

func (p prerelease) createPrereleaseBranch(fromExistingBranch string) error {
	branchNameElements := []string{"pre_release", p.ModuleSetRelease.ModSetName, p.ModuleSetRelease.ModSetVersion()}
	branchName := strings.Join(branchNameElements, "_")
	fmt.Printf("git checkout -b %v %v\n", branchName, fromExistingBranch)
	cmd := exec.Command("git", "checkout", "-b", branchName, fromExistingBranch)

	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("could not create new branch %v: %v (%v)", branchName, string(output), err)
	}

	return nil
}

func (p prerelease) commitChanges() error {
	commitMessage := "Prepare for versions " + p.ModuleSetRelease.ModSetVersion()

	// add changes to git
	cmd := exec.Command("git", "add", ".")
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("'git add .' failed: %v (%v)", string(output), err)
	}

	// commit changes to git
	fmt.Printf("Commit changes to git with message '%v'...\n", commitMessage)
	cmd = exec.Command("git", "commit", "-m", commitMessage)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git commit failed: %v (%v)", string(output), err)
	}

	cmd = exec.Command("git", "log", `--pretty=format:"%h"`, "-1")
	output, err := cmd.Output()
	if err != nil {
		fmt.Println("WARNING: could not automatically get last commit hash.")
	}

	fmt.Printf("Commit successful. Hash of commit: %s\n", output)

	return nil
}

// updateGoModVersions reads the fromFile (a go.mod file), replaces versions
// for all specified modules in newModPaths, and writes the new go.mod to the toFile file.
func (p prerelease) updateGoModVersions(modFilePath common.ModuleFilePath) error {
	newGoModFile, err := ioutil.ReadFile(string(modFilePath))
	if err != nil {
		panic(err)
	}

	for _, modPath := range p.ModuleSetRelease.ModSetPaths() {
		newGoModFile, err = replaceModVersion(modPath, p.ModuleSetRelease.ModSetVersion(), newGoModFile)
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

// updateAllVersionGo updates the version.go file containing a hardcoded semver version string
// for modules within a set, if the file exists.
func (p prerelease) updateAllVersionGo() error {
	for _, modPath := range p.ModuleSetRelease.ModSetPaths() {
		modFilePath := p.ModuleSetRelease.ModuleVersioning.ModPathMap[modPath]

		versionGoDir := filepath.Dir(string(modFilePath))
		versionGoFilePath := filepath.Join(versionGoDir, "version.go")

		// check if version.go file exists
		if _, err := os.Stat(versionGoFilePath); err == nil {
			if updateErr := updateVersionGoFile(versionGoFilePath, p.ModuleSetRelease.ModSetVersion()); updateErr != nil {
				return fmt.Errorf("could not update %v: %v", versionGoFilePath, updateErr)
			}
		} else if os.IsNotExist(err) {
			continue
		} else {
			return fmt.Errorf("could not check existance of %v: %v", versionGoFilePath, err)
		}

	}
	return nil
}

// updateVersionGoFile updates one version.go file.
// TODO: a potential improvement is to use an AST package rather than regex to perform replacement.
func updateVersionGoFile(filePath string, newVersion string) error {
	if !strings.HasSuffix(filePath, "version.go") {
		return fmt.Errorf("cannot update file passed that does not end with version.go")
	}
	log.Printf("... Updating file %v\n", filePath)

	newVersionGoFile, err := ioutil.ReadFile(filePath)
	if err != nil {
		panic(err)
	}

	oldVersionRegex := common.SemverRegexNumberOnly
	r, err := regexp.Compile(oldVersionRegex)
	if err != nil {
		return fmt.Errorf("error compiling regex: %v", err)
	}

	newVersionNumberOnly := strings.TrimPrefix(newVersion, "v")

	newVersionGoFile = r.ReplaceAll(newVersionGoFile, []byte(newVersionNumberOnly))

	// overwrite the version.go file
	if err := ioutil.WriteFile(filePath, newVersionGoFile, 0644); err != nil {
		return fmt.Errorf("error overwriting go.mod file: %v", err)
	}

	return nil
}

func replaceModVersion(modPath common.ModulePath, version string, newGoModFile []byte) ([]byte, error) {
	oldVersionRegex := `(?m:` + filePathToRegex(string(modPath)) + `\s+` + common.SemverRegex + `(\s*\/\/\s*indirect\s*?)?$)`
	r, err := regexp.Compile(oldVersionRegex)
	if err != nil {
		return nil, fmt.Errorf("error compiling regex: %v", err)
	}

	newModVersionString := string(modPath) + " " + version

	// ${6} is the capture group that has " // indirect" if it was present in the original
	newGoModFile = r.ReplaceAll(newGoModFile, []byte(newModVersionString+"${6}"))
	return newGoModFile, nil
}

// updateAllGoModFiles updates ALL modules' requires sections to use the newVersion number
// for the modules given in newModPaths.
func (p prerelease) updateAllGoModFiles() error {
	fmt.Println("Updating all module versions in go.mod files...")
	for _, modFilePath := range p.ModuleSetRelease.ModPathMap {
		if err := p.updateGoModVersions(modFilePath); err != nil {
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
