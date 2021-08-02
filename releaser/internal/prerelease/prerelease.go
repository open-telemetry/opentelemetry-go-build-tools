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
	"log"
	"strings"

	tools "go.opentelemetry.io/build-tools"
	"go.opentelemetry.io/build-tools/releaser/internal/common"
)

func Run(versioningFile string, moduleSetName string, skipMake bool, noCommit bool) {
	repoRoot, err := tools.FindRepoRoot()
	if err != nil {
		log.Fatalf("unable to find repo root: %v", err)
	}
	log.Printf("Using repo with root at %s\n", repoRoot)

	p, err := newPrerelease(versioningFile, moduleSetName, repoRoot)
	if err != nil {
		log.Fatalf("Error creating new prerelease struct: %v", err)
	}

	if err = p.ModuleSetRelease.VerifyGitTagsDoNotAlreadyExist(); err != nil {
		log.Fatalf("VerifyGitTagsDoNotAlreadyExist failed: %v", err)
	}

	if err = common.VerifyWorkingTreeClean(p.ModuleSetRelease.Repo); err != nil {
		log.Fatalf("verifyWorkingTreeClean failed: %v", err)
	}

	branchNameElements := []string{"prerelease", p.ModuleSetRelease.ModSetName, p.ModuleSetRelease.ModSetVersion()}
	branchName := strings.Join(branchNameElements, "_")

	if err = common.CreateGitBranch(branchName, p.ModuleSetRelease.Repo); err != nil {
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
	}
	
	if noCommit {
		log.Println("Changes have not been added nor committed.")
	} else {
		if err = p.commitChanges(); err != nil {
			log.Fatalf("commitChanges failed: %v", err)
		}
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

// TODO: updateVersionGo may be implemented to update any hard-coded values within version.go files as needed.
func (p prerelease) updateVersionGo() error {
	return nil
}

// updateAllGoModFiles updates ALL modules' requires sections to use the newVersion number
// for the modules given in newModPaths.
func (p prerelease) updateAllGoModFiles() error {
	modFilePaths := make([]common.ModuleFilePath, 0, len(p.ModuleSetRelease.ModPathMap))

	for _, filePath := range p.ModuleSetRelease.ModPathMap {
		modFilePaths = append(modFilePaths, filePath)
	}

	if err := common.UpdateAllGoModFiles(
		modFilePaths,
		p.ModuleSetRelease.ModSetPaths(),
		p.ModuleSetRelease.ModSetVersion(),
	); err != nil {
		return fmt.Errorf("could not update all go mod files: %v", err)
	}

	return nil
}

func (p prerelease) commitChanges() error {
	commitMessage := fmt.Sprintf(
		"Prepare %v for version %v",
		p.ModuleSetRelease.ModSetName,
		p.ModuleSetRelease.ModSetVersion(),
	)

	return common.CommitChanges(commitMessage, p.ModuleSetRelease.Repo)
}
