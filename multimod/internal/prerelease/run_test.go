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

package prerelease

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/go-git/go-git/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.opentelemetry.io/build-tools/multimod/internal/shared"
	"go.opentelemetry.io/build-tools/multimod/internal/shared/sharedtest"
)

const runVersions = `module-sets:
  a:
    version: v1.2.3
    modules:
      - example.com/a
`

// setupRun creates a repo with one module, stubs the package-level hooks so
// they use it, and returns the versioning file, repo root and repo.
func setupRun(t *testing.T) (string, string, *git.Repository) {
	t.Helper()
	root := t.TempDir()
	repo, _, err := sharedtest.InitNewRepoWithCommit(root)
	require.NoError(t, err)
	require.NoError(t, sharedtest.WriteTempFiles(map[string][]byte{
		filepath.Join(root, "go.mod"): []byte("module example.com/a\n\ngo 1.21\n"),
	}))
	vFile := filepath.Join(root, "versions.yaml")
	require.NoError(t, os.WriteFile(vFile, []byte(runVersions), 0o600))

	origFR, origWTC, origCommit := findRoot, workingTreeClean, commit
	t.Cleanup(func() { findRoot, workingTreeClean, commit = origFR, origWTC, origCommit })
	findRoot = func() (string, error) { return root, nil }
	workingTreeClean = func(*git.Repository) error { return nil }
	commit = func(shared.ModuleSetRelease, bool, *git.Repository) error { return nil }
	return vFile, root, repo
}

func TestRun(t *testing.T) {
	t.Run("explicit module set", func(t *testing.T) {
		vFile, _, _ := setupRun(t)
		var committed []string
		commit = func(r shared.ModuleSetRelease, _ bool, _ *git.Repository) error {
			committed = append(committed, r.ModSetName)
			return nil
		}
		require.NoError(t, Run(vFile, []string{"a"}, true, false))
		assert.Equal(t, []string{"a"}, committed)
	})

	t.Run("up to date module set is skipped", func(t *testing.T) {
		vFile, _, repo := setupRun(t)
		head, err := repo.ResolveRevision("HEAD")
		require.NoError(t, err)
		_, err = repo.CreateTag("v1.2.3", *head, &git.CreateTagOptions{Message: "m", Tagger: sharedtest.TestAuthor})
		require.NoError(t, err)
		commit = func(shared.ModuleSetRelease, bool, *git.Repository) error {
			t.Error("commit must not be called for an up to date module set")
			return nil
		}
		require.NoError(t, Run(vFile, nil, true, false))
	})

	t.Run("runs go mod tidy", func(t *testing.T) {
		vFile, _, _ := setupRun(t)
		require.NoError(t, Run(vFile, nil, false, false))
	})
}

func TestRunErrors(t *testing.T) {
	t.Run("cannot find repo root", func(t *testing.T) {
		vFile, _, _ := setupRun(t)
		findRoot = func() (string, error) { return "", errors.New("boom") }
		require.ErrorContains(t, Run(vFile, nil, true, false), "unable to find repo root")
	})

	t.Run("cannot list module sets", func(t *testing.T) {
		_, root, _ := setupRun(t)
		err := Run(filepath.Join(root, "missing.yaml"), nil, true, false)
		require.ErrorContains(t, err, "could not automatically get all module set names")
	})

	t.Run("not a git repository", func(t *testing.T) {
		vFile, _, _ := setupRun(t)
		findRoot = func() (string, error) { return t.TempDir(), nil }
		require.ErrorContains(t, Run(vFile, []string{"a"}, true, false), "could not open repo")
	})

	t.Run("dirty working tree", func(t *testing.T) {
		vFile, _, _ := setupRun(t)
		workingTreeClean = func(*git.Repository) error { return errors.New("dirty") }
		require.ErrorContains(t, Run(vFile, []string{"a"}, true, false), "VerifyWorkingTreeClean failed")
	})

	t.Run("invalid versioning file", func(t *testing.T) {
		_, root, _ := setupRun(t)
		err := Run(filepath.Join(root, "missing.yaml"), []string{"a"}, true, false)
		require.ErrorContains(t, err, "prerelease struct")
	})

	t.Run("commit fails", func(t *testing.T) {
		vFile, _, _ := setupRun(t)
		commit = func(shared.ModuleSetRelease, bool, *git.Repository) error { return errors.New("nope") }
		require.ErrorContains(t, Run(vFile, []string{"a"}, true, false), "commitChangesToNewBranch failed")
	})
}
