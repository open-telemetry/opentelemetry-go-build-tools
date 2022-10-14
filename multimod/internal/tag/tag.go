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
	"errors"
	"fmt"
	"log"
	"os/exec"

	"github.com/go-git/go-git/v5/config"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"go.uber.org/multierr"

	"go.opentelemetry.io/build-tools/internal/repo"
	"go.opentelemetry.io/build-tools/multimod/internal/common"
)

func Run(versioningFile, moduleSetName, commitHash string, deleteModuleSetTags bool, shouldPushTags bool, remote string) {

	repoRoot, err := repo.FindRoot()
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
		if err := t.tagAllModules(nil); err != nil {
			log.Fatalf("unable to tag modules: %v", err)
		}
	}

	if shouldPushTags {
		if err := pushTags(t.ModuleSetRelease.ModuleFullTagNames(), t.Repo, remote); err != nil {
			log.Fatalf("failed to pushTags tags: %v", err)
		}
	}
}

type tagger struct {
	common.ModuleSetRelease
	CommitHash plumbing.Hash
	Repo       *git.Repository
}

func newTagger(versioningFilename, modSetToUpdate, repoRoot, hash string, deleteModuleSetTags bool) (tagger, error) {
	modRelease, err := common.NewModuleSetRelease(versioningFilename, modSetToUpdate, repoRoot)
	if err != nil {
		return tagger{}, fmt.Errorf("error creating tagger struct: %w", err)
	}

	repo, err := git.PlainOpen(repoRoot)
	if err != nil {
		return tagger{}, fmt.Errorf("could not open repo at %v: %w", repoRoot, err)
	}

	fullCommitHash, err := getFullCommitHash(hash, repo)
	if err != nil {
		return tagger{}, fmt.Errorf("could not get full commit hash of given hash %v: %w", hash, err)
	}

	modFullTagNames := modRelease.ModuleFullTagNames()

	if deleteModuleSetTags {
		if err = verifyTagsOnCommit(modFullTagNames, repo, fullCommitHash); err != nil {
			return tagger{}, fmt.Errorf("verifyTagsOnCommit failed: %w", err)
		}
	} else {
		if err = modRelease.CheckGitTagsAlreadyExist(repo); err != nil {
			return tagger{}, fmt.Errorf("CheckGitTagsAlreadyExist failed: %w", err)
		}
	}

	return tagger{
		ModuleSetRelease: modRelease,
		CommitHash:       fullCommitHash,
		Repo:             repo,
	}, nil
}

func verifyTagsOnCommit(modFullTagNames []string, repo *git.Repository, targetCommitHash plumbing.Hash) error {
	var tagsNotOnCommit []string

	for _, tagName := range modFullTagNames {
		tagRef, tagRefErr := repo.Tag(tagName)

		if tagRefErr != nil {
			if errors.Is(tagRefErr, git.ErrTagNotFound) {
				tagsNotOnCommit = append(tagsNotOnCommit, tagName)
				continue
			}
			return fmt.Errorf("unable to fetch git tag ref for %v: %w", tagName, tagRefErr)
		}

		tagObj, tagObjErr := repo.TagObject(tagRef.Hash())
		if tagObjErr != nil {
			return fmt.Errorf("unable to get tag object: %w", tagObjErr)
		}

		tagCommit, tagCommitErr := tagObj.Commit()
		if tagCommitErr != nil {
			return fmt.Errorf("could not get tag object commit: %w", tagCommitErr)
		}

		if targetCommitHash != tagCommit.Hash {
			tagsNotOnCommit = append(tagsNotOnCommit, tagName)
		}
	}

	if len(tagsNotOnCommit) > 0 {
		return &errGitTagsNotOnCommit{
			commitHash: targetCommitHash,
			tagNames:   tagsNotOnCommit,
		}
	}

	return nil
}

func getFullCommitHash(hash string, repo *git.Repository) (plumbing.Hash, error) {
	fullHash, err := repo.ResolveRevision(plumbing.Revision(hash))
	if err != nil {
		return plumbing.ZeroHash, &errCouldNotGetCommitHash{err}
	}

	return *fullHash, nil
}

func (t tagger) deleteModuleSetTags() error {
	modFullTagsToDelete := t.ModuleSetRelease.ModuleFullTagNames()

	if err := deleteTags(modFullTagsToDelete, t.Repo); err != nil {
		return fmt.Errorf("unable to delete module tags: %w", err)
	}

	return nil
}

// deleteTags removes the tags created for a certain version. This func is called to remove newly
// created tags if the new module tagging fails.
func deleteTags(modFullTags []string, repo *git.Repository) error {
	for _, modFullTag := range modFullTags {
		log.Printf("Deleting tag %v\n", modFullTag)

		if err := repo.DeleteTag(modFullTag); err != nil {
			return err
		}
	}
	return nil
}

func (t tagger) tagAllModules(customTagger *object.Signature) error {
	modFullTags := t.ModuleSetRelease.ModuleFullTagNames()

	tagMessage := fmt.Sprintf("Module set %v, Version %v",
		t.ModuleSetRelease.ModSetName, t.ModuleSetRelease.ModSetVersion())

	var addedFullTags []string

	log.Printf("Tagging commit %s:\n", t.CommitHash)

	for _, newFullTag := range modFullTags {
		log.Printf("%v\n", newFullTag)

		var err error
		if customTagger == nil {
			cfg, err2 := t.Repo.Config()
			if err2 != nil {
				err = fmt.Errorf("unable to load repo config: %w", err2)
				if cfg == nil || cfg.Core.Worktree == "" {
					// This is not recoverable, do not panic below.
					return err
				}
			}
			// TODO: figure out how to use go-git and gpg-agent without needing to have decrypted private key material
			// #nosec G204
			cmd := exec.Command("git", "tag", "-a", "-s", "-m", tagMessage, newFullTag, t.CommitHash.String())
			cmd.Dir = cfg.Core.Worktree
			output, err2 := cmd.CombinedOutput()
			if err2 != nil {
				err = fmt.Errorf("unable to create tag: %q: %w", string(output), err2)
			}
		} else {
			_, err = t.Repo.CreateTag(newFullTag, t.CommitHash, &git.CreateTagOptions{
				Message: tagMessage,
				Tagger:  customTagger,
			})
		}

		if err != nil {
			log.Println("error creating a tag, removing all newly created tags...")
			err = fmt.Errorf("git tag failed for %v: %w", newFullTag, err)
			// remove newly created tags to prevent inconsistencies
			if delTagsErr := deleteTags(addedFullTags, t.Repo); delTagsErr != nil {
				return multierr.Combine(err, fmt.Errorf("during handling of the above error, failed to not remove all tags: %w", delTagsErr))
			}

			return err
		}

		addedFullTags = append(addedFullTags, newFullTag)
	}

	return nil
}

func pushTags(tagsToPush []string, repo *git.Repository, remote string) error {

	for _, fullTageName := range tagsToPush {
		tagref, err := repo.Tag(fullTageName)
		if err != nil {
			return fmt.Errorf("unable to fetch git tag ref for %v: %w", fullTageName, err)
		}
		refName := fmt.Sprintf("%s:%s", tagref.Name(), tagref.Name())
		rs := config.RefSpec(refName)
		err = rs.Validate()
		if err != nil {
			return fmt.Errorf("failed validation for refspec %s:%w", rs.String(), err)
		}
		err = repo.Push(&git.PushOptions{
			RefSpecs:   []config.RefSpec{rs},
			RemoteName: remote,
		})
		if err != nil {
			if errors.Is(err, git.NoErrAlreadyUpToDate) {
				log.Printf("tag %s is is already present on remote %s", tagref.Name(), remote)
			} else {
				return fmt.Errorf("error pushing tag %s:%w", tagref.Name(), err)
			}
		}
	}
	return nil
}
