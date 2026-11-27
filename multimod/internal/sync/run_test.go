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
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.opentelemetry.io/build-tools/multimod/internal/shared/sharedtest"
)

const (
	runMyVersions = `module-sets:
  my-set:
    version: v1.0.0
    modules:
      - example.com/my/mod
`
	runOtherVersions = `module-sets:
  other-set:
    version: v1.2.3
    modules:
      - example.com/other/mod
  other-set-2:
    version: v0.4.0
    modules:
      - example.com/other/mod2
`
	runMyGoMod = "module example.com/my/mod\n\ngo 1.21\n\nrequire (\n" +
		"\texample.com/other/mod v1.0.0-old\n" +
		"\texample.com/other/mod2 v0.0.1-old\n)\n"
)

// commitAll stages and commits every file in the repository.
func commitAll(t *testing.T, repo *git.Repository, when time.Time) {
	t.Helper()
	wt, err := repo.Worktree()
	require.NoError(t, err)
	require.NoError(t, wt.AddWithOptions(&git.AddOptions{All: true}))
	sig := &object.Signature{Name: "test_author", Email: "test_email", When: when}
	_, err = wt.Commit("commit", &git.CommitOptions{Author: sig, Committer: sig})
	require.NoError(t, err)
}

type runEnv struct {
	myGoMod, myVersions, otherVersions, otherRoot string
}

// setupRun creates a "my" repo (made the working directory) and an "other"
// repo, both with clean working trees.
func setupRun(t *testing.T) runEnv {
	t.Helper()
	myRoot := t.TempDir()
	otherRoot := t.TempDir()

	env := runEnv{
		myGoMod:       filepath.Join(myRoot, "go.mod"),
		myVersions:    filepath.Join(myRoot, "versions.yaml"),
		otherVersions: filepath.Join(otherRoot, "versions.yaml"),
		otherRoot:     otherRoot,
	}
	require.NoError(t, sharedtest.WriteTempFiles(map[string][]byte{
		env.myGoMod:       []byte(runMyGoMod),
		env.myVersions:    []byte(runMyVersions),
		env.otherVersions: []byte(runOtherVersions),
	}))

	myRepo, _, err := sharedtest.InitNewRepoWithCommit(myRoot)
	require.NoError(t, err)
	commitAll(t, myRepo, time.Now())
	otherRepo, _, err := sharedtest.InitNewRepoWithCommit(otherRoot)
	require.NoError(t, err)
	commitAll(t, otherRepo, time.Date(2026, 11, 25, 22, 35, 36, 0, time.UTC))

	t.Chdir(myRoot)
	return env
}

func TestRun(t *testing.T) {
	t.Run("computes versions locally", func(t *testing.T) {
		env := setupRun(t)
		err := Run(env.myVersions, env.otherVersions, env.otherRoot,
			[]string{"other-set"}, "HEAD", false, true, true)
		require.NoError(t, err)

		got, err := os.ReadFile(env.myGoMod)
		require.NoError(t, err)
		assert.Contains(t, string(got), "example.com/other/mod v1.2.4-0.20261125223536-")
		assert.Contains(t, string(got), "example.com/other/mod2 v0.0.1-old")
	})

	t.Run("all module sets", func(t *testing.T) {
		env := setupRun(t)
		err := Run(env.myVersions, env.otherVersions, env.otherRoot,
			nil, "HEAD", true, true, true)
		require.NoError(t, err)

		got, err := os.ReadFile(env.myGoMod)
		require.NoError(t, err)
		assert.Contains(t, string(got), "example.com/other/mod v1.2.4-0.")
		assert.Contains(t, string(got), "example.com/other/mod2 v0.4.1-0.")
	})
}

func TestRunErrors(t *testing.T) {
	t.Run("local versions without commit", func(t *testing.T) {
		env := setupRun(t)
		err := Run(env.myVersions, env.otherVersions, env.otherRoot,
			[]string{"other-set"}, "", false, true, true)
		require.ErrorContains(t, err, "requires a commit hash or branch name")
	})

	t.Run("not in a repository", func(t *testing.T) {
		env := setupRun(t)
		t.Chdir(t.TempDir())
		err := Run(env.myVersions, env.otherVersions, env.otherRoot,
			[]string{"other-set"}, "HEAD", false, true, true)
		require.ErrorContains(t, err, "unable to find repo root")
	})

	t.Run("cannot list all module sets", func(t *testing.T) {
		env := setupRun(t)
		err := Run(env.myVersions, filepath.Join(env.otherRoot, "missing.yaml"), env.otherRoot,
			nil, "HEAD", true, true, true)
		require.ErrorContains(t, err, "could not automatically get all module set names")
	})

	t.Run("dirty working tree", func(t *testing.T) {
		env := setupRun(t)
		require.NoError(t, os.WriteFile(env.myGoMod, []byte(runMyGoMod+"// dirty\n"), 0o600))
		err := Run(env.myVersions, env.otherVersions, env.otherRoot,
			[]string{"other-set"}, "HEAD", false, true, true)
		require.ErrorContains(t, err, "VerifyWorkingTreeClean failed")
	})

	t.Run("invalid versioning file", func(t *testing.T) {
		env := setupRun(t)
		err := Run(filepath.Join(env.otherRoot, "missing.yaml"), env.otherVersions, env.otherRoot,
			[]string{"other-set"}, "HEAD", false, true, true)
		require.ErrorContains(t, err, "error creating new sync struct")
	})

	t.Run("unresolvable revision", func(t *testing.T) {
		env := setupRun(t)
		err := Run(env.myVersions, env.otherVersions, env.otherRoot,
			[]string{"other-set"}, "does-not-exist", false, true, true)
		require.ErrorContains(t, err, "could not compute version locally")
	})
}
