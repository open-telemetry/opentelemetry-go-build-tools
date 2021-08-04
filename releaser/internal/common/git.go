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

package common

import (
	"fmt"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"log"
)

func CommitChangesToNewBranch(branchName string, commitMessage string, repo *git.Repository) error {
	// save reference to current head in storage
	origRef, err := repo.Head()
	if err != nil {
		return fmt.Errorf("could not get repo head: %v", err)
	}

	if err = repo.Storer.SetReference(origRef); err != nil {
		return fmt.Errorf("could not store original head ref")
	}

	if _, err = CheckoutNewGitBranch(branchName, repo); err != nil {
		return fmt.Errorf("createPrereleaseBranch failed: %v", err)
	}

	if err = commitChanges(commitMessage, repo); err != nil {
		return fmt.Errorf("could not commit changes: %v", err)
	}

	// return to original branch
	err = CheckoutExistingGitBranch(origRef.Name(), repo)
	if err != nil {
		log.Fatal("unable to checkout original branch")
	}

	return err
}

func commitChanges(commitMessage string, repo *git.Repository) error {
	// commit changes to git
	log.Printf("Committing changes to git with message '%v'\n", commitMessage)

	worktree, err := GetWorktree(repo)
	if err != nil {
		return err
	}

	commitOptions := &git.CommitOptions{
		All: true,
	}

	hash, err := worktree.Commit(commitMessage, commitOptions)
	if err != nil {
		return fmt.Errorf("could not commit changes to git: %v", err)
	}

	log.Printf("Commit successful. Hash of commit: %s\n", hash)

	return nil
}

func CheckoutExistingGitBranch(branchRefName plumbing.ReferenceName, repo *git.Repository) error {
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
		return fmt.Errorf("could not check out new branch: %v", err)
	}

	return nil
}

func CheckoutNewGitBranch(branchName string, repo *git.Repository) (plumbing.ReferenceName, error) {
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
		return "", fmt.Errorf("could not check out new branch: %v", err)
	}

	return branchRefName, nil
}

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
		return fmt.Errorf("could not get worktree status: %v", err)
	}

	if !status.IsClean() {
		return &errWorkingTreeNotClean{}
	}

	return nil
}
