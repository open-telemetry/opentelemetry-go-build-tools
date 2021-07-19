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

package tag

import (
	"fmt"
	"log"
	"os/exec"
	"strings"

	"go.opentelemetry.io/build-tools/releaser/internal/versions"
)

func RunTag(versioningFile, moduleSetName, commitHash string, deleteModuleSetTags bool) {

	repoRoot, err := versions.ChangeToRepoRoot()
	if err != nil {
		log.Fatalf("unable to change to repo root: %v", err)
	}

	t, err := newTagger(versioningFile, moduleSetName, repoRoot, commitHash, deleteModuleSetTags)
	if err != nil {
		log.Fatalf("Error creating new tagger struct: %v", err)
	}

	// if delete-module-set-tags is specified, then delete all newModTagNames
	// whose versions match the one in the versioning file. Otherwise, tag all
	// modules in the given set.
	if deleteModuleSetTags {
		if err := t.deleteModuleSetTags(); err != nil {
			log.Fatalf("Error deleting tags for the specified module set: %v", err)
		}

		fmt.Println("Successfully deleted module tags")
	} else {
		if err := t.tagAllModules(); err != nil {
			log.Fatalf("unable to tag modules: %v", err)
		}
	}
}

type tagger struct {
	versions.ModuleSetRelease
	CommitHash string
}

func newTagger(versioningFilename, modSetToUpdate, repoRoot, hash string, deleteModuleSetTags bool) (tagger, error) {
	modRelease, err := versions.NewModuleSetRelease(versioningFilename, modSetToUpdate, repoRoot)
	if err != nil {
		return tagger{}, fmt.Errorf("error creating prerelease struct: %v", err)
	}

	var fullCommitHash string
	if !deleteModuleSetTags {
		fullCommitHash, err = getFullCommitHash(hash)
		if err != nil {
			return tagger{}, fmt.Errorf("could not get full commit hash of given hash %v: %v", hash, err)
		}
	}

	return tagger{
		ModuleSetRelease: modRelease,
		CommitHash:       fullCommitHash,
	}, nil
}

func getFullCommitHash(hash string) (string, error) {
	fmt.Printf("git rev-parse --quiet --verify %v\n", hash)
	cmd := exec.Command("git", "rev-parse", "--quiet", "--verify", hash)

	// output stores the complete SHA1 of the commit hash
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("could not retrieve commit hash %v: %v", hash, err)
	}
	if output == nil || len(output) == 0 {
		return "", fmt.Errorf("commit hash not found with 'git rev-parse --quiet --verify %v'", hash)
	}

	SHA := strings.TrimSpace(string(output))

	cmd = exec.Command("git", "merge-base", SHA, "HEAD")
	// output should match SHA
	output, err = cmd.Output()
	if err != nil {
		return "", fmt.Errorf("command 'git merge-base %v HEAD' failed: %v", SHA, err)
	}
	if strings.TrimSpace(string(output)) != SHA {
		return "", fmt.Errorf("commit %v (complete SHA: %v) not found on this branch", hash, SHA)
	}

	return SHA, nil
}

func (t tagger) deleteModuleSetTags() error {
	modFullTagsToDelete := t.ModuleFullTagNames()

	if err := t.deleteTags(modFullTagsToDelete); err != nil {
		return fmt.Errorf("unable to delete module tags: %v", err)
	}

	return nil
}

// deleteTags removes the tags created for a certain version. This func is called to remove newly
// created tags if the new module tagging fails.
func (t tagger) deleteTags(modFullTags []string) error {
	for _, modFullTag := range modFullTags {
		fmt.Printf("Deleting tag %v\n", modFullTag)
		cmd := exec.Command("git", "tag", "-d", modFullTag)
		if output, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("could not delete tag %v:\n%v (%v)", modFullTag, string(output), err)
		}
	}
	return nil
}

func (t tagger) tagAllModules() error {
	modFullTags := t.ModuleFullTagNames()

	var addedFullTags []string

	fmt.Printf("Tagging commit %v:\n", t.CommitHash)

	for _, newFullTag := range modFullTags {
		fmt.Printf("%v\n", newFullTag)

		cmd := exec.Command("git", "tag", "-a", newFullTag, "-s", "-m", "Version "+newFullTag, t.CommitHash)
		if output, err := cmd.CombinedOutput(); err != nil {
			fmt.Println("error creating a tag, removing all newly created tags...")

			// remove newly created tags to prevent inconsistencies
			if delTagsErr := t.deleteTags(addedFullTags); delTagsErr != nil {
				return fmt.Errorf("git tag failed for %v:\n%v (%v).\nCould not remove all tags: %v",
					newFullTag, string(output), err, delTagsErr,
				)
			}

			return fmt.Errorf("git tag failed for %v:\n%v (%v)", newFullTag, string(output), err)
		}

		addedFullTags = append(addedFullTags, newFullTag)
	}

	return nil
}
