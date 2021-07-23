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
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"

	"go.opentelemetry.io/build-tools/releaser/internal/common"
)

func RunPrerelease(versioningFile, moduleSetName, fromExistingBranch string, skipMake bool) {

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

	// TODO: this function currently does nothing, but could be updated to add version.go files
	//  to directories.
	if err = p.updateVersionGo(); err != nil {
		log.Fatalf("updateVersionGo failed: %v", err)
	}

	if err = p.updateAllGoModFiles(); err != nil {
		log.Fatalf("updateAllGoModFiles failed: %v", err)
	}

	if skipMake {
		fmt.Println("Skipping 'make lint'...")
	} else {
		if err = p.runMakeLint(); err != nil {
			log.Fatalf("runMakeLint failed: %v", err)
		}
	}

	if err = p.commitChanges(skipMake); err != nil {
		log.Fatalf("commitChanges failed: %v", err)
	}

	fmt.Println("\nPrerelease finished successfully. Now run the following to verify the changes:")
	fmt.Println("\ngit diff main")
	fmt.Println("\nThen, push the changes to upstream.")
}

type prerelease struct {
	common.ModuleSetRelease
	Repo git.Repository
}

func newPrerelease(versioningFilename, modSetToUpdate, repoRoot string) (prerelease, error) {
	modRelease, err := common.NewModuleSetRelease(versioningFilename, modSetToUpdate, repoRoot)
	if err != nil {
		return prerelease{}, fmt.Errorf("error creating new prerelease struct: %v", err)
	}

	repo, err := git.PlainOpen(repoRoot)
	if err != nil {
		return prerelease{}, fmt.Errorf("error getting git.Repository from repo root dir %v: %v", repoRoot, err)
	}

	return prerelease{
		ModuleSetRelease: modRelease,
		Repo: *repo,
	}, nil
}

// verifyGitTagsDoNotAlreadyExist checks if Git tags have already been created that match the specific module tag name
// and version number for the modules being updated. If the tag already exists, an error is returned.
func (p prerelease) verifyGitTagsDoNotAlreadyExist() error {
	var newTags map[string]bool

	for _, newFullTag := range p.ModuleFullTagNames() {
		newTags[newFullTag] = true
	}

	existingTags, err := p.Repo.Tags()
	if err != nil {
		return fmt.Errorf("error getting repo tags: %v", err)
	}

	var errors []errGitTagAlreadyExists

	existingTags.ForEach(func (ref *plumbing.Reference) error {
		tagObj, err := p.Repo.TagObject(ref.Hash())
		if err != nil {
			return fmt.Errorf("error retrieving tag object: %v", err)
		}
		if _, exists := newTags[tagObj.Name]; exists {
			errors = append(errors, errGitTagAlreadyExists{
				gitTag: tagObj.Name,
			})
		}

		return nil
	})

	if len(errors) > 0 {
		return &errGitTagAlreadyExistsSlice{
			errors: errors,
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
	branchNameElements := []string{"pre_release", p.ModSetName, p.ModSetVersion()}
	branchName := strings.Join(branchNameElements, "_")
	fmt.Printf("git checkout -b %v %v\n", branchName, fromExistingBranch)
	cmd := exec.Command("git", "checkout", "-b", branchName, fromExistingBranch)

	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("could not create new branch %v: %v (%v)", branchName, string(output), err)
	}

	return nil
}

// TODO: updateVersionGo may be implemented to update any hard-coded values within version.go files as needed.
func (p prerelease) updateVersionGo() error {
	return nil
}

// runMakeLint runs 'make lint' to automatically update go.sum files.
func (p prerelease) runMakeLint() error {
	fmt.Println("Updating go.sum with 'make lint'...")

	cmd := exec.Command("make", "lint")
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("'make lint' failed: %v (%v)", string(output), err)
	}

	return nil
}

func (p prerelease) commitChanges(skipMake bool) error {
	commitMessage := "Prepare for versions " + p.ModSetVersion()

	// add changes to git
	cmd := exec.Command("git", "add", ".")
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("'git add .' failed: %v (%v)", string(output), err)
	}

	// make ci
	if skipMake {
		fmt.Println("Skipping 'make ci'...")
	} else {
		fmt.Println("Running 'make ci'...")
		cmd = exec.Command("make", "ci")
		if output, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("'make ci' failed: %v (%v)", string(output), err)
		}
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

	for _, modPath := range p.ModSetPaths() {
		oldVersionRegex := filePathToRegex(string(modPath)) + common.SemverRegex
		r, err := regexp.Compile(oldVersionRegex)
		if err != nil {
			return fmt.Errorf("error compiling regex: %v", err)
		}

		newModVersionString := string(modPath) + " " + p.ModSetVersion()

		newGoModFile = r.ReplaceAll(newGoModFile, []byte(newModVersionString))
	}

	// once all module versions have been updated, overwrite the go.mod file
	ioutil.WriteFile(string(modFilePath), newGoModFile, 0644)

	return nil
}

// updateAllGoModFiles updates ALL modules' requires sections to use the newVersion number
// for the modules given in newModPaths.
func (p prerelease) updateAllGoModFiles() error {
	fmt.Println("Updating all module versions in go.mod files...")
	for _, modFilePath := range p.ModPathMap {
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
