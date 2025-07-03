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
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"

	"go.opentelemetry.io/build-tools/internal/repo"
	"go.opentelemetry.io/build-tools/multimod/internal/shared"
)

// Run runs the prerelease process.
func Run(versioningFile string, moduleSetNames []string, skipModTidy bool, commitToDifferentBranch bool) {
	err := run(versioningFile, moduleSetNames, skipModTidy, commitToDifferentBranch)
	if err != nil {
		log.Fatalf("prerelease failed: %v", err)
	}
}

// Used for testing overrides.
var (
	workingTreeClean = shared.VerifyWorkingTreeClean
	findRoot         = repo.FindRoot
	commit           = commitChanges
)

func run(versioningFile string, moduleSetNames []string, skipModTidy bool, commitToDifferentBranch bool) error {
	repoRoot, err := findRoot()
	if err != nil {
		return fmt.Errorf("unable to find repo root: %w", err)
	}
	log.Printf("Using repo with root at %s\n\n", repoRoot)

	// Default to all module sets.
	if len(moduleSetNames) == 0 {
		moduleSetNames, err = shared.GetAllModuleSetNames(versioningFile, repoRoot)
		if err != nil {
			return fmt.Errorf("could not automatically get all module set names: %w", err)
		}
	}

	repo, err := git.PlainOpen(repoRoot)
	if err != nil {
		return fmt.Errorf("could not open repo at %v: %w", repoRoot, err)
	}

	if err = workingTreeClean(repo); err != nil {
		return fmt.Errorf("VerifyWorkingTreeClean failed: %w", err)
	}

	for _, moduleSetName := range moduleSetNames {
		p, err := newPrerelease(versioningFile, moduleSetName, repoRoot)
		if err != nil {
			return fmt.Errorf("prerelease struct: %w", err)
		}

		log.Printf("===== Module Set: %v =====\n", moduleSetName)

		modSetUpToDate, err := p.checkModuleSetUpToDate(repo)
		if err != nil {
			return err
		}
		if modSetUpToDate {
			log.Println("Module set already up to date (git tags already exist). Skipping...")
			continue
		}
		log.Println("Updating versions for module set...")

		if err = p.updateAllVersionGo(); err != nil {
			return fmt.Errorf("updateAllVersionGo failed: %w", err)
		}

		if err = p.updateAllGoModFiles(); err != nil {
			return fmt.Errorf("updateAllGoModFiles failed: %w", err)
		}

		if skipModTidy {
			log.Println("Skipping 'go mod tidy'...")
		} else {
			if err = shared.RunGoModTidy(p.ModPathMap); err != nil {
				return fmt.Errorf("could not run Go Mod Tidy: %w", err)
			}
		}

		if err = commit(p.ModuleSetRelease, commitToDifferentBranch, repo); err != nil {
			return fmt.Errorf("commitChangesToNewBranch failed: %w", err)
		}
	}

	log.Println(`=========
Prerelease finished successfully. Now checkout the new branch(es) and verify the changes.

Then, if necessary, commit changes and push to upstream/make a pull request.`)
	return nil
}

// prerelease holds fields needed to update one module set at a time.
type prerelease struct {
	shared.ModuleSetRelease
}

func newPrerelease(versioningFilename, modSetToUpdate, repoRoot string) (prerelease, error) {
	modRelease, err := shared.NewModuleSetRelease(versioningFilename, modSetToUpdate, repoRoot)
	if err != nil {
		return prerelease{}, fmt.Errorf("error creating new prerelease struct: %w", err)
	}

	return prerelease{
		ModuleSetRelease: modRelease,
	}, nil
}

func (p prerelease) checkModuleSetUpToDate(repo *git.Repository) (bool, error) {
	err := p.CheckGitTagsAlreadyExist(repo)
	if err != nil {
		if errors.As(err, &shared.ErrGitTagsAlreadyExist{}) {
			return true, nil
		}
		if errors.As(err, &shared.ErrInconsistentGitTagsExist{}) {
			return false, fmt.Errorf("cannot proceed with inconsistently tagged module set %v: %w", p.ModSetName, err)
		}
		return false, fmt.Errorf("unhandled error: %w", err)
	}

	return false, nil
}

// updateAllVersionGo updates the version.go file containing a hardcoded semver version string
// for modules within a set, if the file exists.
func (p prerelease) updateAllVersionGo() error {
	var err error
	for _, modPath := range p.ModSetPaths() {
		modFilePath := p.ModPathMap[modPath]
		root := filepath.Dir(string(modFilePath))

		vRefs := p.ModInfoMap[modPath].VersionRefs(root)
		if len(vRefs) == 0 {
			vRefs = defaultFileRefs(root)
		}

		for _, vRef := range vRefs {
			e := updateVersionGoFile(vRef, p.ModSetVersion())
			err = errors.Join(err, e)
		}
	}
	return err
}

func defaultFileRefs(root string) []string {
	path := filepath.Join(root, "version.go")
	_, err := os.Stat(path)
	if err != nil {
		if !os.IsNotExist(err) {
			log.Printf("Warning: could not check existence of %v: %v\n", path, err)
		}
		// The file does not exist, or we cannot check its existence.
		return nil
	}
	return []string{path}
}

var verRegex = regexp.MustCompile(shared.SemverRegexNumberOnly)

// updateVersionGoFile updates all versions within the file at path to use the
// new version number ver.
func updateVersionGoFile(path string, ver string) error {
	// TODO: There is a potential improvement is to use an AST package rather than regex
	// to perform replacement.
	log.Printf("... Updating version references in %s to %s\n", path, ver)

	data, err := os.ReadFile(filepath.Clean(path))
	if err != nil {
		return fmt.Errorf("error reading version.go file %v: %w", path, err)
	}

	v := strings.TrimPrefix(ver, "v")
	data = verRegex.ReplaceAll(data, []byte(v))

	// Overwrite filePath.
	if err := os.WriteFile(path, data, 0600); err != nil {
		return fmt.Errorf("error overwriting %s file: %w", path, err)
	}

	return nil
}

// updateAllGoModFiles updates ALL modules' requires sections to use the newVersion number
// for the modules given in newModPaths.
func (p prerelease) updateAllGoModFiles() error {
	modFilePaths := make([]shared.ModuleFilePath, 0, len(p.ModPathMap))

	for _, filePath := range p.AllModPathMap {
		modFilePaths = append(modFilePaths, filePath)
	}

	var newModRefs []shared.ModuleRef
	ver := p.ModSetVersion()
	for _, mod := range p.ModSetPaths() {
		newModRefs = append(newModRefs, shared.ModuleRef{Path: mod, Version: ver})
	}
	if err := shared.UpdateGoModFiles(modFilePaths, newModRefs); err != nil {
		return fmt.Errorf("could not update all go mod files: %w", err)
	}

	return nil
}

func commitChanges(msr shared.ModuleSetRelease, commitToDifferentBranch bool, repo *git.Repository) error {
	commitMessage := fmt.Sprintf("Prepare %v for version %v", msr.ModSetName, msr.ModSetVersion())

	var hash plumbing.Hash
	var err error
	if commitToDifferentBranch {
		branchNameElements := []string{"prerelease", msr.ModSetName, msr.ModSetVersion()}
		branchName := strings.Join(branchNameElements, "_")
		hash, err = shared.CommitChangesToNewBranch(branchName, commitMessage, repo, nil)
	} else {
		hash, err = shared.CommitChanges(commitMessage, repo, nil)
	}
	if err != nil {
		return err
	}
	log.Printf("Commit successful. Hash of commit: %s\n", hash)
	return nil
}
