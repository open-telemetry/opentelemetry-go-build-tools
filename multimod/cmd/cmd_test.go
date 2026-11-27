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

package cmd

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.opentelemetry.io/build-tools/multimod/internal/shared/sharedtest"
)

func TestSubcommandsRegistered(t *testing.T) {
	want := map[string]*cobra.Command{
		"diff":       diffCmd,
		"prerelease": prereleaseCmd,
		"sync":       syncCmd,
		"tag":        tagCmd,
		"verify":     verifyCmd,
	}
	for name, c := range want {
		t.Run(name, func(t *testing.T) {
			found, _, err := rootCmd.Find([]string{name})
			require.NoError(t, err)
			assert.Same(t, c, found)
			assert.NotEmpty(t, c.Short)
		})
	}
}

func TestRootVersioningFileFlag(t *testing.T) {
	f := rootCmd.PersistentFlags().Lookup("versioning-file")
	require.NotNil(t, f)
	assert.Equal(t, "v", f.Shorthand)
	assert.Contains(t, f.DefValue, "versions.yaml")
}

func TestCommandFlags(t *testing.T) {
	tests := []struct {
		cmd       *cobra.Command
		flag      string
		shorthand string
		required  bool
	}{
		{diffCmd, "module-set-name", "m", true},
		{diffCmd, "previous-version", "p", true},
		{prereleaseCmd, "all-module-sets", "a", false},
		{prereleaseCmd, "module-set-names", "m", false},
		{prereleaseCmd, "commit-to-different-branch", "b", false},
		{syncCmd, "other-repo-root", "o", true},
		{syncCmd, "module-set-names", "m", true},
		{syncCmd, "other-versioning-file", "", false},
		{syncCmd, "commit-hash", "", false},
		{syncCmd, "compute-versions-locally", "", false},
		{syncCmd, "all-module-sets", "a", false},
		{syncCmd, "skip-go-mod-tidy", "s", false},
		{tagCmd, "commit-hash", "c", true},
		{tagCmd, "module-set-name", "m", true},
		{tagCmd, "delete-module-set-tags", "d", false},
		{tagCmd, "print-tags", "p", false},
	}
	for _, tc := range tests {
		t.Run(tc.cmd.Use+"/"+tc.flag, func(t *testing.T) {
			f := tc.cmd.Flags().Lookup(tc.flag)
			require.NotNil(t, f)
			assert.Equal(t, tc.shorthand, f.Shorthand)
			_, isRequired := f.Annotations[cobra.BashCompOneRequiredFlag]
			if tc.required {
				assert.Equal(t, []string{"true"}, f.Annotations[cobra.BashCompOneRequiredFlag])
			} else if isRequired {
				assert.Equal(t, []string{"false"}, f.Annotations[cobra.BashCompOneRequiredFlag])
			}
		})
	}
}

func TestSyncPreRunAllModuleSets(t *testing.T) {
	old := allModuleSetsSync
	t.Cleanup(func() {
		allModuleSetsSync = old
		syncCmd.Flags().Lookup("module-set-names").Annotations[cobra.BashCompOneRequiredFlag] = []string{"true"}
	})

	allModuleSetsSync = true
	syncCmd.PreRun(syncCmd, nil)
	f := syncCmd.Flags().Lookup("module-set-names")
	assert.Equal(t, []string{"false"}, f.Annotations[cobra.BashCompOneRequiredFlag])
}

func TestPrereleasePreRunAllModuleSets(t *testing.T) {
	require.NotNil(t, prereleaseCmd.PreRun)
	old := allModuleSets
	t.Cleanup(func() { allModuleSets = old })
	allModuleSets = true
	assert.NotPanics(t, func() { prereleaseCmd.PreRun(prereleaseCmd, nil) })
}

const (
	runMyVersions = `module-sets:
  my-set:
    version: v1.0.0
    modules:
      - example.com/a
      - example.com/a/b
`
	runOtherVersions = `module-sets:
  other-set:
    version: v1.2.3
    modules:
      - example.com/other
`
	runRootGoMod = "module example.com/a\n\ngo 1.21\n"
	runSubGoMod  = "module example.com/a/b\n\ngo 1.21\n\nrequire (\n" +
		"\texample.com/a v0.0.1-old\n\texample.com/other v0.0.1-old\n)\n"
)

// newRepo creates a git repo containing the given files, all committed.
func newRepo(t *testing.T, files map[string]string) (string, *git.Repository) {
	t.Helper()
	root := t.TempDir()
	repo, _, err := sharedtest.InitNewRepoWithCommit(root)
	require.NoError(t, err)

	contents := make(map[string][]byte, len(files))
	for p, c := range files {
		contents[filepath.Join(root, p)] = []byte(c)
	}
	require.NoError(t, sharedtest.WriteTempFiles(contents))

	wt, err := repo.Worktree()
	require.NoError(t, err)
	require.NoError(t, wt.AddWithOptions(&git.AddOptions{All: true}))
	_, err = wt.Commit("files", &git.CommitOptions{Author: sharedtest.TestAuthor, Committer: sharedtest.TestAuthor})
	require.NoError(t, err)
	return root, repo
}

// setupMyRepo creates the repo the commands operate on, makes it the working
// directory and points the versioning file at it.
func setupMyRepo(t *testing.T) (string, *git.Repository) {
	t.Helper()
	// go-git reads the committer identity from the global git config.
	home := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(home, ".gitconfig"),
		[]byte("[user]\n\tname = test\n\temail = test@example.com\n"), 0o600))
	t.Setenv("HOME", home)

	root, repo := newRepo(t, map[string]string{
		"go.mod":        runRootGoMod,
		"b/go.mod":      runSubGoMod,
		"versions.yaml": runMyVersions,
	})
	t.Chdir(root)

	old := versioningFile
	versioningFile = filepath.Join(root, "versions.yaml")
	t.Cleanup(func() { versioningFile = old })
	return root, repo
}

func TestVerifyRun(t *testing.T) {
	setupMyRepo(t)
	assert.NotPanics(t, func() { verifyCmd.Run(verifyCmd, nil) })
}

func TestTagRun(t *testing.T) {
	_, repo := setupMyRepo(t)
	head, err := repo.Head()
	require.NoError(t, err)
	for _, name := range []string{"v1.0.0", "b/v1.0.0"} {
		_, err = repo.CreateTag(name, head.Hash(), &git.CreateTagOptions{Message: "m", Tagger: sharedtest.TestAuthor})
		require.NoError(t, err)
	}

	oldName, oldHash, oldDelete, oldPrint := moduleSetName, commitHash, deleteModuleSetTags, printTags
	t.Cleanup(func() {
		moduleSetName, commitHash, deleteModuleSetTags, printTags = oldName, oldHash, oldDelete, oldPrint
	})
	moduleSetName, commitHash, deleteModuleSetTags, printTags = "my-set", head.Hash().String(), true, true

	tagCmd.Run(tagCmd, nil)

	_, err = repo.Tag("v1.0.0")
	require.ErrorIs(t, err, git.ErrTagNotFound)
}

func TestPrereleaseRun(t *testing.T) {
	_, repo := setupMyRepo(t)

	oldNames, oldSkip, oldBranch := moduleSetNames, skipGoModTidy, commitToDifferentBranch
	t.Cleanup(func() { moduleSetNames, skipGoModTidy, commitToDifferentBranch = oldNames, oldSkip, oldBranch })
	moduleSetNames, skipGoModTidy, commitToDifferentBranch = []string{"my-set"}, true, true

	prereleaseCmd.Run(prereleaseCmd, nil)

	// The release commit lands on a new branch; the current one is untouched.
	branches, err := repo.Branches()
	require.NoError(t, err)
	n := 0
	require.NoError(t, branches.ForEach(func(*plumbing.Reference) error { n++; return nil }))
	assert.Equal(t, 2, n)
}

func TestSyncRun(t *testing.T) {
	root, _ := setupMyRepo(t)
	otherRoot, _ := newRepo(t, map[string]string{"versions.yaml": runOtherVersions})

	oldRoot, oldFile, oldCommit, oldNames, oldSkip, oldLocal := otherRepoRoot, otherVersioningFile, otherVersionCommit, moduleSetNamesSync, skipGoModTidySync, computeLocalSync
	t.Cleanup(func() {
		otherRepoRoot, otherVersioningFile, otherVersionCommit = oldRoot, oldFile, oldCommit
		moduleSetNamesSync, skipGoModTidySync, computeLocalSync = oldNames, oldSkip, oldLocal
	})
	// otherVersioningFile is left empty to exercise the default.
	otherRepoRoot, otherVersioningFile, otherVersionCommit = otherRoot, "", "HEAD"
	moduleSetNamesSync, skipGoModTidySync, computeLocalSync = []string{"other-set"}, true, true

	syncCmd.Run(syncCmd, nil)

	got, err := os.ReadFile(filepath.Join(root, "b", "go.mod")) // #nosec G304
	require.NoError(t, err)
	assert.Contains(t, string(got), "example.com/other v1.2.4-0.")
	assert.Equal(t, filepath.Join(otherRoot, "versions.yaml"), otherVersioningFile)
}

// TestRunFailures checks that every command exits non-zero with a message
// when its Run fails. log.Fatalf ends the process, so each case re-executes
// the test binary.
func TestRunFailures(t *testing.T) {
	if name := os.Getenv("MULTIMOD_CMD_FATAL"); name != "" {
		// Not inside a git repository, so every command fails immediately.
		t.Chdir(t.TempDir())
		versioningFile = "versions.yaml"
		otherVersionCommit, computeLocalSync = "HEAD", true
		switch name {
		case "verify":
			verifyCmd.Run(verifyCmd, nil)
		case "tag":
			tagCmd.Run(tagCmd, nil)
		case "prerelease":
			prereleaseCmd.Run(prereleaseCmd, nil)
		case "sync":
			syncCmd.Run(syncCmd, nil)
		}
		return
	}

	for _, name := range []string{"verify", "tag", "prerelease", "sync"} {
		t.Run(name, func(t *testing.T) {
			c := exec.Command(os.Args[0], "-test.run=^TestRunFailures$") // #nosec G702 G204
			c.Env = append(os.Environ(), "MULTIMOD_CMD_FATAL="+name)
			out, err := c.CombinedOutput()

			var exitErr *exec.ExitError
			require.ErrorAs(t, err, &exitErr, "expected the command to exit non-zero:\n%s", out)
			assert.NotEqual(t, 0, exitErr.ExitCode())
			assert.Contains(t, string(out), name+" failed")
		})
	}
}
