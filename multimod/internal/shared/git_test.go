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
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"go.opentelemetry.io/build-tools/multimod/internal/shared/sharedtest"
)

func TestCommitChangesWithPullRequestMergeRef(t *testing.T) {
	root := t.TempDir()
	repository, _, err := sharedtest.InitNewRepoWithCommit(root)
	require.NoError(t, err)
	worktree, err := repository.Worktree()
	require.NoError(t, err)
	versionFile := filepath.Join(root, "version.go")
	require.NoError(t, os.WriteFile(versionFile, []byte("package test\n"), 0o600))
	_, err = worktree.Add("version.go")
	require.NoError(t, err)
	_, err = CommitChanges("add version file", repository, sharedtest.TestAuthor)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(versionFile, []byte("package test\n// updated\n"), 0o600))

	name := exec.Command("git", "config", "user.name", "test_author")
	name.Dir = root
	require.NoError(t, name.Run())
	email := exec.Command("git", "config", "user.email", "test@example.com")
	email.Dir = root
	require.NoError(t, email.Run())
	merge := exec.Command("git", "config", "branch.pr.merge", "refs/pull/123/head")
	merge.Dir = root
	require.NoError(t, merge.Run())

	original, err := repository.Head()
	require.NoError(t, err)
	hash, err := CommitChangesToNewBranch("prerelease_test", "update version", repository, nil)
	require.NoError(t, err)
	require.NotEmpty(t, hash.String())
	ref, err := repository.Head()
	require.NoError(t, err)
	require.Equal(t, original.Hash(), ref.Hash())
	branch, err := repository.Reference("refs/heads/prerelease_test", true)
	require.NoError(t, err)
	require.Equal(t, hash, branch.Hash())
	commit, err := repository.CommitObject(hash)
	require.NoError(t, err)
	file, err := commit.File("version.go")
	require.NoError(t, err)
	contents, err := file.Contents()
	require.NoError(t, err)
	require.Contains(t, contents, "// updated")
}
