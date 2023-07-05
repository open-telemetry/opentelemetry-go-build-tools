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

	headRef, err := r.Head()
	if err != nil {
		return changed, changedFiles, err
	}
	headCommit, err := r.CommitObject(headRef.Hash())
	if err != nil {
		return changed, changedFiles, err
	}

	// get all modules in modset
	for _, tagName := range tagNames {
		// check tag exists
		var tag string

		if tagName == common.RepoRootTag {
			tag = ver
		} else {
			tag = fmt.Sprintf("%s/%s", tagName, ver)
		}

		tagRef, err := r.Tag(tag)
		if err != nil {
			if errors.Is(err, git.ErrTagNotFound) {
				log.Printf("Module %s does not have a %s tag", tagName, ver)
				log.Printf("%s release is required.", modset)
				return changed, changedFiles, fmt.Errorf("tag not found %s", tag)
			}
			return changed, changedFiles, err
		}

		o, err := r.TagObject(tagRef.Hash())
		if err != nil {
			return changed, changedFiles, fmt.Errorf("tag object error %s %w", tagRef.Hash().String(), err)
		}
		// diff files since tag
		// tagRef.Hash()
		commit, err := r.CommitObject(o.Target)
		if err != nil {
			return changed, changedFiles, fmt.Errorf("tag commit object error %s %w", o.Target.String(), err)
		}

		p, err := headCommit.Patch(commit)
		if err != nil {
			return changed, changedFiles, fmt.Errorf("patch error %s", tag)
		}

		for _, f := range p.FilePatches() {
			from, to := f.Files()
			if from != nil && strings.HasSuffix(from.Path(), ".go") && strings.HasPrefix(from.Path(), string(tagName)) {
				changed = true
				changedFiles = append(changedFiles, from.Path())
				continue
			}
			if to != nil && strings.HasSuffix(to.Path(), ".go") && strings.HasPrefix(to.Path(), string(tagName)) {
				changed = true
				changedFiles = append(changedFiles, to.Path())
			}
		}
	}

	return changed, changedFiles, nil
}
