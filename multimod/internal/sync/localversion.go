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

package sync // nolint:revive // TODO: rename this internal package

import (
	"errors"
	"fmt"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"golang.org/x/mod/module"
	"golang.org/x/mod/semver"
)

// pseudoVersionHashLen is the number of hash characters in a Go pseudo-version.
const pseudoVersionHashLen = 12

// localPseudoVersion computes the Go pseudo-version of a revision of the
// repository checked out at repoRoot, without asking the Go module proxy.
//
// The proxy only knows about a commit after it has fetched it, and negative
// lookups are cached, so asking it about a brand-new commit can take a long
// time. The pseudo-version is fully determined by the commit time, the commit
// hash and the version it follows, so it can be computed from the local
// checkout instead.
//
// baseVersion is the version that the pseudo-version follows. It is assumed to
// be the most recent release reachable from rev, which is what the module set
// version in the repository's versioning file records. The result is
// vX.Y.(Z+1)-0.yyyymmddhhmmss-abcdefabcdef for a release baseVersion.
func localPseudoVersion(repoRoot, rev, baseVersion string) (string, error) {
	if rev == "" {
		return "", errors.New("a commit hash or branch name is required to compute a version locally")
	}
	if !semver.IsValid(baseVersion) {
		return "", fmt.Errorf("base version %q is not a valid semantic version", baseVersion)
	}

	repo, err := git.PlainOpenWithOptions(repoRoot, &git.PlainOpenOptions{DetectDotGit: true})
	if err != nil {
		return "", fmt.Errorf("could not open repo at %q: %w", repoRoot, err)
	}

	hash, err := repo.ResolveRevision(plumbing.Revision(rev))
	if err != nil {
		return "", fmt.Errorf("could not resolve revision %q in %q: %w", rev, repoRoot, err)
	}

	commit, err := resolveCommit(repo, *hash)
	if err != nil {
		return "", fmt.Errorf("revision %q in %q is not a commit: %w", rev, repoRoot, err)
	}

	// Go uses the commit (not author) time, in UTC.
	return module.PseudoVersion(
		semver.Major(baseVersion),
		baseVersion,
		commit.Committer.When.UTC(),
		commit.Hash.String()[:pseudoVersionHashLen],
	), nil
}

// resolveCommit returns the commit a hash refers to, following annotated tags.
func resolveCommit(repo *git.Repository, hash plumbing.Hash) (*object.Commit, error) {
	if commit, err := repo.CommitObject(hash); err == nil {
		return commit, nil
	}
	tag, err := repo.TagObject(hash)
	if err != nil {
		return nil, fmt.Errorf("object %s is neither a commit nor an annotated tag", hash)
	}
	return tag.Commit()
}
