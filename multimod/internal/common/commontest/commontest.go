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

package commontest

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
)

var (
	TestAuthor = &object.Signature{
		Name:  "test_author",
		Email: "test_email",
		When:  time.Now(),
	}
)

// WriteTempFiles is a helper function to dynamically write files such as go.mod or version.go used for testing.
func WriteTempFiles(modFiles map[string][]byte) error {
	perm := os.FileMode(0700)

	for modFilePath, file := range modFiles {
		path := filepath.Dir(modFilePath)
		if err := os.MkdirAll(path, perm); err != nil {
			return fmt.Errorf("error calling os.MkdirAll(%v, %v): %w", path, perm, err)
		}

		if err := os.WriteFile(modFilePath, file, perm); err != nil {
			return fmt.Errorf("could not write temporary file %w", err)
		}
	}

	return nil
}

func InitNewRepoWithCommit(repoRoot string) (*git.Repository, plumbing.Hash, error) {
	// initialize temporary local git repository
	repo, err := git.PlainInit(repoRoot, false)
	if err != nil {
		return nil, plumbing.ZeroHash, fmt.Errorf("could not initialize temp git repo: %w", err)
	}

	worktree, err := repo.Worktree()
	if err != nil {
		return nil, plumbing.ZeroHash, err
	}
	commitMessage := "test commit"

	commitHash, err := worktree.Commit(commitMessage, &git.CommitOptions{
		All:               true,
		Author:            TestAuthor,
		AllowEmptyCommits: true,
	})
	if err != nil {
		return nil, plumbing.ZeroHash, fmt.Errorf("could not commit changes to git: %w", err)
	}

	return repo, commitHash, nil
}
