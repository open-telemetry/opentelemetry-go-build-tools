// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package diff

import (
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/go-git/go-git/v5"

	"go.opentelemetry.io/build-tools/multimod/internal/common"
)

// check

// normalizeVersion ensures the version is prefixed with a `v`.
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

func HasChanged(repoRoot string, versioningFile string, ver string, modset string) (bool, []string, error) {
	changed := false
	changedFiles := []string{}
	ver = normalizeVersion(ver)

	r, err := git.PlainOpen(repoRoot)
	if err != nil {
		return changed, changedFiles, fmt.Errorf("could not open repo at %v: %w", repoRoot, err)
	}

	if e := common.VerifyWorkingTreeClean(r); e != nil {
		return changed, changedFiles, fmt.Errorf("VerifyWorkingTreeClean failed: %w", e)
	}

	mset, err := common.NewModuleSetRelease(versioningFile, modset, repoRoot)
	if err != nil {
		return changed, changedFiles, err
	}

	// get tag names of mods to update
	tagNames, err := common.ModulePathsToTagNames(
		mset.ModSet.Modules,
		mset.ModuleVersioning.ModPathMap,
		repoRoot,
	)
	if err != nil {
		return changed, changedFiles, fmt.Errorf("could not retrieve tag names from module paths: %w", err)
	}

	return filesChanged(r, modset, ver, tagNames, common.GitClient{})
}

func filesChanged(r *git.Repository, modset string, ver string, tagNames []common.ModuleTagName, client common.Client) (bool, []string, error) {
	changed := false
	changedFiles := []string{}
	headCommit, err := client.HeadCommit(r)
	if err != nil {
		return changed, changedFiles, err
	}

	// get all modules in modset
	for _, tagName := range tagNames {
		// check tag exists
		tag := normalizeTag(tagName, ver)

		tagCommit, err := client.TagCommit(r, tag)
		if err != nil {
			if errors.Is(err, git.ErrTagNotFound) {
				log.Printf("Module %s does not have a %s tag", tagName, ver)
				log.Printf("%s release is required.", modset)
				return changed, changedFiles, fmt.Errorf("tag not found %s", tag)
			}
			return changed, changedFiles, err
		}

		files, err := client.FilesChanged(headCommit, tagCommit, string(tagName), ".go")
		if err != nil {
			return changed, changedFiles, err
		}
		if len(files) > 0 {
			changed = true
			changedFiles = append(changedFiles, files...)
		}
	}

	return changed, changedFiles, nil
}
