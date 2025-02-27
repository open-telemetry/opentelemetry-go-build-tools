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

package prerelease

import (
	"io"
	"log"
	"os"
	"path/filepath"
	"testing"

	"github.com/go-git/go-git/v5"
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

func TestNewPrerelease(t *testing.T) {
	testName := "new_prerelease"
	versionsYamlDir := filepath.Join(testDataDir, testName)

	tmpRootDir := t.TempDir()
	modFiles := map[string][]byte{
		filepath.Join(tmpRootDir, "test", "test1", "go.mod"): []byte("module \"go.opentelemetry.io/test/test1\"\n\ngo 1.16\n\n" +
			"require (\n\t\"go.opentelemetry.io/testroot/v2\" v2.0.0\n)\n"),
		filepath.Join(tmpRootDir, "test", "go.mod"):          []byte("module go.opentelemetry.io/test2\n\ngo 1.16\n"),
		filepath.Join(tmpRootDir, "go.mod"):                  []byte("module go.opentelemetry.io/testroot/v2\n\ngo 1.16\n"),
		filepath.Join(tmpRootDir, "test", "test2", "go.mod"): []byte("module \"go.opentelemetry.io/test/testexcluded\"\n\ngo 1.16\n"),
	}

	require.NoError(t, commontest.WriteTempFiles(modFiles), "could not create go mod file tree")

	// initialize temporary local git repository
	_, err := git.PlainInit(tmpRootDir, true)
	require.NoError(t, err, "could not initialize temp git repo")

	testCases := []struct {
		name                   string
		versioningFilename     string
		repoRoot               string
		shouldError            bool
		expectedModuleSetMap   common.ModuleSetMap
		expectedModulePathMap  common.ModulePathMap
		expectedModuleInfoMap  common.ModuleInfoMap
		expectedTagNames       map[string][]common.ModuleTagName
		expectedFullTagNames   map[string][]string
		expectedModSetVersions map[string]string
		expectedModSetPaths    map[string][]common.ModulePath
	}{
		{
			name:               "valid versioning",
			versioningFilename: filepath.Join(versionsYamlDir, "/versions_valid.yaml"),
			repoRoot:           tmpRootDir,
			shouldError:        false,
			expectedModuleSetMap: common.ModuleSetMap{
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
			},
			expectedModulePathMap: common.ModulePathMap{
				"go.opentelemetry.io/test/test1":  common.ModuleFilePath(filepath.Join(tmpRootDir, "test", "test1", "go.mod")),
				"go.opentelemetry.io/test2":       common.ModuleFilePath(filepath.Join(tmpRootDir, "test", "go.mod")),
				"go.opentelemetry.io/testroot/v2": common.ModuleFilePath(filepath.Join(tmpRootDir, "go.mod")),
			},
			expectedModuleInfoMap: common.ModuleInfoMap{
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
			},
			expectedTagNames: map[string][]common.ModuleTagName{
				"mod-set-1": {"test/test1"},
				"mod-set-2": {"test"},
				"mod-set-3": {common.RepoRootTag},
			},
			expectedFullTagNames: map[string][]string{
				"mod-set-1": {"test/test1/v1.2.3-RC1+meta"},
				"mod-set-2": {"test/v0.1.0"},
				"mod-set-3": {"v2.2.2"},
			},
			expectedModSetVersions: map[string]string{
				"mod-set-1": "v1.2.3-RC1+meta",
				"mod-set-2": "v0.1.0",
				"mod-set-3": "v2.2.2",
			},
			expectedModSetPaths: map[string][]common.ModulePath{
				"mod-set-1": {"go.opentelemetry.io/test/test1"},
				"mod-set-2": {"go.opentelemetry.io/test2"},
				"mod-set-3": {"go.opentelemetry.io/testroot/v2"},
			},
		},
		{
			name:                  "invalid version file syntax",
			versioningFilename:    filepath.Join(versionsYamlDir, "versions_invalid_syntax.yaml"),
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
				actual, err := newPrerelease(tc.versioningFilename, expectedModSetName, tc.repoRoot)

				if tc.shouldError {
					assert.Error(t, err)
				} else {
					require.NoError(t, err)
				}

				assert.IsType(t, prerelease{}, actual)
				assert.IsType(t, common.ModuleSetRelease{}, actual.ModuleSetRelease)
				assert.Equal(t, tc.expectedTagNames[expectedModSetName], actual.ModuleSetRelease.TagNames)
				assert.Equal(t, expectedModSet, actual.ModuleSetRelease.ModSet)
				assert.Equal(t, expectedModSetName, actual.ModuleSetRelease.ModSetName)

				assert.IsType(t, common.ModuleVersioning{}, actual.ModuleSetRelease.ModuleVersioning)
				assert.Equal(t, tc.expectedModuleSetMap, actual.ModuleSetRelease.ModuleVersioning.ModSetMap)
				assert.Equal(t, tc.expectedModulePathMap, actual.ModuleSetRelease.ModuleVersioning.ModPathMap)
				assert.Equal(t, tc.expectedModuleInfoMap, actual.ModuleSetRelease.ModuleVersioning.ModInfoMap)

				// property functions
				assert.Equal(t, tc.expectedFullTagNames[expectedModSetName], actual.ModuleSetRelease.ModuleFullTagNames())
				assert.Equal(t, tc.expectedModSetVersions[expectedModSetName], actual.ModuleSetRelease.ModSetVersion())
				assert.Equal(t, tc.expectedModSetPaths[expectedModSetName], actual.ModuleSetRelease.ModSetPaths())
			}
		})
	}
}

func TestUpdateAllVersionGo(t *testing.T) {
	testName := "update_all_version_go"
	versionsYamlDir := filepath.Join(testDataDir, testName)

	versioningFilename := filepath.Join(versionsYamlDir, "versions_valid.yaml")

	tmpRootDir := t.TempDir()
	modFiles := map[string][]byte{
		filepath.Join(tmpRootDir, "test", "test1", "go.mod"): []byte("module \"go.opentelemetry.io/test/test1\"\n\ngo 1.16\n\n" +
			"require (\n\t\"go.opentelemetry.io/testroot/v2\" v2.2.2\n)\n"),
		filepath.Join(tmpRootDir, "test", "go.mod"):          []byte("module go.opentelemetry.io/test2\n\ngo 1.16\n"),
		filepath.Join(tmpRootDir, "go.mod"):                  []byte("module go.opentelemetry.io/testroot/v2\n\ngo 1.16\n"),
		filepath.Join(tmpRootDir, "test", "test2", "go.mod"): []byte("module \"go.opentelemetry.io/test/testexcluded\"\n\ngo 1.16\n"),
	}

	versionGoFiles := map[string][]byte{
		filepath.Join(tmpRootDir, "test", "test1", "version.go"): []byte("package test1 // import \"go.opentelemetry.io/test/test1\"\n\n" +
			"// Version is the current release version of OpenTelemetry in use.\n" +
			"func Version() string {\n\t" +
			"return \"1.0.0-OLD\"\n" +
			"}\n"),
		filepath.Join(tmpRootDir, "test", "version.go"): []byte("package test2 // import \"go.opentelemetry.io/test/test2\"\n\n" +
			"// version is the current release version of OpenTelemetry in use.\n" +
			"func version() string {\n\t" +
			"return \"0.1.0-OLD\"\n" +
			"}\n"),
	}

	testCases := []struct {
		name                     string
		modSetName               string
		expectedVersionGoOutputs map[string][]byte
	}{
		{
			name:       "update_version_1",
			modSetName: "mod-set-1",
			expectedVersionGoOutputs: map[string][]byte{
				filepath.Join(tmpRootDir, "test", "test1", "version.go"): []byte("package test1 // import \"go.opentelemetry.io/test/test1\"\n\n" +
					"// Version is the current release version of OpenTelemetry in use.\n" +
					"func Version() string {\n\t" +
					"return \"1.2.3-RC1+meta\"\n" +
					"}\n"),
				filepath.Join(tmpRootDir, "test", "version.go"): []byte("package test2 // import \"go.opentelemetry.io/test/test2\"\n\n" +
					"// version is the current release version of OpenTelemetry in use.\n" +
					"func version() string {\n\t" +
					"return \"0.1.0-OLD\"\n" +
					"}\n"),
			},
		},
		{
			name:       "update_version_2",
			modSetName: "mod-set-2",
			expectedVersionGoOutputs: map[string][]byte{
				filepath.Join(tmpRootDir, "test", "test1", "version.go"): []byte("package test1 // import \"go.opentelemetry.io/test/test1\"\n\n" +
					"// Version is the current release version of OpenTelemetry in use.\n" +
					"func Version() string {\n\t" +
					"return \"1.0.0-OLD\"\n" +
					"}\n"),
				filepath.Join(tmpRootDir, "test", "version.go"): []byte("package test2 // import \"go.opentelemetry.io/test/test2\"\n\n" +
					"// version is the current release version of OpenTelemetry in use.\n" +
					"func version() string {\n\t" +
					"return \"0.1.0\"\n" +
					"}\n"),
			},
		},
		{
			name:       "no_version_go_in_set",
			modSetName: "mod-set-3",
			expectedVersionGoOutputs: map[string][]byte{
				filepath.Join(tmpRootDir, "test", "test1", "version.go"): []byte("package test1 // import \"go.opentelemetry.io/test/test1\"\n\n" +
					"// Version is the current release version of OpenTelemetry in use.\n" +
					"func Version() string {\n\t" +
					"return \"1.0.0-OLD\"\n" +
					"}\n"),
				filepath.Join(tmpRootDir, "test", "version.go"): []byte("package test2 // import \"go.opentelemetry.io/test/test2\"\n\n" +
					"// version is the current release version of OpenTelemetry in use.\n" +
					"func version() string {\n\t" +
					"return \"0.1.0-OLD\"\n" +
					"}\n"),
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			require.NoError(t, commontest.WriteTempFiles(modFiles), "could not create go mod file tree")
			require.NoError(t, commontest.WriteTempFiles(versionGoFiles), "could not create version.go file tree")

			p, err := newPrerelease(versioningFilename, tc.modSetName, tmpRootDir)
			require.NoError(t, err)

			err = p.updateAllVersionGo()
			require.NoError(t, err)

			for versionGoFilePath, expectedByteOutput := range tc.expectedVersionGoOutputs {
				actual, err := os.ReadFile(filepath.Clean(versionGoFilePath))
				require.NoError(t, err)

				assert.Equal(t, expectedByteOutput, actual)
			}
		})
	}
}

func TestUpdateAllGoModFiles(t *testing.T) {
	testName := "update_all_go_mod_files"
	versionsYamlDir := filepath.Join(testDataDir, testName)

	testCases := []struct {
		modSetName             string
		expectedOutputModFiles map[string][]byte
	}{
		{
			modSetName: "mod-set-1",
			expectedOutputModFiles: map[string][]byte{
				filepath.Join("test", "test1", "go.mod"): []byte("module go.opentelemetry.io/build-tools/multimod/internal/prerelease/test/test1\n\n" +
					"go 1.16\n\n" +
					"require (\n\t" +
					"go.opentelemetry.io/build-tools/multimod/internal/prerelease/test/test2 v1.2.3-RC1+meta\n\t" +
					"go.opentelemetry.io/other/test/test1 v1.0.0\n\t" +
					"go.opentelemetry.io/other/testroot/v2 v2.2.2\n" +
					")"),
				filepath.Join("test", "test2", "go.mod"): []byte("module go.opentelemetry.io/build-tools/multimod/internal/prerelease/test/test2\n\n" +
					"go 1.16\n\n" +
					"require (\n\t" +
					"go.opentelemetry.io/build-tools/multimod/internal/prerelease/test/test1 v1.2.3-RC1+meta\n\t" +
					"go.opentelemetry.io/other/test/test1 v1.0.0\n\t" +
					"go.opentelemetry.io/other/testroot/v2 v2.2.2\n" +
					")"),
				filepath.Join("test", "go.mod"): []byte("module go.opentelemetry.io/build-tools/multimod/internal/prerelease/test3\n\n" +
					"go 1.16\n\n" +
					"require (\n\t" +
					"go.opentelemetry.io/build-tools/multimod/internal/prerelease/test/test1 v1.2.3-RC1+meta\n\t" +
					"go.opentelemetry.io/build-tools/multimod/internal/prerelease/test/test2 v1.2.3-RC1+meta\n\t" +
					"go.opentelemetry.io/build-tools/multimod/internal/prerelease/testroot v0.1.0-shouldBe2\n\t" +
					"go.opentelemetry.io/other/test2 v0.1.0\n" +
					")"),
				"go.mod": []byte("module go.opentelemetry.io/build-tools/multimod/internal/prerelease/testroot\n\n" +
					"go 1.16\n\n" +
					"require (\n\t" +
					"go.opentelemetry.io/build-tools/multimod/internal/prerelease/test/test1 v1.2.3-RC1+meta\n\t" +
					"go.opentelemetry.io/build-tools/multimod/internal/prerelease/test/test2 v1.2.3-RC1+meta\n\t" +
					"go.opentelemetry.io/build-tools/multimod/internal/prerelease/test3 v0.1.0-OLD\n\t" +
					"go.opentelemetry.io/other/test/test1 v1.0.0\n\t" +
					")"),
				filepath.Join("excluded", "go.mod"): []byte("module go.opentelemetry.io/my/test/testexcluded\n\n" +
					"go 1.16\n\n" +
					"require (\n\t" +
					"go.opentelemetry.io/build-tools/multimod/internal/prerelease/test/test2 v1.2.3-RC1+meta\n\t" +
					"go.opentelemetry.io/other/test/test1 v1.0.0\n\t" +
					"go.opentelemetry.io/other/testroot/v2 v2.2.2\n" +
					")"),
			},
		},
		{
			modSetName: "mod-set-2",
			expectedOutputModFiles: map[string][]byte{
				filepath.Join("test", "test1", "go.mod"): []byte("module go.opentelemetry.io/build-tools/multimod/internal/prerelease/test/test1\n\n" +
					"go 1.16\n\n" +
					"require (\n\t" +
					"go.opentelemetry.io/build-tools/multimod/internal/prerelease/test/test2 v1.2.3-OLD\n\t" +
					"go.opentelemetry.io/other/test/test1 v1.0.0\n\t" +
					"go.opentelemetry.io/other/testroot/v2 v2.2.2\n" +
					")"),
				filepath.Join("test", "test2", "go.mod"): []byte("module go.opentelemetry.io/build-tools/multimod/internal/prerelease/test/test2\n\n" +
					"go 1.16\n\n" +
					"require (\n\t" +
					"go.opentelemetry.io/build-tools/multimod/internal/prerelease/test/test1 v1.2.3-OLD\n\t" +
					"go.opentelemetry.io/other/test/test1 v1.0.0\n\t" +
					"go.opentelemetry.io/other/testroot/v2 v2.2.2\n" +
					")"),
				filepath.Join("test", "go.mod"): []byte("module go.opentelemetry.io/build-tools/multimod/internal/prerelease/test3\n\n" +
					"go 1.16\n\n" +
					"require (\n\t" +
					"go.opentelemetry.io/build-tools/multimod/internal/prerelease/test/test1 v1.2.3-OLD\n\t" +
					"go.opentelemetry.io/build-tools/multimod/internal/prerelease/test/test2 v1.2.3-OLD\n\t" +
					"go.opentelemetry.io/build-tools/multimod/internal/prerelease/testroot v0.1.0-shouldBe2\n\t" +
					"go.opentelemetry.io/other/test2 v0.1.0\n" +
					")"),
				"go.mod": []byte("module go.opentelemetry.io/build-tools/multimod/internal/prerelease/testroot\n\n" +
					"go 1.16\n\n" +
					"require (\n\t" +
					"go.opentelemetry.io/build-tools/multimod/internal/prerelease/test/test1 v1.2.3-OLD\n\t" +
					"go.opentelemetry.io/build-tools/multimod/internal/prerelease/test/test2 v1.2.3-OLD\n\t" +
					"go.opentelemetry.io/build-tools/multimod/internal/prerelease/test3 v0.1.0\n\t" +
					"go.opentelemetry.io/other/test/test1 v1.0.0\n\t" +
					")"),
				filepath.Join("excluded", "go.mod"): []byte("module go.opentelemetry.io/my/test/testexcluded\n\n" +
					"go 1.16\n\n" +
					"require (\n\t" +
					"go.opentelemetry.io/build-tools/multimod/internal/prerelease/test/test2 v1.2.3-OLD\n\t" +
					"go.opentelemetry.io/other/test/test1 v1.0.0\n\t" +
					"go.opentelemetry.io/other/testroot/v2 v2.2.2\n" +
					")"),
			},
		},
		{
			modSetName: "mod-set-3",
			expectedOutputModFiles: map[string][]byte{
				filepath.Join("test", "test1", "go.mod"): []byte("module go.opentelemetry.io/build-tools/multimod/internal/prerelease/test/test1\n\n" +
					"go 1.16\n\n" +
					"require (\n\t" +
					"go.opentelemetry.io/build-tools/multimod/internal/prerelease/test/test2 v1.2.3-OLD\n\t" +
					"go.opentelemetry.io/other/test/test1 v1.0.0\n\t" +
					"go.opentelemetry.io/other/testroot/v2 v2.2.2\n" +
					")"),
				filepath.Join("test", "test2", "go.mod"): []byte("module go.opentelemetry.io/build-tools/multimod/internal/prerelease/test/test2\n\n" +
					"go 1.16\n\n" +
					"require (\n\t" +
					"go.opentelemetry.io/build-tools/multimod/internal/prerelease/test/test1 v1.2.3-OLD\n\t" +
					"go.opentelemetry.io/other/test/test1 v1.0.0\n\t" +
					"go.opentelemetry.io/other/testroot/v2 v2.2.2\n" +
					")"),
				filepath.Join("test", "go.mod"): []byte("module go.opentelemetry.io/build-tools/multimod/internal/prerelease/test3\n\n" +
					"go 1.16\n\n" +
					"require (\n\t" +
					"go.opentelemetry.io/build-tools/multimod/internal/prerelease/test/test1 v1.2.3-OLD\n\t" +
					"go.opentelemetry.io/build-tools/multimod/internal/prerelease/test/test2 v1.2.3-OLD\n\t" +
					"go.opentelemetry.io/build-tools/multimod/internal/prerelease/testroot v0.2.0\n\t" +
					"go.opentelemetry.io/other/test2 v0.1.0\n" +
					")"),
				"go.mod": []byte("module go.opentelemetry.io/build-tools/multimod/internal/prerelease/testroot\n\n" +
					"go 1.16\n\n" +
					"require (\n\t" +
					"go.opentelemetry.io/build-tools/multimod/internal/prerelease/test/test1 v1.2.3-OLD\n\t" +
					"go.opentelemetry.io/build-tools/multimod/internal/prerelease/test/test2 v1.2.3-OLD\n\t" +
					"go.opentelemetry.io/build-tools/multimod/internal/prerelease/test3 v0.1.0-OLD\n\t" +
					"go.opentelemetry.io/other/test/test1 v1.0.0\n\t" +
					")"),
				filepath.Join("excluded", "go.mod"): []byte("module go.opentelemetry.io/my/test/testexcluded\n\n" +
					"go 1.16\n\n" +
					"require (\n\t" +
					"go.opentelemetry.io/build-tools/multimod/internal/prerelease/test/test2 v1.2.3-OLD\n\t" +
					"go.opentelemetry.io/other/test/test1 v1.0.0\n\t" +
					"go.opentelemetry.io/other/testroot/v2 v2.2.2\n" +
					")"),
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.modSetName, func(t *testing.T) {
			versioningFilename := filepath.Join(versionsYamlDir, "versions_valid.yaml")

			tmpRootDir := t.TempDir()
			modFiles := map[string][]byte{
				filepath.Join(tmpRootDir, "test", "test1", "go.mod"): []byte("module go.opentelemetry.io/build-tools/multimod/internal/prerelease/test/test1\n\n" +
					"go 1.16\n\n" +
					"require (\n\t" +
					"go.opentelemetry.io/build-tools/multimod/internal/prerelease/test/test2 v1.2.3-OLD\n\t" +
					"go.opentelemetry.io/other/test/test1 v1.0.0\n\t" +
					"go.opentelemetry.io/other/testroot/v2 v2.2.2\n" +
					")"),
				filepath.Join(tmpRootDir, "test", "test2", "go.mod"): []byte("module go.opentelemetry.io/build-tools/multimod/internal/prerelease/test/test2\n\n" +
					"go 1.16\n\n" +
					"require (\n\t" +
					"go.opentelemetry.io/build-tools/multimod/internal/prerelease/test/test1 v1.2.3-OLD\n\t" +
					"go.opentelemetry.io/other/test/test1 v1.0.0\n\t" +
					"go.opentelemetry.io/other/testroot/v2 v2.2.2\n" +
					")"),
				filepath.Join(tmpRootDir, "test", "go.mod"): []byte("module go.opentelemetry.io/build-tools/multimod/internal/prerelease/test3\n\n" +
					"go 1.16\n\n" +
					"require (\n\t" +
					"go.opentelemetry.io/build-tools/multimod/internal/prerelease/test/test1 v1.2.3-OLD\n\t" +
					"go.opentelemetry.io/build-tools/multimod/internal/prerelease/test/test2 v1.2.3-OLD\n\t" +
					"go.opentelemetry.io/build-tools/multimod/internal/prerelease/testroot v0.1.0-shouldBe2\n\t" +
					"go.opentelemetry.io/other/test2 v0.1.0\n" +
					")"),
				filepath.Join(tmpRootDir, "go.mod"): []byte("module go.opentelemetry.io/build-tools/multimod/internal/prerelease/testroot\n\n" +
					"go 1.16\n\n" +
					"require (\n\t" +
					"go.opentelemetry.io/build-tools/multimod/internal/prerelease/test/test1 v1.2.3-OLD\n\t" +
					"go.opentelemetry.io/build-tools/multimod/internal/prerelease/test/test2 v1.2.3-OLD\n\t" +
					"go.opentelemetry.io/build-tools/multimod/internal/prerelease/test3 v0.1.0-OLD\n\t" +
					"go.opentelemetry.io/other/test/test1 v1.0.0\n\t" +
					")"),
				filepath.Join(tmpRootDir, "excluded", "go.mod"): []byte("module go.opentelemetry.io/my/test/testexcluded\n\n" +
					"go 1.16\n\n" +
					"require (\n\t" +
					"go.opentelemetry.io/build-tools/multimod/internal/prerelease/test/test2 v1.2.3-OLD\n\t" +
					"go.opentelemetry.io/other/test/test1 v1.0.0\n\t" +
					"go.opentelemetry.io/other/testroot/v2 v2.2.2\n" +
					")"),
			}

			require.NoError(t, commontest.WriteTempFiles(modFiles), "could not create go mod file tree")

			p, err := newPrerelease(versioningFilename, tc.modSetName, tmpRootDir)
			require.NoError(t, err)

			err = p.updateAllGoModFiles()
			require.NoError(t, err)

			for modFilePathSuffix, expectedByteOutput := range tc.expectedOutputModFiles {
				actual, err := os.ReadFile(filepath.Clean(filepath.Join(tmpRootDir, modFilePathSuffix)))
				require.NoError(t, err)

				assert.Equal(t, expectedByteOutput, actual)
			}
		})
	}
}
