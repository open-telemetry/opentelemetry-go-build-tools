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

package sync

import (
	"fmt"
	"log"
	"strings"

	"github.com/go-git/go-git/v5"

	tools "go.opentelemetry.io/build-tools"
	"go.opentelemetry.io/build-tools/releaser/internal/common"
)

func Run(myVersioningFile string, otherVersioningFile string, otherRepoRoot string, otherModuleSetNames []string, allModuleSets bool, skipModTidy bool) {
	myRepoRoot, err := tools.FindRepoRoot()
	if err != nil {
		log.Fatalf("unable to find repo root: %v", err)
	}
	log.Printf("Using repo with root at %s\n\n", myRepoRoot)

	if allModuleSets {
		otherModuleSetNames, err = common.GetAllModuleSetNames(otherVersioningFile, otherRepoRoot)
		if err != nil {
			log.Fatal("could not automatically get all module set names:", err)
		}
	}

	repo, err := git.PlainOpen(myRepoRoot)
	if err != nil {
		log.Fatalf("could not open repo at %v: %v", myRepoRoot, err)
	}

	if err = common.VerifyWorkingTreeClean(repo); err != nil {
		log.Fatal("VerifyWorkingTreeClean failed:", err)
	}

	for _, moduleSetName := range otherModuleSetNames {
		s, err := newSync(myVersioningFile, otherVersioningFile, moduleSetName, myRepoRoot, otherRepoRoot)
		if err != nil {
			log.Fatal("Error creating new sync struct:", err)
		}

		log.Printf("===== Module Set: %v =====\n", moduleSetName)

		if err = s.updateAllGoModFiles(); err != nil {
			log.Fatal("updateAllGoModFiles failed:", err)
		}

		modSetUpToDate, err := s.checkModuleSetUpToDate(repo)
		if err != nil {
			log.Fatal(err)
		}
		if modSetUpToDate {
			log.Println("Module set already up to date. Skipping...")
			continue
		} else {
			log.Println("Updating versions for module set...")
		}

		if skipModTidy {
			log.Println("Skipping go mod tidy...")
		} else {
			if err := common.RunGoModTidy(s.MyModuleVersioning.ModPathMap); err != nil {
				log.Printf("WARNING: failed to run 'go mod tidy': %v\n", err)
			}
		}

		if err = s.commitChangesToNewBranch(repo); err != nil {
			log.Fatal("commitChangesToNewBranch failed:", err)
		}
	}

	log.Println(`=========
Prerelease finished successfully. Now run the following to verify the changes:

git diff main

Then, if necessary, commit changes and push to upstream/make a pull request.`)
}

// sync holds fields needed to update one module set at a time.
type sync struct {
	OtherModuleSetRelease common.ModuleSetRelease
	MyModuleVersioning    common.ModuleVersioning
}

func newSync(myVersioningFilename, otherVersioningFilename, modSetToUpdate, myRepoRoot string, otherRepoRoot string) (sync, error) {
	otherModRelease, err := common.NewModuleSetRelease(otherVersioningFilename, modSetToUpdate, otherRepoRoot)
	if err != nil {
		return sync{}, fmt.Errorf("error creating new sync struct: %v", err)
	}

	myModVersioning, err := common.NewModuleVersioning(myVersioningFilename, myRepoRoot)

	return sync{
		OtherModuleSetRelease: otherModRelease,
		MyModuleVersioning:    myModVersioning,
	}, nil
}

// updateAllGoModFiles updates ALL modules' requires sections to use the newVersion number
// for the modules given in newModPaths.
func (s sync) updateAllGoModFiles() error {
	modFilePaths := make([]common.ModuleFilePath, 0, len(s.MyModuleVersioning.ModPathMap))

	for _, filePath := range s.MyModuleVersioning.ModPathMap {
		modFilePaths = append(modFilePaths, filePath)
	}

	if err := common.UpdateGoModFiles(
		modFilePaths,
		s.OtherModuleSetRelease.ModSetPaths(),
		s.OtherModuleSetRelease.ModSetVersion(),
	); err != nil {
		return fmt.Errorf("could not update all go mod files: %v", err)
	}

	return nil
}

func (s sync) commitChangesToNewBranch(repo *git.Repository) error {
	branchNameElements := []string{"sync", s.OtherModuleSetRelease.ModSetName, s.OtherModuleSetRelease.ModSetVersion()}
	branchName := strings.Join(branchNameElements, "_")

	commitMessage := fmt.Sprintf(
		"Sync repo to use %v with version %v",
		s.OtherModuleSetRelease.ModSetName,
		s.OtherModuleSetRelease.ModSetVersion(),
	)

	return common.CommitChangesToNewBranch(branchName, commitMessage, repo)
}

func (s sync) checkModuleSetUpToDate(repo *git.Repository) (bool, error) {
	worktree, err := common.GetWorktree(repo)
	if err != nil {
		return false, err
	}

	status, err := worktree.Status()
	if err != nil {
		return false, fmt.Errorf("could not get worktree status: %v", err)
	}

	if status.IsClean() {
		return true, nil
	} else {
		return false, nil
	}
}
