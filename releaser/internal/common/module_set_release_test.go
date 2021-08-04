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
	"os"
	"path/filepath"
	"testing"

	"github.com/go-git/go-git/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.opentelemetry.io/build-tools/releaser/internal/common/commontest"
)

func TestNewModuleSetRelease(t *testing.T) {
	tmpRootDir, err := os.MkdirTemp(testDataDir, "NewModuleSetRelease")
	if err != nil {
		t.Fatal("error creating temp dir:", err)
	}

	defer commontest.RemoveAll(t, tmpRootDir)

	modFiles := map[string][]byte{
		filepath.Join(tmpRootDir, "test", "test1", "go.mod"): []byte("module \"go.opentelemetry.io/test/test1\"\n\ngo 1.16\n\n" +
			"require (\n\t\"go.opentelemetry.io/testroot/v2\" v2.0.0\n)\n"),
		filepath.Join(tmpRootDir, "test", "go.mod"):          []byte("module go.opentelemetry.io/test3\n\ngo 1.16\n"),
		filepath.Join(tmpRootDir, "go.mod"):                  []byte("module go.opentelemetry.io/testroot/v2\n\ngo 1.16\n"),
		filepath.Join(tmpRootDir, "test", "test2", "go.mod"): []byte("module \"go.opentelemetry.io/test/testexcluded\"\n\ngo 1.16\n"),
	}

	if err := commontest.WriteGoModFiles(modFiles); err != nil {
		t.Fatal("could not create go mod file tree", err)
	}

	// initialize temporary local git repository
	_, err = git.PlainInit(tmpRootDir, true)
	if err != nil {
		t.Fatal("could not initialize temp git repo:", err)
	}

	testCases := []struct {
		name                   string
		versioningFilename     string
		repoRoot               string
		shouldError            bool
		expectedModuleSetMap   ModuleSetMap
		expectedModulePathMap  ModulePathMap
		expectedModuleInfoMap  ModuleInfoMap
		expectedTagNames       map[string][]ModuleTagName
		expectedFullTagNames   map[string][]string
		expectedModSetVersions map[string]string
		expectedModSetPaths    map[string][]ModulePath
	}{
		{
			name:               "valid versioning",
			versioningFilename: filepath.Join(testDataDir, "new_module_set_release/versions_valid.yaml"),
			repoRoot:           tmpRootDir,
			shouldError:        false,
			expectedModuleSetMap: ModuleSetMap{
				"mod-set-1": ModuleSet{
					Version: "v1.2.3-RC1+meta",
					Modules: []ModulePath{
						"go.opentelemetry.io/test/test1",
					},
				},
				"mod-set-2": ModuleSet{
					Version: "v0.1.0",
					Modules: []ModulePath{
						"go.opentelemetry.io/test3",
					},
				},
				"mod-set-3": ModuleSet{
					Version: "v2.2.2",
					Modules: []ModulePath{
						"go.opentelemetry.io/testroot/v2",
					},
				},
			},
			expectedModulePathMap: ModulePathMap{
				"go.opentelemetry.io/test/test1":  ModuleFilePath(filepath.Join(tmpRootDir, "test", "test1", "go.mod")),
				"go.opentelemetry.io/test3":       ModuleFilePath(filepath.Join(tmpRootDir, "test", "go.mod")),
				"go.opentelemetry.io/testroot/v2": ModuleFilePath(filepath.Join(tmpRootDir, "go.mod")),
			},
			expectedModuleInfoMap: ModuleInfoMap{
				"go.opentelemetry.io/test/test1": ModuleInfo{
					ModuleSetName: "mod-set-1",
					Version:       "v1.2.3-RC1+meta",
				},
				"go.opentelemetry.io/testroot/v2": ModuleInfo{
					ModuleSetName: "mod-set-3",
					Version:       "v2.2.2",
				},
				"go.opentelemetry.io/test3": ModuleInfo{
					ModuleSetName: "mod-set-2",
					Version:       "v0.1.0",
				},
			},
			expectedTagNames: map[string][]ModuleTagName{
				"mod-set-1": []ModuleTagName{"test/test1"},
				"mod-set-2": []ModuleTagName{"test"},
				"mod-set-3": []ModuleTagName{repoRootTag},
			},
			expectedFullTagNames: map[string][]string{
				"mod-set-1": []string{"test/test1/v1.2.3-RC1+meta"},
				"mod-set-2": []string{"test/v0.1.0"},
				"mod-set-3": []string{"v2.2.2"},
			},
			expectedModSetVersions: map[string]string{
				"mod-set-1": "v1.2.3-RC1+meta",
				"mod-set-2": "v0.1.0",
				"mod-set-3": "v2.2.2",
			},
			expectedModSetPaths: map[string][]ModulePath{
				"mod-set-1": []ModulePath{"go.opentelemetry.io/test/test1"},
				"mod-set-2": []ModulePath{"go.opentelemetry.io/test3"},
				"mod-set-3": []ModulePath{"go.opentelemetry.io/testroot/v2"},
			},
		},
		{
			name:                  "invalid version file syntax",
			versioningFilename:    filepath.Join(testDataDir, "new_module_set_release/versions_invalid_syntax.yaml"),
			repoRoot:              tmpRootDir,
			shouldError:           true,
			expectedModuleSetMap:  nil,
			expectedModulePathMap: nil,
			expectedModuleInfoMap: nil,
			expectedTagNames:      nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			for expectedModSetName, expectedModSet := range tc.expectedModuleSetMap {
				actual, err := NewModuleSetRelease(tc.versioningFilename, expectedModSetName, tc.repoRoot)

				if tc.shouldError {
					assert.Error(t, err)
				} else {
					require.NoError(t, err)
				}

				assert.IsType(t, ModuleVersioning{}, actual.ModuleVersioning)
				assert.Equal(t, tc.expectedModuleSetMap, actual.ModuleVersioning.ModSetMap)
				assert.Equal(t, tc.expectedModulePathMap, actual.ModuleVersioning.ModPathMap)
				assert.Equal(t, tc.expectedModuleInfoMap, actual.ModuleVersioning.ModInfoMap)

				assert.IsType(t, ModuleSetRelease{}, actual)
				assert.Equal(t, tc.expectedTagNames[expectedModSetName], actual.TagNames)
				assert.Equal(t, expectedModSet, actual.ModSet)
				assert.Equal(t, expectedModSetName, actual.ModSetName)

				// property functions
				assert.Equal(t, tc.expectedFullTagNames[expectedModSetName], actual.ModuleFullTagNames())
				assert.Equal(t, tc.expectedModSetVersions[expectedModSetName], actual.ModSetVersion())
				assert.Equal(t, tc.expectedModSetPaths[expectedModSetName], actual.ModSetPaths())
			}
		})
	}
}

func TestVerifyGitTagsDoNotAlreadyExist(t *testing.T) {
	tmpRootDir, err := os.MkdirTemp(testDataDir, "CheckGitTagsAlreadyExist")
	if err != nil {
		t.Fatal("error creating temp dir:", err)
	}

	defer commontest.RemoveAll(t, tmpRootDir)

	modFiles := map[string][]byte{
		filepath.Join(tmpRootDir, "test", "test1", "go.mod"): []byte("module \"go.opentelemetry.io/test/test1\"\n\ngo 1.16\n\n" +
			"require (\n\t\"go.opentelemetry.io/testroot/v2\" v2.0.0\n)\n"),
		filepath.Join(tmpRootDir, "test", "test4", "go.mod"): []byte("module \"go.opentelemetry.io/test/test4\"\n\ngo 1.16\n\n" +
			"require (\n\t\"go.opentelemetry.io/testroot/v2\" v2.0.0\n)\n"),
		filepath.Join(tmpRootDir, "test", "go.mod"):          []byte("module go.opentelemetry.io/test3\n\ngo 1.16\n"),
		filepath.Join(tmpRootDir, "go.mod"):                  []byte("module go.opentelemetry.io/testroot/v2\n\ngo 1.16\n"),
		filepath.Join(tmpRootDir, "test", "test2", "go.mod"): []byte("module \"go.opentelemetry.io/test/testexcluded\"\n\ngo 1.16\n"),
	}

	if err := commontest.WriteGoModFiles(modFiles); err != nil {
		t.Fatal("could not create go mod file tree", err)
	}

	versioningFilename := filepath.Join(testDataDir, "verify_git_tags_do_not_already_exist/versions_valid.yaml")
	repoRoot := tmpRootDir

	// initialize temporary local git repository
	repo, err := git.PlainInit(tmpRootDir, false)
	if err != nil {
		t.Fatal("could not initialize temp git repo:", err)
	}

	worktree, err := repo.Worktree()
	if err != nil {
		t.Fatal("could not get repo worktree:", err)
	}

	commitHash, err := worktree.Commit("test commit", &git.CommitOptions{})
	if err != nil {
		t.Fatal("could not commit to worktree:", err)
	}

	// create new tags in repo
	for _, tag := range []string{
		"test/test1/v1.2.3-RC1+meta",
		"test/test4/v1.2.3-RC1+meta",
		"v2.2.2",
		"test/test4/v1.0.0-previousVersion",
		"test/test5/v1.0.0-modNotExist",
		"v1.0.0-previousVersion",
	} {
		_, err = repo.CreateTag(tag, commitHash, &git.CreateTagOptions{
			Message: tag,
		})
		if err != nil {
			t.Fatal("error creating tag:", err)
		}
	}

	testCases := []struct {
		name          string
		modSetName    string
		expectedError error
	}{
		{
			name:       "multiple git tags exist",
			modSetName: "mod-set-1",
			expectedError: &ErrGitTagsAlreadyExist{
				tagNames: []string{
					"test/test1/v1.2.3-RC1+meta",
					"test/test4/v1.2.3-RC1+meta",
				},
			},
		},
		{
			name:          "git tags do not exist",
			modSetName:    "mod-set-2",
			expectedError: nil,
		},
		{
			name:       "root git tag exists",
			modSetName: "mod-set-3",
			expectedError: &ErrGitTagsAlreadyExist{
				tagNames: []string{
					"v2.2.2",
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			modSetRelease, err := NewModuleSetRelease(versioningFilename, tc.modSetName, repoRoot)
			require.NoError(t, err)

			repo, err := git.PlainOpen(repoRoot)
			require.NoError(t, err)

			actual := modSetRelease.CheckGitTagsAlreadyExist(repo)
			assert.Equal(t, tc.expectedError, actual)
		})
	}
}
