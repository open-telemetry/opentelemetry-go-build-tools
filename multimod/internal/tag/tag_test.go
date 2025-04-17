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

package tag

import (
	"io"
	"log"
	"os"
	"path/filepath"
	"testing"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.opentelemetry.io/build-tools/multimod/internal/common"
	"go.opentelemetry.io/build-tools/multimod/internal/common/commontest"
)

var (
	testDataDir, _ = filepath.Abs("./test_data")
)

// TestMain performs setup for the tests and suppress printing logs.
func TestMain(m *testing.M) {
	log.SetOutput(io.Discard)
	os.Exit(m.Run())
}

func TestNewTagger(t *testing.T) {
	testName := "new_tagger"
	versionsYamlDir := filepath.Join(testDataDir, testName)

	tmpRootDir := t.TempDir()
	repo, _, err := commontest.InitNewRepoWithCommit(tmpRootDir)
	require.NoError(t, err)

	fullHash, err := common.CommitChangesToNewBranch("test_commit", "commit used in a test", repo, commontest.TestAuthor)
	require.NoError(t, err)
	hashPrefix := fullHash.String()[:8]

	modFiles := map[string][]byte{
		filepath.Join(tmpRootDir, "test", "test1", "go.mod"): []byte("module \"go.opentelemetry.io/test/test1\"\n\ngo 1.16\n\n" +
			"require (\n\t\"go.opentelemetry.io/testroot/v2\" v2.0.0\n)\n"),
		filepath.Join(tmpRootDir, "test", "go.mod"):                 []byte("module go.opentelemetry.io/test2\n\ngo 1.16\n"),
		filepath.Join(tmpRootDir, "go.mod"):                         []byte("module go.opentelemetry.io/testroot/v2\n\ngo 1.16\n"),
		filepath.Join(tmpRootDir, "test", "testexcluded", "go.mod"): []byte("module \"go.opentelemetry.io/test/testexcluded\"\n\ngo 1.16\n"),
	}

	require.NoError(t, commontest.WriteTempFiles(modFiles), "could not create go mod file tree")

	versioningFilename := filepath.Join(versionsYamlDir, "versions_valid.yaml")
	repoRoot := tmpRootDir
	expectedModuleSetMap := common.ModuleSetMap{
		"mod-set-1": common.ModuleSet{
			Version: "v1.2.3-RC1+meta",
			Modules: []common.ModulePath{
				"go.opentelemetry.io/test/test1",
			},
		},
		"mod-set-2": common.ModuleSet{
			Version: "v0.1.0",
			Modules: []common.ModulePath{
				"go.opentelemetry.io/test2",
			},
		},
		"mod-set-3": common.ModuleSet{
			Version: "v2.2.2",
			Modules: []common.ModulePath{
				"go.opentelemetry.io/testroot/v2",
			},
		},
	}
	expectedModulePathMap := common.ModulePathMap{
		"go.opentelemetry.io/test/test1":  common.ModuleFilePath(filepath.Join(tmpRootDir, "test", "test1", "go.mod")),
		"go.opentelemetry.io/test2":       common.ModuleFilePath(filepath.Join(tmpRootDir, "test", "go.mod")),
		"go.opentelemetry.io/testroot/v2": common.ModuleFilePath(filepath.Join(tmpRootDir, "go.mod")),
	}
	expectedModuleInfoMap := common.ModuleInfoMap{
		"go.opentelemetry.io/test/test1": common.ModuleInfo{
			ModuleSetName: "mod-set-1",
			Version:       "v1.2.3-RC1+meta",
		},
		"go.opentelemetry.io/testroot/v2": common.ModuleInfo{
			ModuleSetName: "mod-set-3",
			Version:       "v2.2.2",
		},
		"go.opentelemetry.io/test2": common.ModuleInfo{
			ModuleSetName: "mod-set-2",
			Version:       "v0.1.0",
		},
	}
	expectedTagNames := map[string][]common.ModuleTagName{
		"mod-set-1": {"test/test1"},
		"mod-set-2": {"test"},
		"mod-set-3": {common.RepoRootTag},
	}
	expectedFullTagNames := map[string][]string{
		"mod-set-1": {"test/test1/v1.2.3-RC1+meta"},
		"mod-set-2": {"test/v0.1.0"},
		"mod-set-3": {"v2.2.2"},
	}
	expectedModSetVersions := map[string]string{
		"mod-set-1": "v1.2.3-RC1+meta",
		"mod-set-2": "v0.1.0",
		"mod-set-3": "v2.2.2",
	}
	expectedModSetPaths := map[string][]common.ModulePath{
		"mod-set-1": {"go.opentelemetry.io/test/test1"},
		"mod-set-2": {"go.opentelemetry.io/test2"},
		"mod-set-3": {"go.opentelemetry.io/testroot/v2"},
	}

	for expectedModSetName, expectedModSet := range expectedModuleSetMap {
		actual, err := newTagger(versioningFilename, expectedModSetName, repoRoot, hashPrefix, false)
		require.NoError(t, err)

		assert.IsType(t, tagger{}, actual)

		assert.Equal(t, fullHash, actual.CommitHash)
		assert.IsType(t, &git.Repository{}, actual.Repo)

		assert.IsType(t, common.ModuleSetRelease{}, actual.ModuleSetRelease)
		assert.Equal(t, expectedTagNames[expectedModSetName], actual.TagNames)
		assert.Equal(t, expectedModSet, actual.ModSet)
		assert.Equal(t, expectedModSetName, actual.ModSetName)

		assert.IsType(t, common.ModuleVersioning{}, actual.ModuleVersioning)
		assert.Equal(t, expectedModuleSetMap, actual.ModSetMap)
		assert.Equal(t, expectedModulePathMap, actual.ModPathMap)
		assert.Equal(t, expectedModuleInfoMap, actual.ModInfoMap)

		// property functions
		assert.Equal(t, expectedFullTagNames[expectedModSetName], actual.ModuleFullTagNames())
		assert.Equal(t, expectedModSetVersions[expectedModSetName], actual.ModSetVersion())
		assert.Equal(t, expectedModSetPaths[expectedModSetName], actual.ModSetPaths())
	}
}

func TestVerifyTagsOnCommit(t *testing.T) {
	tmpRootDir := t.TempDir()
	repo, firstHash, err := commontest.InitNewRepoWithCommit(tmpRootDir)
	require.NoError(t, err)

	secondHash, err := common.CommitChangesToNewBranch("test_commit", "commit used in a test", repo, commontest.TestAuthor)
	require.NoError(t, err)

	createTagOptions := &git.CreateTagOptions{
		Message: "test tag message",
		Tagger:  commontest.TestAuthor,
	}

	for _, tagName := range []string{
		"test_tag_first_hash_1/v1.0.0",
		"test_tag_first_hash_2/v1.0.0",
		"test_tag_first_hash_3/v1.0.0",
	} {
		_, err = repo.CreateTag(tagName, firstHash, createTagOptions)
		require.NoError(t, err)
	}

	for _, tagName := range []string{
		"test_tag_second_hash_1/v1.0.0",
		"test_tag_second_hash_2/v1.0.0",
		"test_tag_second_hash_3/v1.0.0",
	} {
		_, err = repo.CreateTag(tagName, secondHash, createTagOptions)
		require.NoError(t, err)
	}

	testCases := []struct {
		name           string
		moduleFullTags []string
		commitHash     plumbing.Hash
		expectedError  error
	}{
		{
			name: "tags_exist",
			moduleFullTags: []string{
				"test_tag_first_hash_1/v1.0.0",
				"test_tag_first_hash_2/v1.0.0",
				"test_tag_first_hash_3/v1.0.0",
			},
			commitHash:    firstHash,
			expectedError: nil,
		},
		{
			name: "tags_do_not_exist",
			moduleFullTags: []string{
				"test_tag_first_hash_1/v2.2.2",
				"test_tag_first_hash_2/v2.2.2",
				"test_tag_first_hash_3/v2.2.2",
			},
			commitHash: firstHash,
			expectedError: &errGitTagsNotOnCommit{
				commitHash: firstHash,
				tagNames: []string{
					"test_tag_first_hash_1/v2.2.2",
					"test_tag_first_hash_2/v2.2.2",
					"test_tag_first_hash_3/v2.2.2",
				},
			},
		},
		{
			name: "some_tags_do_not_exist",
			moduleFullTags: []string{
				"test_tag_first_hash_1/v1.0.0",
				"test_tag_first_hash_2/v1.0.0",
				"test_tag_first_hash_3/v2.2.2",
			},
			commitHash: firstHash,
			expectedError: &errGitTagsNotOnCommit{
				commitHash: firstHash,
				tagNames: []string{
					"test_tag_first_hash_3/v2.2.2",
				},
			},
		},
		{
			name: "tags_on_wrong_commit",
			moduleFullTags: []string{
				"test_tag_first_hash_1/v2.2.2",
				"test_tag_first_hash_2/v2.2.2",
				"test_tag_first_hash_3/v2.2.2",
			},
			commitHash: secondHash,
			expectedError: &errGitTagsNotOnCommit{
				commitHash: secondHash,
				tagNames: []string{
					"test_tag_first_hash_1/v2.2.2",
					"test_tag_first_hash_2/v2.2.2",
					"test_tag_first_hash_3/v2.2.2",
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual := verifyTagsOnCommit(tc.moduleFullTags, repo, tc.commitHash)
			assert.Equal(t, tc.expectedError, actual)
		})

	}
}

func TestGetFullCommitHash(t *testing.T) {
	tmpRootDir := t.TempDir()
	repo, _, err := commontest.InitNewRepoWithCommit(tmpRootDir)
	require.NoError(t, err)

	fullHash, err := common.CommitChangesToNewBranch("test_commit", "commit used in a test", repo, commontest.TestAuthor)
	require.NoError(t, err)
	hashPrefix := fullHash.String()[:8]

	testCases := []struct {
		name                   string
		commitHashString       string
		expectedFullCommitHash plumbing.Hash
		expectedError          error
	}{
		{
			name:                   "prefix",
			commitHashString:       hashPrefix,
			expectedFullCommitHash: fullHash,
			expectedError:          nil,
		},
		{
			name:                   "full",
			commitHashString:       fullHash.String(),
			expectedFullCommitHash: fullHash,
			expectedError:          nil,
		},
		{
			name:                   "not_valid_commit_hash",
			commitHashString:       "12345678_cannot_be_hash",
			expectedFullCommitHash: plumbing.ZeroHash,
			expectedError:          &errCouldNotGetCommitHash{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual, err := getFullCommitHash(tc.commitHashString, repo)

			if tc.expectedError != nil {
				assert.IsType(t, tc.expectedError, err)

			} else {
				require.NoError(t, err)
			}
			assert.Equal(t, tc.expectedFullCommitHash, actual)
		})
	}
}

// integration test
func TestDeleteModuleSetTags(t *testing.T) {
	testName := "delete_module_set_tags"
	versionsYamlDir := filepath.Join(testDataDir, testName)

	tmpRootDir := t.TempDir()
	repo, _, err := commontest.InitNewRepoWithCommit(tmpRootDir)
	require.NoError(t, err)

	fullHash, err := common.CommitChangesToNewBranch("test_commit", "commit used in a test", repo, commontest.TestAuthor)
	require.NoError(t, err)
	hashPrefix := fullHash.String()[:8]

	createTagOptions := &git.CreateTagOptions{
		Message: "test tag message",
		Tagger:  commontest.TestAuthor,
	}

	tagNames := []string{
		"test/test1/v1.2.3-RC1+meta",
		"test/v0.1.0",
		"v2.2.2",
		"v2.2.0-shouldNotBeDeleted",
	}

	for _, tagName := range tagNames {
		_, err = repo.CreateTag(tagName, fullHash, createTagOptions)
		require.NoError(t, err)
	}

	modFiles := map[string][]byte{
		filepath.Join(tmpRootDir, "test", "test1", "go.mod"): []byte("module \"go.opentelemetry.io/test/test1\"\n\ngo 1.16\n\n" +
			"require (\n\t\"go.opentelemetry.io/testroot/v2\" v2.0.0\n)\n"),
		filepath.Join(tmpRootDir, "test", "go.mod"):                 []byte("module go.opentelemetry.io/test2\n\ngo 1.16\n"),
		filepath.Join(tmpRootDir, "go.mod"):                         []byte("module go.opentelemetry.io/testroot/v2\n\ngo 1.16\n"),
		filepath.Join(tmpRootDir, "test", "testexcluded", "go.mod"): []byte("module \"go.opentelemetry.io/test/testexcluded\"\n\ngo 1.16\n"),
	}

	require.NoError(t, commontest.WriteTempFiles(modFiles), "could not create go mod file tree")

	versioningFilename := filepath.Join(versionsYamlDir, "versions_valid.yaml")
	repoRoot := tmpRootDir

	tagger, err := newTagger(versioningFilename, "mod-set-3", repoRoot, hashPrefix, true)
	require.NoError(t, err)

	err = tagger.deleteModuleSetTags()
	require.NoError(t, err)

	shouldStillExist := []string{
		"test/test1/v1.2.3-RC1+meta",
		"test/v0.1.0",
		"v2.2.0-shouldNotBeDeleted",
	}

	for _, tagName := range shouldStillExist {
		tagRef, tagRefErr := repo.Tag(tagName)

		require.NoError(t, tagRefErr)
		assert.NotNil(t, tagRef)
	}

	shouldNotExist := []string{
		"v2.2.2",
		"v1.0.0-notExist",
	}

	for _, tagName := range shouldNotExist {
		tagRef, tagRefErr := repo.Tag(tagName)

		require.Error(t, tagRefErr)
		assert.Nil(t, tagRef)
	}
}

func TestDeleteTags(t *testing.T) {
	testCases := []struct {
		name           string
		moduleFullTags []string
		shouldError    bool
	}{
		{
			name: "tags_exist",
			moduleFullTags: []string{
				"test_tag_first_hash_1/v1.0.0",
				"test_tag_first_hash_2/v1.0.0",
				"test_tag_first_hash_3/v1.0.0",
			},
			shouldError: false,
		},
		{
			name: "delete_one_tag",
			moduleFullTags: []string{
				"test_tag_first_hash_1/v1.0.0",
			},
			shouldError: false,
		},
		{
			name: "tag_not_exists",
			moduleFullTags: []string{
				"tag_does_not_exist/v1.0.0",
			},
			shouldError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tmpRootDir := t.TempDir()
			repo, firstHash, err := commontest.InitNewRepoWithCommit(tmpRootDir)
			require.NoError(t, err)

			createTagOptions := &git.CreateTagOptions{
				Message: "test tag message",
				Tagger:  commontest.TestAuthor,
			}

			tagNames := []string{
				"test_tag_first_hash_1/v1.0.0",
				"test_tag_first_hash_2/v1.0.0",
				"test_tag_first_hash_3/v1.0.0",
			}

			for _, tagName := range tagNames {
				_, err = repo.CreateTag(tagName, firstHash, createTagOptions)
				require.NoError(t, err)
			}

			actualErr := deleteTags(tc.moduleFullTags, repo)
			if tc.shouldError {
				assert.Error(t, actualErr)
			} else {
				require.NoError(t, actualErr)
			}

			for _, tagName := range tagNames {
				tagRefShouldExist := true
				for _, deletedTagName := range tc.moduleFullTags {
					if tagName == deletedTagName {
						tagRefShouldExist = false
					}
				}

				tagRef, tagRefErr := repo.Tag(tagName)
				if tagRefShouldExist {
					require.NoError(t, tagRefErr)
					assert.NotNil(t, tagRef)
				} else {
					require.Error(t, tagRefErr)
					assert.Nil(t, tagRef)
				}
			}
		})

	}
}

// integration test
func TestTagAllModules(t *testing.T) {
	testName := "tag_all_modules"
	versionsYamlDir := filepath.Join(testDataDir, testName)

	versioningFilename := filepath.Join(versionsYamlDir, "versions_valid.yaml")

	createTagOptions := &git.CreateTagOptions{
		Message: "test tag message",
		Tagger:  commontest.TestAuthor,
	}

	tagNames := []string{
		"test/test1/v1.2.3-oldVersion",
		"test/test2/v0.1.0-oldVersion",
		"test/v0.1.0-oldVersion",
		"v2.2.2",
	}

	testCases := []struct {
		name               string
		modSetName         string
		shouldExistTags    []string
		shouldNotExistTags []string
		shouldError        bool
	}{
		{
			name:       "mod_set_1",
			modSetName: "mod-set-1",
			shouldExistTags: []string{
				"test/test1/v1.2.3-oldVersion",
				"test/test2/v0.1.0-oldVersion",
				"test/v0.1.0-oldVersion",
				"v2.2.2",

				"test/test1/v1.2.3-RC1+meta",
			},
			shouldNotExistTags: []string{
				"test/test2/v0.1.0",
				"test/v0.1.0",
				"v1.0.0-doesNotExist",
			},
			shouldError: false,
		},
		{
			name:       "mod_set_2_multiple",
			modSetName: "mod-set-2",
			shouldExistTags: []string{
				"test/test1/v1.2.3-oldVersion",
				"test/test2/v0.1.0-oldVersion",
				"test/v0.1.0-oldVersion",
				"v2.2.2",

				"test/test2/v0.1.0",
				"test/v0.1.0",
			},
			shouldNotExistTags: []string{
				"test/test1/v1.2.3-RC1+meta",
				"v1.0.0-doesNotExist",
			},
			shouldError: false,
		},
		{
			name:       "mod_set_3_already_exists",
			modSetName: "mod-set-3",
			shouldExistTags: []string{
				"test/test1/v1.2.3-oldVersion",
				"test/test2/v0.1.0-oldVersion",
				"test/v0.1.0-oldVersion",
				"v2.2.2",
			},
			shouldNotExistTags: []string{
				"test/test1/v1.2.3-RC1+meta",
				"test/test2/v0.1.0",
				"test/v0.1.0",
				"v1.0.0-doesNotExist",
			},
			shouldError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tmpRootDir := t.TempDir()
			repo, _, err := commontest.InitNewRepoWithCommit(tmpRootDir)
			require.NoError(t, err)

			fullHash, err := common.CommitChangesToNewBranch("test_commit", "commit used in a test", repo, commontest.TestAuthor)
			require.NoError(t, err)
			hashPrefix := fullHash.String()[:8]

			modFiles := map[string][]byte{
				filepath.Join(tmpRootDir, "test", "test1", "go.mod"): []byte("module \"go.opentelemetry.io/test/test1\"\n\ngo 1.16\n\n" +
					"require (\n\t\"go.opentelemetry.io/testroot/v2\" v2.0.0\n)\n"),
				filepath.Join(tmpRootDir, "test", "test2", "go.mod"):        []byte("module go.opentelemetry.io/test2\n\ngo 1.16\n"),
				filepath.Join(tmpRootDir, "test", "go.mod"):                 []byte("module go.opentelemetry.io/test3\n\ngo 1.16\n"),
				filepath.Join(tmpRootDir, "go.mod"):                         []byte("module go.opentelemetry.io/testroot/v2\n\ngo 1.16\n"),
				filepath.Join(tmpRootDir, "test", "testexcluded", "go.mod"): []byte("module \"go.opentelemetry.io/test/testexcluded\"\n\ngo 1.16\n"),
			}

			require.NoError(t, commontest.WriteTempFiles(modFiles), "could not create go mod file tree")

			for _, tagName := range tagNames {
				_, err = repo.CreateTag(tagName, fullHash, createTagOptions)
				require.NoError(t, err)
			}

			tagger, err := newTagger(versioningFilename, tc.modSetName, tmpRootDir, hashPrefix, false)
			if tc.shouldError {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.NoError(t, tagger.tagAllModules(commontest.TestAuthor))
			for _, tagName := range tc.shouldExistTags {
				tagRef, tagRefErr := repo.Tag(tagName)

				require.NoErrorf(t, tagRefErr, "tag name %v not found but should exist", tagName)
				assert.NotNil(t, tagRef)
			}

			for _, tagName := range tc.shouldNotExistTags {
				tagRef, tagRefErr := repo.Tag(tagName)

				require.Error(t, tagRefErr)
				assert.Nil(t, tagRef)
			}
		})
	}

}
