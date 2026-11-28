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

package shared // nolint:revive // keeping generic package name until a proper refactoring is done

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
)

// CommitChangesToNewBranch creates a new branch, commits to it, and returns to the original worktree.
func CommitChangesToNewBranch(branchName string, commitMessage string, repo *git.Repository, customAuthor *object.Signature) (plumbing.Hash, error) {
	// save reference to current head in storage
	origRef, err := repo.Head()
	if err != nil {
		return plumbing.ZeroHash, fmt.Errorf("could not get repo head: %w", err)
	}

	if err = repo.Storer.SetReference(origRef); err != nil {
		return plumbing.ZeroHash, errors.New("could not store original head ref")
	}

	if _, err = checkoutNewBranch(branchName, repo); err != nil {
		return plumbing.ZeroHash, fmt.Errorf("createPrereleaseBranch failed: %w", err)
	}

	hash, err := CommitChanges(commitMessage, repo, customAuthor)
	if err != nil {
		return plumbing.ZeroHash, fmt.Errorf("could not commit changes: %w", err)
	}

	// return to original branch
	err = checkoutExistingBranch(origRef.Name(), repo)
	if err != nil {
		return hash, fmt.Errorf("could not checkout original branch: %w", err)
	}

	return hash, err
}

// CommitChanges commits changes to the current branch.
func CommitChanges(commitMessage string, repo *git.Repository, customAuthor *object.Signature) (plumbing.Hash, error) {
	// commit changes to git
	log.Printf("Committing changes to git with message '%v'\n", commitMessage)

	worktree, err := GetWorktree(repo)
	if err != nil {
		return plumbing.ZeroHash, err
	}

	var commitOptions *git.CommitOptions
	if customAuthor == nil {
		commitOptions = &git.CommitOptions{
			All:               true,
			AllowEmptyCommits: true,
		}
	} else {
		commitOptions = &git.CommitOptions{
			All:               true,
			Author:            customAuthor,
			AllowEmptyCommits: true,
		}
	}

	hash, err := worktree.Commit(commitMessage, commitOptions)
	if err != nil {
		// go-git rejects valid branch merge refs such as refs/pull/123/head.
		// Git itself accepts them, so use it when that config validation fails.
		if strings.Contains(err.Error(), "branch config: invalid merge") {
			// The commit message is passed as a separate argument, never through a shell.
			// #nosec G204
			cmd := exec.Command("git", "commit", "--all", "--allow-empty", "-m", commitMessage)
			cmd.Dir = worktree.Filesystem.Root()
			if customAuthor != nil {
				cmd.Env = append(os.Environ(),
					"GIT_AUTHOR_NAME="+customAuthor.Name,
					"GIT_AUTHOR_EMAIL="+customAuthor.Email,
					"GIT_AUTHOR_DATE="+customAuthor.When.Format("2006-01-02T15:04:05-0700"),
					"GIT_COMMITTER_NAME="+customAuthor.Name,
					"GIT_COMMITTER_EMAIL="+customAuthor.Email,
					"GIT_COMMITTER_DATE="+customAuthor.When.Format("2006-01-02T15:04:05-0700"),
				)
			}
			if output, gitErr := cmd.CombinedOutput(); gitErr != nil {
				return plumbing.ZeroHash, fmt.Errorf("could not commit changes to git: %s: %w", output, gitErr)
			}
			ref, headErr := repo.Head()
			if headErr != nil {
				return plumbing.ZeroHash, fmt.Errorf("could not get committed head: %w", headErr)
			}
			return ref.Hash(), nil
		}
		return plumbing.ZeroHash, fmt.Errorf("could not commit changes to git: %w", err)
	}

	return hash, nil
}

func checkoutExistingBranch(branchRefName plumbing.ReferenceName, repo *git.Repository) error {
	worktree, err := repo.Worktree()
	if err != nil {
		return &errGetWorktreeFailed{reason: err}
	}

	checkoutOptions := &git.CheckoutOptions{
		Branch: branchRefName,
		Create: false,
		Keep:   false,
	}

	log.Printf("git checkout %v\n", branchRefName)
	if err = worktree.Checkout(checkoutOptions); err != nil {
		return fmt.Errorf("could not check out new branch: %w", err)
	}

	return nil
}

func checkoutNewBranch(branchName string, repo *git.Repository) (plumbing.ReferenceName, error) {
	worktree, err := repo.Worktree()
	if err != nil {
		return "", &errGetWorktreeFailed{reason: err}
	}

	branchRefName := plumbing.NewBranchReferenceName(branchName)

	checkoutOptions := &git.CheckoutOptions{
		Branch: branchRefName,
		Create: true,
		Keep:   true,
	}

	log.Printf("git branch %v\n", branchName)
	if err = worktree.Checkout(checkoutOptions); err != nil {
		return "", fmt.Errorf("could not check out new branch: %w", err)
	}

	return branchRefName, nil
}

// GetWorktree returns the worktree of a repo.
func GetWorktree(repo *git.Repository) (*git.Worktree, error) {
	worktree, err := repo.Worktree()
	if err != nil {
		return nil, &errGetWorktreeFailed{reason: err}
	}

	return worktree, nil
}

// VerifyWorkingTreeClean returns nil if the working tree is clean or an error if not.
func VerifyWorkingTreeClean(repo *git.Repository) error {
	worktree, err := GetWorktree(repo)
	if err != nil {
		return err
	}

	status, err := worktree.Status()
	if err != nil {
		return fmt.Errorf("could not get worktree status: %w", err)
	}

	if !status.IsClean() {
		return &errWorkingTreeNotClean{}
	}

	return nil
}
