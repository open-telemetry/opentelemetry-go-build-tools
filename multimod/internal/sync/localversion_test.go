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
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.opentelemetry.io/build-tools/multimod/internal/shared/sharedtest"
)

// commitAt commits an empty change with the given committer time.
func commitAt(t *testing.T, repo *git.Repository, when time.Time) string {
	t.Helper()
	wt, err := repo.Worktree()
	require.NoError(t, err)
	sig := &object.Signature{Name: "test_author", Email: "test_email", When: when}
	hash, err := wt.Commit("commit", &git.CommitOptions{
		Author:            sig,
		Committer:         sig,
		AllowEmptyCommits: true,
	})
	require.NoError(t, err)
	return hash.String()
}

func TestLocalPseudoVersion(t *testing.T) {
	repoRoot := t.TempDir()
	repo, _, err := sharedtest.InitNewRepoWithCommit(repoRoot)
	require.NoError(t, err)

	// A committer time in a non-UTC zone: the version must still use UTC.
	zone := time.FixedZone("test", 5*60*60)
	hash := commitAt(t, repo, time.Date(2026, 11, 25, 22, 35, 36, 0, zone))
	wantSuffix := "20261125173536-" + hash[:12]

	headTag, err := repo.CreateTag("v1.72.0-annotated", plumbingHash(t, repo, hash), &git.CreateTagOptions{
		Tagger:  &object.Signature{Name: "test_author", Email: "test_email", When: time.Now()},
		Message: "annotated",
	})
	require.NoError(t, err)
	require.NotNil(t, headTag)

	tests := []struct {
		name    string
		rev     string
		base    string
		want    string
		wantErr string
	}{
		{name: "full hash, release base", rev: hash, base: "v1.72.0", want: "v1.72.1-0." + wantSuffix},
		{name: "short hash", rev: hash[:8], base: "v1.72.0", want: "v1.72.1-0." + wantSuffix},
		{name: "branch name", rev: "master", base: "v0.166.0", want: "v0.166.1-0." + wantSuffix},
		{name: "annotated tag", rev: "v1.72.0-annotated", base: "v1.72.0", want: "v1.72.1-0." + wantSuffix},
		{name: "prerelease base", rev: hash, base: "v1.2.3-rc.1", want: "v1.2.3-rc.1.0." + wantSuffix},
		{name: "empty revision", rev: "", base: "v1.0.0", wantErr: "commit hash or branch name is required"},
		{name: "invalid base", rev: hash, base: "1.0", wantErr: "not a valid semantic version"},
		{name: "unknown revision", rev: "does-not-exist", base: "v1.0.0", wantErr: "could not resolve revision"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := localPseudoVersion(repoRoot, tc.rev, tc.base)
			if tc.wantErr != "" {
				require.ErrorContains(t, err, tc.wantErr)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.want, got)
		})
	}

	t.Run("not a repository", func(t *testing.T) {
		_, err := localPseudoVersion(t.TempDir(), "main", "v1.0.0")
		require.ErrorContains(t, err, "could not open repo")
	})
}

func TestUpdateAllGoModFilesWithLocalVersion(t *testing.T) {
	versionsYamlDir := filepath.Join(testDataDir, "update_all_go_mod_files_with_commit_hash")
	myVersioningFilename := filepath.Join(versionsYamlDir, "versions_valid.yaml")
	otherVersioningFilename := filepath.Join(versionsYamlDir, "other_versions_valid.yaml")

	tmpRootDir := t.TempDir()
	modFile := filepath.Join(tmpRootDir, "my", "test", "test1", "go.mod")
	require.NoError(t, sharedtest.WriteTempFiles(map[string][]byte{
		modFile: []byte("module go.opentelemetry.io/build-tools/multimod/internal/sync/test/test1\n\n" +
			"go 1.16\n\n" +
			"require (\n\t" +
			"go.opentelemetry.io/other/test/test1 v1.0.0-old\n" +
			")"),
	}))

	s, err := newSync(myVersioningFilename, otherVersioningFilename, "other-mod-set-1", tmpRootDir, "main")
	require.NoError(t, err)
	s.localVersion = "v1.2.4-0.20261125173536-abcdefabcdef"
	// Any request to the proxy fails the test: the version must not be looked up.
	s.client = &http.Client{Transport: roundTripFunc(func(req *http.Request) *http.Response {
		t.Errorf("unexpected proxy request to %s", req.URL)
		return nil
	})}

	require.NoError(t, s.updateAllGoModFiles())

	actual, err := os.ReadFile(filepath.Clean(modFile))
	require.NoError(t, err)
	assert.Contains(t, string(actual), "go.opentelemetry.io/other/test/test1 v1.2.4-0.20261125173536-abcdefabcdef")
	assert.NotContains(t, string(actual), "v1.0.0-old")
}

func plumbingHash(t *testing.T, repo *git.Repository, hash string) plumbing.Hash {
	t.Helper()
	h, err := repo.ResolveRevision(plumbing.Revision(hash))
	require.NoError(t, err)
	return *h
}
