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

	"github.com/go-git/go-git/v5/plumbing"

	"go.opentelemetry.io/build-tools/internal/repo"
	"go.opentelemetry.io/build-tools/multimod/internal/common"
)

func Run(versioningFile string, moduleSetNames []string, allModuleSets bool, skipModTidy bool, commitToDifferentBranch bool) {
	repoRoot, err := repo.FindRoot()
	if err != nil {
		log.Fatalf("unable to find repo root: %v", err)
	}
	log.Printf("Using repo with root at %s\n\n", repoRoot)

	if allModuleSets {
		moduleSetNames, err = common.GetAllModuleSetNames(versioningFile, repoRoot)
		if err != nil {
			log.Fatalf("could not automatically get all module set names: %v", err)
		}
	}

	repo, err := git.PlainOpen(repoRoot)
	if err != nil {
		log.Fatalf("could not open repo at %v: %v", repoRoot, err)
	}

	if err = common.VerifyWorkingTreeClean(repo); err != nil {
		log.Fatalf("VerifyWorkingTreeClean failed: %v", err)
	}

	for _, moduleSetName := range moduleSetNames {
		p, err := newPrerelease(versioningFile, moduleSetName, repoRoot)
		if err != nil {
			log.Fatalf("Error creating new prerelease struct: %v", err)
		}

		log.Printf("===== Module Set: %v =====\n", moduleSetName)

		modSetUpToDate, err := p.checkModuleSetUpToDate(repo)
		if err != nil {
			log.Fatal(err)
		}
		if modSetUpToDate {
			log.Println("Module set already up to date (git tags already exist). Skipping...")
			continue
		} else {
			log.Println("Updating versions for module set...")
		}

		if err = p.updateAllVersionGo(); err != nil {
			log.Fatalf("updateAllVersionGo failed: %v", err)
		}

		if err = p.updateAllGoModFiles(); err != nil {
			log.Fatalf("updateAllGoModFiles failed: %v", err)
		}

		if skipModTidy {
			log.Println("Skipping 'go mod tidy'...")
		} else {
			if err = common.RunGoModTidy(p.ModuleSetRelease.ModuleVersioning.ModPathMap); err != nil {
				log.Fatal("could not run Go Mod Tidy: ", err)
			}
		}

		if err = commitChanges(p.ModuleSetRelease, commitToDifferentBranch, repo); err != nil {
			log.Fatalf("commitChangesToNewBranch failed: %v", err)
		}
	}

	log.Println(`=========
Prerelease finished successfully. Now checkout the new branch(es) and verify the changes.

Then, if necessary, commit changes and push to upstream/make a pull request.`)
}

// prerelease holds fields needed to update one module set at a time.
type prerelease struct {
	common.ModuleSetRelease
}

func newPrerelease(versioningFilename, modSetToUpdate, repoRoot string) (prerelease, error) {
	modRelease, err := common.NewModuleSetRelease(versioningFilename, modSetToUpdate, repoRoot)
	if err != nil {
		return prerelease{}, fmt.Errorf("error creating new prerelease struct: %w", err)
	}

	return prerelease{
		ModuleSetRelease: modRelease,
	}, nil
}

func (p prerelease) checkModuleSetUpToDate(repo *git.Repository) (bool, error) {
	err := p.ModuleSetRelease.CheckGitTagsAlreadyExist(repo)
	if err != nil {
		if errors.As(err, &common.ErrGitTagsAlreadyExist{}) {
			return true, nil
		}
		if errors.As(err, &common.ErrInconsistentGitTagsExist{}) {
			return false, fmt.Errorf("cannot proceed with inconsistently tagged module set %v: %w", p.ModuleSetRelease.ModSetName, err)
		}
		return false, fmt.Errorf("unhandled error: %w", err)
	}

	return false, nil
}

// updateAllVersionGo updates the version.go file containing a hardcoded semver version string
// for modules within a set, if the file exists.
func (p prerelease) updateAllVersionGo() error {
	for _, modPath := range p.ModuleSetRelease.ModSetPaths() {
		modFilePath := p.ModuleSetRelease.ModuleVersioning.ModPathMap[modPath]

		versionGoDir := filepath.Dir(string(modFilePath))
		versionGoFilePath := filepath.Join(versionGoDir, "version.go")

		// check if version.go file exists
		_, err := os.Stat(versionGoFilePath)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			} else {
				return fmt.Errorf("could not check existence of %v: %w", versionGoFilePath, err)
			}
		}
		if err = updateVersionGoFile(versionGoFilePath, p.ModuleSetRelease.ModSetVersion()); err != nil {
			return fmt.Errorf("could not update %v: %w", versionGoFilePath, err)
		}

	}
	return nil
}

// updateVersionGoFile updates one version.go file.
// TODO: a potential improvement is to use an AST package rather than regex to perform replacement.
func updateVersionGoFile(filePath string, newVersion string) error {
	if !strings.HasSuffix(filePath, "version.go") {
		return errors.New("cannot update file passed that does not end with version.go")
	}
	log.Printf("... Updating file %v\n", filePath)

	newVersionGoFile, err := os.ReadFile(filepath.Clean(filePath))
	if err != nil {
		panic(err)
	}

	oldVersionRegex := common.SemverRegexNumberOnly
	r, err := regexp.Compile(oldVersionRegex)
	if err != nil {
		return fmt.Errorf("error compiling regex: %w", err)
	}

	newVersionNumberOnly := strings.TrimPrefix(newVersion, "v")

	newVersionGoFile = r.ReplaceAll(newVersionGoFile, []byte(newVersionNumberOnly))

	// overwrite the version.go file
	if err := os.WriteFile(filePath, newVersionGoFile, 0600); err != nil {
		return fmt.Errorf("error overwriting go.mod file: %w", err)
	}

	return nil
}

// updateAllGoModFiles updates ALL modules' requires sections to use the newVersion number
// for the modules given in newModPaths.
func (p prerelease) updateAllGoModFiles() error {
	modFilePaths := make([]common.ModuleFilePath, 0, len(p.ModuleSetRelease.ModuleVersioning.ModPathMap))

	for _, filePath := range p.ModuleSetRelease.ModuleVersioning.ModPathMap {
		modFilePaths = append(modFilePaths, filePath)
	}

	if err := common.UpdateGoModFiles(modFilePaths, p.ModuleSetRelease.ModSetPaths(), p.ModuleSetRelease.ModSetVersion()); err != nil {
		return fmt.Errorf("could not update all go mod files: %w", err)
	}

	return nil
}

func commitChanges(msr common.ModuleSetRelease, commitToDifferentBranch bool, repo *git.Repository) error {
	commitMessage := fmt.Sprintf("Prepare %v modules for version %v", msr.ModSetName, msr.ModSetVersion())

	var hash plumbing.Hash
	var err error
	if commitToDifferentBranch {
		branchNameElements := []string{"prerelease", msr.ModSetName, msr.ModSetVersion()}
		branchName := strings.Join(branchNameElements, "_")
		hash, err = common.CommitChangesToNewBranch(branchName, commitMessage, repo, nil)
	} else {
		hash, err = common.CommitChanges(commitMessage, repo, nil)
	}
	if err != nil {
		return err
	}
	log.Printf("Commit successful. Hash of commit: %s\n", hash)
	return nil
}
