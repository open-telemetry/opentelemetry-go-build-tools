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
	"path/filepath"
	"regexp"
	"strings"

	"github.com/go-git/go-git/v5"

	tools "go.opentelemetry.io/build-tools"
	"go.opentelemetry.io/build-tools/releaser/internal/common"
)

func Run(versioningFile string, moduleSetNames []string, allModuleSets bool, skipModTidy bool) {
	repoRoot, err := tools.FindRepoRoot()
	if err != nil {
		log.Fatalf("unable to find repo root: %v", err)
	}
	log.Printf("Using repo with root at %s\n\n", repoRoot)

	if allModuleSets {
		moduleSetNames, err = common.GetAllModuleSetNames(versioningFile, repoRoot)
		if err != nil {
			log.Fatal("could not automatically get all module set names:", err)
		}
	}

	repo, err := git.PlainOpen(repoRoot)
	if err != nil {
		log.Fatalf("could not open repo at %v: %v", repoRoot, err)
	}

	if err = common.VerifyWorkingTreeClean(repo); err != nil {
		log.Fatal("VerifyWorkingTreeClean failed:", err)
	}

	for _, moduleSetName := range moduleSetNames {
		p, err := newPrerelease(versioningFile, moduleSetName, repoRoot)
		if err != nil {
			log.Fatal("Error creating new prerelease struct:", err)
		}

		log.Printf("===== Module Set: %v =====\n", moduleSetName)

		modSetUpToDate, err := p.checkModuleSetUpToDate(repo)
		if err != nil {
			log.Fatal(err)
		}
		if modSetUpToDate {
			log.Printf("Git tags already exist for module set %v. Skipping...\n", p.ModuleSetRelease.ModSetName)
			continue
		} else {
			log.Printf("Updating versions for module set %v...\n", p.ModuleSetRelease.ModSetName)
		}

		if err = p.updateAllVersionGo(); err != nil {
			log.Fatal("updateVersionGo failed:", err)
		}

		if err = p.updateAllGoModFiles(); err != nil {
			log.Fatal("updateAllGoModFiles failed:", err)
		}

		if skipModTidy {
			log.Println("Skipping go mod tidy...")
		} else {
		}

		if err = p.commitChangesToNewBranch(repo); err != nil {
			log.Fatal("commitChangesToNewBranch failed:", err)
		}
	}

	log.Println(`=========
Prerelease finished successfully. Now run the following to verify the changes:

git diff main

Then, if necessary, commit changes and push to upstream/make a pull request.`)
}

// prerelease holds fields needed to update one module set at a time.
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

// updateAllVersionGo updates the version.go file containing a hardcoded semver version string
// for modules within a set, if the file exists.
func (p prerelease) updateAllVersionGo() error {
	for _, modPath := range p.ModuleSetRelease.ModSetPaths() {
		modFilePath := p.ModuleSetRelease.ModuleVersioning.ModPathMap[modPath]

		versionGoFilePath := filepath.Dir(string(modFilePath))
		versionGoFilePath = filepath.Join(versionGoFilePath, "version.go")

		// check if version.go file exists
		if _, err := os.Stat(versionGoFilePath); err == nil {
			if updateErr := updateVersionGoFile(versionGoFilePath, p.ModuleSetRelease.ModSetVersion()); updateErr != nil {
				return fmt.Errorf("could not update %v: %v", versionGoFilePath, err)
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
// TODO: use the ast package rather than regex to ensure that the string is correctly replaced.
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

// updateAllGoModFiles updates ALL modules' requires sections to use the newVersion number
// for the modules given in newModPaths.
func (p prerelease) updateAllGoModFiles() error {
	modFilePaths := make([]common.ModuleFilePath, 0, len(p.ModuleSetRelease.ModPathMap))

	for _, filePath := range p.ModuleSetRelease.ModPathMap {
		modFilePaths = append(modFilePaths, filePath)
	}

	if err := common.UpdateGoModFiles(
		modFilePaths,
		p.ModuleSetRelease.ModSetPaths(),
		p.ModuleSetRelease.ModSetVersion(),
	); err != nil {
		return fmt.Errorf("could not update all go mod files: %v", err)
	}

	return nil
}

func (p prerelease) commitChangesToNewBranch(repo *git.Repository) error {
	branchNameElements := []string{"prerelease", p.ModuleSetRelease.ModSetName, p.ModuleSetRelease.ModSetVersion()}
	branchName := strings.Join(branchNameElements, "_")

	commitMessage := fmt.Sprintf(
		"Prepare %v for version %v",
		p.ModuleSetRelease.ModSetName,
		p.ModuleSetRelease.ModSetVersion(),
	)

	return common.CommitChangesToNewBranch(branchName, commitMessage, repo)
}

func (p prerelease) checkModuleSetUpToDate(repo *git.Repository) (bool, error) {
	err := p.ModuleSetRelease.CheckGitTagsAlreadyExist(repo)

	switch err.(type) {
	case *common.ErrGitTagsAlreadyExist:
		return true, nil
	case nil:
		return false, nil
	case *common.ErrInconsistentGitTagsExist:
		return false, fmt.Errorf("cannot proceed with inconsistently tagged module set %v: %v", p.ModuleSetRelease.ModSetName, err)
	default:
		return false, fmt.Errorf("unhandled error: %v", err)
	}
}
