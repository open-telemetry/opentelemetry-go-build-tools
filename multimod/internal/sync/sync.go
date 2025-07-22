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
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/go-git/go-git/v5"
	"golang.org/x/sync/errgroup"

	"github.com/hashicorp/go-retryablehttp"

	"go.opentelemetry.io/build-tools/internal/repo"
	"go.opentelemetry.io/build-tools/multimod/internal/shared"
)

// Run runs the synchronization process.
func Run(myVersioningFile string, otherVersioningFile string, otherRepoRoot string, otherModuleSetNames []string, otherVersionCommit string, allModuleSets bool, skipModTidy bool) {
	myRepoRoot, err := repo.FindRoot()
	if err != nil {
		log.Fatalf("unable to find repo root: %v", err)
	}
	log.Printf("Using repo with root at %s\n\n", myRepoRoot)

	if allModuleSets {
		otherModuleSetNames, err = shared.GetAllModuleSetNames(otherVersioningFile, otherRepoRoot)
		if err != nil {
			log.Fatalf("could not automatically get all module set names: %v", err)
		}
	}

	repo, err := git.PlainOpen(myRepoRoot)
	if err != nil {
		log.Fatalf("could not open repo at %v: %v", myRepoRoot, err)
	}

	if err = shared.VerifyWorkingTreeClean(repo); err != nil {
		log.Fatalf("VerifyWorkingTreeClean failed: %v", err)
	}

	for _, moduleSetName := range otherModuleSetNames {
		s, err := newSync(myVersioningFile, otherVersioningFile, moduleSetName, myRepoRoot, otherVersionCommit)
		if err != nil {
			log.Fatalf("error creating new sync struct: %v", err)
		}

		log.Printf("===== Module Set: %v =====\n", moduleSetName)

		if err = s.updateAllGoModFiles(); err != nil {
			log.Fatalf("updateAllGoModFiles failed: %v", err)
		}

		modSetUpToDate, err := checkModuleSetUpToDate(repo)
		if err != nil {
			log.Fatal(err)
		}
		if modSetUpToDate {
			log.Println("Module set already up to date. Skipping...")
			continue
		}
		log.Println("Updating versions for module set...")

		if skipModTidy {
			log.Println("Skipping go mod tidy...")
		} else {
			if err := shared.RunGoModTidy(s.MyModuleVersioning.ModPathMap); err != nil {
				log.Printf("WARNING: failed to run 'go mod tidy': %v\n", err)
			}
		}
	}

	log.Println(`=========
Prerelease finished successfully. Now run the following to verify the changes:

git diff main

Then, if necessary, commit changes and push to upstream/make a pull request.`)
}

// sync holds fields needed to update one module set at a time.
type sync struct {
	OtherModuleSetName       string
	OtherModuleVersionCommit string
	OtherModuleSet           shared.ModuleSet
	MyModuleVersioning       shared.ModuleVersioning
	client                   *http.Client
}

func newSync(myVersioningFilename, otherVersioningFilename, modSetToUpdate, myRepoRoot string, otherVersionCommit string) (sync, error) {
	otherModuleSet, err := shared.GetModuleSet(modSetToUpdate, otherVersioningFilename)
	if err != nil {
		return sync{}, fmt.Errorf("error creating new sync struct: %w", err)
	}

	// when syncing, always sync deps for modules, including the excluded-modules
	myModVersioning, err := shared.NewModuleVersioningWithIgnoreExcluded(myVersioningFilename, myRepoRoot, true)
	if err != nil {
		return sync{}, fmt.Errorf("could not get my ModuleVersioning: %w", err)
	}

	return sync{
		OtherModuleSetName:       modSetToUpdate,
		OtherModuleSet:           otherModuleSet,
		MyModuleVersioning:       myModVersioning,
		OtherModuleVersionCommit: otherVersionCommit,
		client:                   retryablehttp.NewClient().StandardClient(),
	}, nil
}

func (s sync) parseVersionInfo(pkg, tag string) (string, error) {
	res, err := s.client.Get(fmt.Sprintf("https://proxy.golang.org/%s/@v/%s.info", pkg, tag))
	if err != nil {
		return "", err
	}
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return "", nil
	}

	if res.StatusCode >= 400 {
		return "", fmt.Errorf("request to proxy for package %q at version %q failed with status %d (%s): %s",
			pkg, tag, res.StatusCode, res.Status, string(body))
	}

	var data struct{ Version string }
	err = json.Unmarshal(body, &data)
	if err != nil {
		return "", fmt.Errorf("failed to unmarshal response for package %q at version %q: %w", pkg, tag, err)
	}

	return fmt.Sprint(data.Version), err
}

// updateAllGoModFiles updates ALL modules' requires sections to use the newVersion number
// for the modules given in newModPaths.
func (s sync) updateAllGoModFiles() error {
	modFilePaths := make([]shared.ModuleFilePath, 0, len(s.MyModuleVersioning.ModPathMap))

	for _, filePath := range s.MyModuleVersioning.ModPathMap {
		modFilePaths = append(modFilePaths, filePath)
	}

	g := new(errgroup.Group)
	g.SetLimit(8)
	modRefsChan := make(chan shared.ModuleRef, len(s.OtherModuleSet.Modules))
	defer close(modRefsChan)

	for _, mod := range s.OtherModuleSet.Modules {
		g.Go(func() error {
			ver := s.OtherModuleSet.Version
			if s.OtherModuleVersionCommit != "" {
				version, err := s.parseVersionInfo(string(mod), s.OtherModuleVersionCommit)
				if err != nil {
					return err
				}
				ver = version
			}
			log.Printf("Version for module %q: %s\n", mod, ver)
			modRefsChan <- shared.ModuleRef{Path: mod, Version: ver}
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return fmt.Errorf("error while fetching module versions: %w", err)
	}

	var newModRefs []shared.ModuleRef
	// Collect all module references from the channel.
	for i := 0; i < len(s.OtherModuleSet.Modules); i++ {
		modRef := <-modRefsChan
		newModRefs = append(newModRefs, modRef)
	}

	if err := shared.UpdateGoModFiles(
		modFilePaths,
		newModRefs,
	); err != nil {
		return fmt.Errorf("could not update all go mod files: %w", err)
	}

	return nil
}

func checkModuleSetUpToDate(repo *git.Repository) (bool, error) {
	worktree, err := shared.GetWorktree(repo)
	if err != nil {
		return false, err
	}

	status, err := worktree.Status()
	if err != nil {
		return false, fmt.Errorf("could not get worktree status: %w", err)
	}

	return status.IsClean(), nil
}
