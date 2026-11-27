// Copyright The OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package tag

import (
	"path/filepath"
	"testing"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/stretchr/testify/require"

	"go.opentelemetry.io/build-tools/multimod/internal/shared"
	"go.opentelemetry.io/build-tools/multimod/internal/shared/sharedtest"
)

const runVersions = `module-sets:
  a:
    version: v1.0.0
    modules:
      - example.com/a
`

// setupRun creates a repo with a single module and a commit, makes it the
// working directory, and returns the versioning file path, the repo and the
// commit hash.
func setupRun(t *testing.T) (string, *git.Repository, string) {
	t.Helper()
	root := t.TempDir()
	repo, _, err := sharedtest.InitNewRepoWithCommit(root)
	require.NoError(t, err)

	require.NoError(t, sharedtest.WriteTempFiles(map[string][]byte{
		filepath.Join(root, "go.mod"):        []byte("module example.com/a\n\ngo 1.21\n"),
		filepath.Join(root, "versions.yaml"): []byte(runVersions),
	}))
	hash, err := shared.CommitChangesToNewBranch("branch", "commit", repo, sharedtest.TestAuthor)
	require.NoError(t, err)

	t.Chdir(root)
	return filepath.Join(root, "versions.yaml"), repo, hash.String()
}

func TestRun(t *testing.T) {
	t.Run("delete tags and print them", func(t *testing.T) {
		vFile, repo, hash := setupRun(t)
		headHash, err := repo.ResolveRevision(plumbing.Revision(hash))
		require.NoError(t, err)
		_, err = repo.CreateTag("v1.0.0", *headHash, &git.CreateTagOptions{Message: "m", Tagger: sharedtest.TestAuthor})
		require.NoError(t, err)

		require.NoError(t, Run(vFile, "a", hash, true, true))

		_, err = repo.Tag("v1.0.0")
		require.ErrorIs(t, err, git.ErrTagNotFound)
	})
}

func TestRunErrors(t *testing.T) {
	t.Run("not in a repository", func(t *testing.T) {
		t.Chdir(t.TempDir())
		err := Run("versions.yaml", "a", "HEAD", false, false)
		require.ErrorContains(t, err, "unable to change to repo root")
	})

	t.Run("invalid versioning file", func(t *testing.T) {
		vFile, _, hash := setupRun(t)
		err := Run(vFile+".missing", "a", hash, false, false)
		require.ErrorContains(t, err, "Error creating new tagger struct")
	})

	t.Run("delete tags that do not exist", func(t *testing.T) {
		vFile, _, hash := setupRun(t)
		err := Run(vFile, "a", hash, true, false)
		require.ErrorContains(t, err, "Error creating new tagger struct")
	})

	t.Run("tag already exists", func(t *testing.T) {
		vFile, repo, hash := setupRun(t)
		headHash, err := repo.ResolveRevision(plumbing.Revision(hash))
		require.NoError(t, err)
		_, err = repo.CreateTag("v1.0.0", *headHash, &git.CreateTagOptions{Message: "m", Tagger: sharedtest.TestAuthor})
		require.NoError(t, err)

		err = Run(vFile, "a", hash, false, false)
		require.ErrorContains(t, err, "Error creating new tagger struct")
	})
}
