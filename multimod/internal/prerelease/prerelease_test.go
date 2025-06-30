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

	"go.opentelemetry.io/build-tools/multimod/internal/shared"
	"go.opentelemetry.io/build-tools/multimod/internal/shared/sharedtest"
)

var testDataDir, _ = filepath.Abs("./test_data")

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

	require.NoError(t, sharedtest.WriteTempFiles(modFiles), "could not create go mod file tree")

	// initialize temporary local git repository
	_, err := git.PlainInit(tmpRootDir, true)
	require.NoError(t, err, "could not initialize temp git repo")

	testCases := []struct {
		name                   string
		versioningFilename     string
		repoRoot               string
		shouldError            bool
		expectedModuleSetMap   shared.ModuleSetMap
		expectedModulePathMap  shared.ModulePathMap
		expectedModuleInfoMap  shared.ModuleInfoMap
		expectedTagNames       map[string][]shared.ModuleTagName
		expectedFullTagNames   map[string][]string
		expectedModSetVersions map[string]string
		expectedModSetPaths    map[string][]shared.ModulePath
	}{
		{
			name:               "valid versioning",
			versioningFilename: filepath.Join(versionsYamlDir, "/versions_valid.yaml"),
			repoRoot:           tmpRootDir,
			shouldError:        false,
			expectedModuleSetMap: shared.ModuleSetMap{
				"mod-set-1": shared.ModuleSet{
					Version: "v1.2.3-RC1+meta",
					Modules: []shared.ModulePath{
						"go.opentelemetry.io/test/test1",
					},
				},
				"mod-set-2": shared.ModuleSet{
					Version: "v0.1.0",
					Modules: []shared.ModulePath{
						"go.opentelemetry.io/test2",
					},
				},
				"mod-set-3": shared.ModuleSet{
					Version: "v2.2.2",
					Modules: []shared.ModulePath{
						"go.opentelemetry.io/testroot/v2",
					},
				},
			},
			expectedModulePathMap: shared.ModulePathMap{
				"go.opentelemetry.io/test/test1":  shared.ModuleFilePath(filepath.Join(tmpRootDir, "test", "test1", "go.mod")),
				"go.opentelemetry.io/test2":       shared.ModuleFilePath(filepath.Join(tmpRootDir, "test", "go.mod")),
				"go.opentelemetry.io/testroot/v2": shared.ModuleFilePath(filepath.Join(tmpRootDir, "go.mod")),
			},
			expectedModuleInfoMap: shared.ModuleInfoMap{
				"go.opentelemetry.io/test/test1": shared.ModuleInfo{
					ModuleSetName: "mod-set-1",
					Version:       "v1.2.3-RC1+meta",
				},
				"go.opentelemetry.io/testroot/v2": shared.ModuleInfo{
					ModuleSetName: "mod-set-3",
					Version:       "v2.2.2",
				},
				"go.opentelemetry.io/test2": shared.ModuleInfo{
					ModuleSetName: "mod-set-2",
					Version:       "v0.1.0",
				},
			},
			expectedTagNames: map[string][]shared.ModuleTagName{
				"mod-set-1": {"test/test1"},
				"mod-set-2": {"test"},
				"mod-set-3": {shared.RepoRootTag},
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
			expectedModSetPaths: map[string][]shared.ModulePath{
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
				assert.IsType(t, shared.ModuleSetRelease{}, actual.ModuleSetRelease)
				assert.Equal(t, tc.expectedTagNames[expectedModSetName], actual.TagNames)
				assert.Equal(t, expectedModSet, actual.ModSet)
				assert.Equal(t, expectedModSetName, actual.ModSetName)

				assert.IsType(t, shared.ModuleVersioning{}, actual.ModuleVersioning)
				assert.Equal(t, tc.expectedModuleSetMap, actual.ModSetMap)
				assert.Equal(t, tc.expectedModulePathMap, actual.ModPathMap)
				assert.Equal(t, tc.expectedModuleInfoMap, actual.ModInfoMap)

				// property functions
				assert.Equal(t, tc.expectedFullTagNames[expectedModSetName], actual.ModuleFullTagNames())
				assert.Equal(t, tc.expectedModSetVersions[expectedModSetName], actual.ModSetVersion())
				assert.Equal(t, tc.expectedModSetPaths[expectedModSetName], actual.ModSetPaths())
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
			require.NoError(t, sharedtest.WriteTempFiles(modFiles), "could not create go mod file tree")
			require.NoError(t, sharedtest.WriteTempFiles(versionGoFiles), "could not create version.go file tree")

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

			require.NoError(t, sharedtest.WriteTempFiles(modFiles), "could not create go mod file tree")

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

func TestUpdateAll(t *testing.T) {
	origCC := commitChanges
	commit = func(shared.ModuleSetRelease, bool, *git.Repository) error { return nil }
	t.Cleanup(func() { commit = origCC })

	origWTC := workingTreeClean
	workingTreeClean = func(*git.Repository) error { return nil }
	t.Cleanup(func() { workingTreeClean = origWTC })

	dir := filepath.Join(testDataDir, "update_all")
	vFile := filepath.Join(dir, "versions.yaml")

	root := t.TempDir()
	_, err := git.PlainInit(root, false)
	require.NoError(t, err, "could not initialize temp git repo")

	origFR := findRoot
	findRoot = func() (string, error) { return root, nil }
	t.Cleanup(func() { findRoot = origFR })

	modFiles := map[string][]byte{
		filepath.Join(root, "test", "test1", "go.mod"): []byte("module go.opentelemetry.io/all/test/test1\n\n" +
			"go 1.16\n\n" +
			"require (\n\t" +
			"go.opentelemetry.io/all/test/test2 v0.0.1\n\t" +
			"go.opentelemetry.io/all/v2 v2.2.2\n" +
			"go.opentelemetry.io/other/test/test1 v1.0.0\n\t" +
			"go.opentelemetry.io/other/v2 v2.2.2\n" +
			")"),
		filepath.Join(root, "test", "test2", "go.mod"): []byte("module go.opentelemetry.io/all/test/test2\n\n" +
			"go 1.16\n\n" +
			"require (\n\t" +
			"go.opentelemetry.io/all/test/test1 v1.2.3-OLD\n\t" +
			"go.opentelemetry.io/other/test/test1 v1.0.0\n\t" +
			"go.opentelemetry.io/other/v2 v2.2.2\n" +
			")"),
		filepath.Join(root, "test", "go.mod"): []byte("module go.opentelemetry.io/all/test3\n\n" +
			"go 1.16\n\n" +
			"require (\n\t" +
			"go.opentelemetry.io/all/test/test1 v1.2.3-OLD\n\t" +
			"go.opentelemetry.io/all/test/test2 v0.0.1\n\t" +
			"go.opentelemetry.io/all v0.1.0-shouldBe2\n\t" +
			"go.opentelemetry.io/other/test2 v0.1.0\n" +
			")"),
		filepath.Join(root, "go.mod"): []byte("module go.opentelemetry.io/all\n\n" +
			"go 1.16\n\n" +
			"require (\n\t" +
			"go.opentelemetry.io/all/test/test1 v1.2.3-OLD\n\t" +
			"go.opentelemetry.io/all/test/test2 v0.0.1\n\t" +
			"go.opentelemetry.io/all/test3 v0.1.0-OLD\n\t" +
			"go.opentelemetry.io/other/test/test1 v1.0.0\n\t" +
			")"),
		filepath.Join(root, "v2", "go.mod"): []byte("module go.opentelemetry.io/all/v2\n\n" +
			"go 1.16\n\n" +
			"require (\n\t" +
			"go.opentelemetry.io/all/test/test1 v1.2.3-OLD\n\t" +
			"go.opentelemetry.io/all/test/test2 v0.0.1\n\t" +
			"go.opentelemetry.io/all/test3 v0.1.0-OLD\n\t" +
			"go.opentelemetry.io/other/test/test1 v1.0.0\n\t" +
			")"),
		filepath.Join(root, "excluded", "go.mod"): []byte("module go.opentelemetry.io/all/test/testexcluded\n\n" +
			"go 1.16\n\n" +
			"require (\n\t" +
			"go.opentelemetry.io/all/test/test2 v0.0.1\n\t" +
			"go.opentelemetry.io/other/test/test1 v1.0.0\n\t" +
			"go.opentelemetry.io/other/v2 v2.2.2\n" +
			")"),
	}

	require.NoError(t, sharedtest.WriteTempFiles(modFiles))
	require.NoError(t, run(vFile, nil, true, false))

	expected := map[string][]byte{
		filepath.Join("test", "test1", "go.mod"): []byte("module go.opentelemetry.io/all/test/test1\n\n" +
			"go 1.16\n\n" +
			"require (\n\t" +
			"go.opentelemetry.io/all/test/test2 v0.1.0\n\t" +
			"go.opentelemetry.io/all/v2 v2.2.2\n" +
			"go.opentelemetry.io/other/test/test1 v1.0.0\n\t" +
			"go.opentelemetry.io/other/v2 v2.2.2\n" +
			")"),
		filepath.Join("test", "test2", "go.mod"): []byte("module go.opentelemetry.io/all/test/test2\n\n" +
			"go 1.16\n\n" +
			"require (\n\t" +
			"go.opentelemetry.io/all/test/test1 v1.2.3-RC1+meta\n\t" +
			"go.opentelemetry.io/other/test/test1 v1.0.0\n\t" +
			"go.opentelemetry.io/other/v2 v2.2.2\n" +
			")"),
		filepath.Join("test", "go.mod"): []byte("module go.opentelemetry.io/all/test3\n\n" +
			"go 1.16\n\n" +
			"require (\n\t" +
			"go.opentelemetry.io/all/test/test1 v1.2.3-RC1+meta\n\t" +
			"go.opentelemetry.io/all/test/test2 v0.1.0\n\t" +
			"go.opentelemetry.io/all v0.1.0\n\t" +
			"go.opentelemetry.io/other/test2 v0.1.0\n" +
			")"),
		"go.mod": []byte("module go.opentelemetry.io/all\n\n" +
			"go 1.16\n\n" +
			"require (\n\t" +
			"go.opentelemetry.io/all/test/test1 v1.2.3-RC1+meta\n\t" +
			"go.opentelemetry.io/all/test/test2 v0.1.0\n\t" +
			"go.opentelemetry.io/all/test3 v1.2.3-RC1+meta\n\t" +
			"go.opentelemetry.io/other/test/test1 v1.0.0\n\t" +
			")"),
		filepath.Join("v2", "go.mod"): []byte("module go.opentelemetry.io/all/v2\n\n" +
			"go 1.16\n\n" +
			"require (\n\t" +
			"go.opentelemetry.io/all/test/test1 v1.2.3-RC1+meta\n\t" +
			"go.opentelemetry.io/all/test/test2 v0.1.0\n\t" +
			"go.opentelemetry.io/all/test3 v1.2.3-RC1+meta\n\t" +
			"go.opentelemetry.io/other/test/test1 v1.0.0\n\t" +
			")"),
		filepath.Join("excluded", "go.mod"): []byte("module go.opentelemetry.io/all/test/testexcluded\n\n" +
			"go 1.16\n\n" +
			"require (\n\t" +
			"go.opentelemetry.io/all/test/test2 v0.1.0\n\t" +
			"go.opentelemetry.io/other/test/test1 v1.0.0\n\t" +
			"go.opentelemetry.io/other/v2 v2.2.2\n" +
			")"),
	}
	for file, want := range expected {
		path := filepath.Clean(filepath.Join(root, file))
		got, err := os.ReadFile(path)
		require.NoError(t, err)

		assert.Equalf(
			t,
			string(want), string(got),
			"file %s does not match expected output", file,
		)
	}
}
