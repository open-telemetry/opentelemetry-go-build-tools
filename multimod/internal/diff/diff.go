// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

// Package diff provides functionality to check if files in a module set have changed.
package diff

import (
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"

	"go.opentelemetry.io/build-tools/multimod/internal/common"
)

// Client is an interface for a git client. It is used to abstract away the
// implementation details of the git client, allowing for easier testing and
// mocking.
type Client interface {
	HeadCommit(r *git.Repository) (*object.Commit, error)
	TagCommit(r *git.Repository, tag string) (*object.Commit, error)
	FilesChanged(headCommit *object.Commit, tagCommit *object.Commit, prefix string, suffix string) ([]string, error)
}

// GitClient handles interactions with git.
type GitClient struct{}

// HeadCommit returns the commit object for the HEAD of the repository.
func (g GitClient) HeadCommit(r *git.Repository) (*object.Commit, error) {
	headRef, err := r.Head()
	if err != nil {
		return nil, err
	}
	return r.CommitObject(headRef.Hash())
}

// TagCommit returns the commit object for a given tag.
func (g GitClient) TagCommit(r *git.Repository, tag string) (*object.Commit, error) {
	tagRef, err := r.Tag(tag)
	if err != nil {
		return nil, err
	}

	o, err := r.TagObject(tagRef.Hash())
	if err != nil {
		return nil, fmt.Errorf("tag object error %s %w", tagRef.Hash().String(), err)
	}
	return r.CommitObject(o.Target)
}

// FilesChanged returns a list of files that have changed between two commits.
func (g GitClient) FilesChanged(headCommit *object.Commit, tagCommit *object.Commit, prefix string, suffix string) ([]string, error) {
	changedFiles := []string{}
	p, err := headCommit.Patch(tagCommit)
	if err != nil {
		return changedFiles, err
	}

	for _, f := range p.FilePatches() {
		from, to := f.Files()
		if from != nil && strings.HasSuffix(from.Path(), suffix) && strings.HasPrefix(from.Path(), prefix) {
			changedFiles = append(changedFiles, from.Path())
			continue
		}
		if to != nil && strings.HasSuffix(to.Path(), suffix) && strings.HasPrefix(to.Path(), prefix) {
			changedFiles = append(changedFiles, to.Path())
		}
	}
	return changedFiles, nil
}

// normalizeVersion ensures the version is prefixed with a `v`. The missing v prefix in
// the version has caused problems in the collector repo. This logic was originally implemented
// in the Makefile.
func normalizeVersion(ver string) string {
	if strings.HasPrefix(ver, "v") {
		return ver
	}
	return fmt.Sprintf("v%s", ver)
}

func normalizeTag(tagName common.ModuleTagName, ver string) string {
	if tagName == common.RepoRootTag {
		return ver
	}
	return fmt.Sprintf("%s/%s", tagName, ver)
}

// HasChanged checks if the files in the module set have changed since the last release.
func HasChanged(repoRoot string, versioningFile string, ver string, modset string) ([]string, error) {
	changedFiles := []string{}
	ver = normalizeVersion(ver)

	r, err := git.PlainOpen(repoRoot)
	if err != nil {
		return changedFiles, fmt.Errorf("could not open repo at %v: %w", repoRoot, err)
	}

	if e := common.VerifyWorkingTreeClean(r); e != nil {
		return changedFiles, fmt.Errorf("VerifyWorkingTreeClean failed: %w", e)
	}

	mset, err := common.NewModuleSetRelease(versioningFile, modset, repoRoot)
	if err != nil {
		return changedFiles, err
	}

	// get tag names of mods to update
	tagNames, err := common.ModulePathsToTagNames(
		mset.ModSet.Modules,
		mset.ModPathMap,
		repoRoot,
	)
	if err != nil {
		return changedFiles, fmt.Errorf("could not retrieve tag names from module paths: %w", err)
	}

	return filesChanged(r, modset, ver, tagNames, GitClient{})
}

func filesChanged(r *git.Repository, modset string, ver string, tagNames []common.ModuleTagName, client Client) ([]string, error) {
	changedFiles := []string{}
	headCommit, err := client.HeadCommit(r)
	if err != nil {
		return changedFiles, err
	}

	// get all modules in modset
	for _, tagName := range tagNames {
		tag := normalizeTag(tagName, ver)
		tagCommit, err := client.TagCommit(r, tag)
		if err != nil {
			if errors.Is(err, git.ErrTagNotFound) {
				log.Printf("Module %s does not have a %s tag", tagName, ver)
				log.Printf("%s release is required.", modset)
				return changedFiles, fmt.Errorf("tag not found %s", tag)
			}
			return changedFiles, err
		}

		files, err := client.FilesChanged(headCommit, tagCommit, string(tagName), ".go")
		if err != nil {
			return changedFiles, err
		}
		if len(files) > 0 {
			changedFiles = append(changedFiles, files...)
		}
	}

	return changedFiles, nil
}
