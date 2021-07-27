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

	tools "go.opentelemetry.io/build-tools"
	"go.opentelemetry.io/build-tools/releaser/internal/common"
)

func Run(versioningFile string, moduleSetName string, skipMake bool) {
	repoRoot, err := tools.FindRepoRoot()
	if err != nil {
		log.Fatalf("unable to change to repo root: %v", err)
	}
	log.Printf("Using repo with root at %s\n", repoRoot)

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

	if err = p.createPrereleaseBranch(); err != nil {
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
		log.Println("Skipping 'make lint' and 'make ci'")
	} else {
		if err = p.runMakeLint(); err != nil {
			log.Fatalf("runMakeLint failed: %v", err)
		}
		if err = p.runMakeCI(); err != nil {
			log.Fatalf("runMakeCI failed: %v", err)
		}
	}

	if err = p.commitChanges(); err != nil {
		log.Fatalf("commitChanges failed: %v", err)
	}

	log.Println("\nPrerelease finished successfully. Now run the following to verify the changes:")
	log.Println("\ngit diff main")
	log.Println("\nThen, push the changes to upstream.")
}

type prerelease struct {
	common.ModuleSetRelease
}

func newPrerelease(versioningFilename, modSetToUpdate, repoRoot string) (prerelease, error) {
	modRelease, err := common.NewModuleSetRelease(versioningFilename, modSetToUpdate, repoRoot)
	if err != nil {
		return prerelease{}, fmt.Errorf("error creating new prerelease struct: %v", err)
	}

	return prerelease{
		ModuleSetRelease: modRelease,
	}, nil
}

// verifyGitTagsDoNotAlreadyExist checks if Git tags have already been created that match the specific module tag name
// and version number for the modules being updated. If the tag already exists, an error is returned.
func (p prerelease) verifyGitTagsDoNotAlreadyExist() error {
	newTags := make(map[string]bool)

	modFullTags := p.ModuleSetRelease.ModuleFullTagNames()

	for _, newFullTag := range modFullTags {
		newTags[newFullTag] = true
	}

	existingTags, err := p.ModuleSetRelease.Repo.Tags()
	if err != nil {
		return fmt.Errorf("error getting repo tags: %v", err)
	}

	var errors []errGitTagAlreadyExists

	_ = existingTags.ForEach(func(ref *plumbing.Reference) error {
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

// verifyWorkingTreeClean returns nil if the working tree is clean or an error if not.
func (p prerelease) verifyWorkingTreeClean() error {
	worktree, err := p.ModuleSetRelease.Repo.Worktree()
	if err != nil {
		return &errGetWorktreeFailed{reason: err}
	}

	status, err := worktree.Status()
	if err != nil {
		return fmt.Errorf("could not get worktree status: %v", err)
	}

	if !status.IsClean() {
		return &errWorkingTreeNotClean{}
	}

	return nil
}

func (p prerelease) createPrereleaseBranch() error {
	branchNameElements := []string{"prerelease", p.ModuleSetRelease.ModSetName, p.ModuleSetRelease.ModSetVersion()}
	branchName := strings.Join(branchNameElements, "_")

	worktree, err := p.ModuleSetRelease.Repo.Worktree()
	if err != nil {
		return &errGetWorktreeFailed{reason: err}
	}

	checkoutOptions := &git.CheckoutOptions{
		Branch: plumbing.ReferenceName(branchName),
		Create: true,
		Keep:   true,
	}

	log.Printf("git checkout -b %v\n", branchName)
	if err = worktree.Checkout(checkoutOptions); err != nil {
		return fmt.Errorf("could not check out new branch: %v", err)
	}

	return nil
}

// TODO: updateVersionGo may be implemented to update any hard-coded values within version.go files as needed.
func (p prerelease) updateVersionGo() error {
	return nil
}

// runMakeLint runs 'make lint' to automatically update go.sum files.
func (p prerelease) runMakeLint() error {
	log.Println("Updating go.sum with 'make lint'")

	cmd := exec.Command("make", "lint")
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("'make lint' failed: %v (%v)", string(output), err)
	}

	return nil
}

func (p prerelease) runMakeCI() error {
	log.Println("Running 'make ci'")

	cmd := exec.Command("make", "ci")
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("'make ci' failed: %v (%v)", string(output), err)
	}

	return nil
}

func (p prerelease) commitChanges() error {
	commitMessage := "Prepare for version " + p.ModuleSetRelease.ModSetVersion()

	// commit changes to git
	log.Printf("Committing changes to git with message '%v'\n", commitMessage)

	worktree, err := p.ModuleSetRelease.Repo.Worktree()
	if err != nil {
		return &errGetWorktreeFailed{reason: err}
	}

	commitOptions := &git.CommitOptions{}

	hash, err := worktree.Commit(commitMessage, commitOptions)
	if err != nil {
		return fmt.Errorf("could not commit changes to git: %v", err)
	}

	log.Printf("Commit successful. Hash of commit: %s\n", hash)

	return nil
}

func (p prerelease) gitAddFile(modFilePath common.ModuleFilePath) error {
	worktree, err := p.ModuleSetRelease.Repo.Worktree()
	if err != nil {
		return &errGetWorktreeFailed{reason: err}
	}

	log.Printf("git add %s", modFilePath)
	if _, err = worktree.Add(string(modFilePath)); err != nil {
		return &errGitAddFailed{reason: err}
	}

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
		oldVersionRegex := filePathToRegex(string(modPath)) + common.SemverRegex
		r, err := regexp.Compile(oldVersionRegex)
		if err != nil {
			return fmt.Errorf("error compiling regex: %v", err)
		}

		newModVersionString := string(modPath) + " " + p.ModuleSetRelease.ModSetVersion()

		newGoModFile = r.ReplaceAll(newGoModFile, []byte(newModVersionString))
	}

	// once all module versions have been updated, overwrite the go.mod file
	err = ioutil.WriteFile(string(modFilePath), newGoModFile, 0644)
	if err != nil {
		return fmt.Errorf("error writing file: %v", err)
	}

	return nil
}

// updateAllGoModFiles updates ALL modules' requires sections to use the newVersion number
// for the modules given in newModPaths.
func (p prerelease) updateAllGoModFiles() error {
	log.Println("Updating all module versions in go.mod files...")
	for _, modFilePath := range p.ModuleSetRelease.ModPathMap {
		if err := p.updateGoModVersions(modFilePath); err != nil {
			return fmt.Errorf("could not update module versions in file %v: %v", modFilePath, err)
		}
		if err := p.gitAddFile(modFilePath); err != nil {
			return fmt.Errorf("error adding file to git: %v", err)
		}
	}
	return nil
}

func filePathToRegex(fpath string) string {
	quotedMeta := regexp.QuoteMeta(fpath)
	replacedSlashes := strings.Replace(quotedMeta, string(filepath.Separator), `\/`, -1)
	return replacedSlashes
}
