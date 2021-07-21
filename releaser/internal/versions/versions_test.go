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

package versions

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestMain performs setup for the tests and suppress printing logs.
func TestMain(m *testing.M) {
	log.SetOutput(ioutil.Discard)
	os.Exit(m.Run())
}

func TestMockModuleVersioning(t *testing.T) {
	modSetMap := ModuleSetMap{
		"mod-set-1": ModuleSet{
			Version: "v1.2.3-RC1+meta",
			Modules: []ModulePath{
				"go.opentelemetry.io/test/test1",
				"go.opentelemetry.io/test/test2",
			},
		},
		"mod-set-2": ModuleSet{
			Version: "v0.1.0",
			Modules: []ModulePath{
				"go.opentelemetry.io/test3",
			},
		},
	}

	modPathMap := ModulePathMap{
		"go.opentelemetry.io/test/test1": "root/path/to/mod/test/test1/go.mod",
		"go.opentelemetry.io/test/test2": "root/path/to/mod/test/test2/go.mod",
		"go.opentelemetry.io/test3":      "root/test3/go.mod",
	}

	expected := ModuleVersioning{
		ModSetMap:  modSetMap,
		ModPathMap: modPathMap,
		ModInfoMap: ModuleInfoMap{
			"go.opentelemetry.io/test/test1": ModuleInfo{
				ModuleSetName: "mod-set-1",
				Version:       "v1.2.3-RC1+meta",
			},
			"go.opentelemetry.io/test/test2": ModuleInfo{
				ModuleSetName: "mod-set-1",
				Version:       "v1.2.3-RC1+meta",
			},
			"go.opentelemetry.io/test3": ModuleInfo{
				ModuleSetName: "mod-set-2",
				Version:       "v0.1.0",
			},
		},
	}

	actual, err := MockModuleVersioning(modSetMap, modPathMap)

	require.NoError(t, err)
	assert.Equal(t, expected, actual)
}

func TestNewModuleVersioning(t *testing.T) {
	repoRoot1 := "./test_data/new_module_versioning/"
	repoRoot2 := "./test_data/new_module_versioning/"

	testCases := []struct {
		name                  string
		versioningFilename    string
		repoRoot              string
		shouldError           bool
		expectedModuleSetMap  ModuleSetMap
		expectedModulePathMap ModulePathMap
		expectedModuleInfoMap ModuleInfoMap
	}{
		{
			name:               "valid versioning",
			versioningFilename: "./test_data/new_module_versioning/versions_valid.yaml",
			repoRoot:           repoRoot1,
			shouldError:        false,
			expectedModuleSetMap: ModuleSetMap{
				"mod-set-1": ModuleSet{
					Version: "v1.2.3-RC1+meta",
					Modules: []ModulePath{
						"go.opentelemetry.io/test/test1",
						"go.opentelemetry.io/test/test2",
					},
				},
				"mod-set-2": ModuleSet{
					Version: "v0.1.0",
					Modules: []ModulePath{
						"go.opentelemetry.io/test3",
					},
				},
			},
			expectedModulePathMap: ModulePathMap{
				"go.opentelemetry.io/test/test1": ModuleFilePath(filepath.Join(repoRoot1, "test", "test1", "go.mod")),
				"go.opentelemetry.io/test3":      ModuleFilePath(filepath.Join(repoRoot1, "test", "go.mod")),
				"go.opentelemetry.io/testroot":   ModuleFilePath(filepath.Join(repoRoot1, "go.mod")),
			},
			expectedModuleInfoMap: ModuleInfoMap{
				"go.opentelemetry.io/test/test1": ModuleInfo{
					ModuleSetName: "mod-set-1",
					Version:       "v1.2.3-RC1+meta",
				},
				"go.opentelemetry.io/test/test2": ModuleInfo{
					ModuleSetName: "mod-set-1",
					Version:       "v1.2.3-RC1+meta",
				},
				"go.opentelemetry.io/test3": ModuleInfo{
					ModuleSetName: "mod-set-2",
					Version:       "v0.1.0",
				},
			},
		},
		{
			name:                  "invalid version file syntax",
			versioningFilename:    "./test_data/new_module_versioning/versions_invalid_syntax.yaml",
			repoRoot:              repoRoot2,
			shouldError:           true,
			expectedModuleSetMap:  nil,
			expectedModulePathMap: nil,
			expectedModuleInfoMap: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual, err := NewModuleVersioning(tc.versioningFilename, tc.repoRoot)

			if tc.shouldError {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
			}
			assert.IsType(t, ModuleVersioning{}, actual)
			assert.Equal(t, tc.expectedModuleSetMap, actual.ModSetMap)
			assert.Equal(t, tc.expectedModulePathMap, actual.ModPathMap)
			assert.Equal(t, tc.expectedModuleInfoMap, actual.ModInfoMap)
		})
	}
}

func TestReadVersioningFile(t *testing.T) {
	testCases := []struct {
		name                    string
		versioningFilename      string
		ShouldError             bool
		ExpectedModuleSets      ModuleSetMap
		ExpectedExcludedModules []ModulePath
	}{
		{
			name:               "valid versioning",
			versioningFilename: "./test_data/read_versioning_filename/versions_valid.yaml",
			ShouldError:        false,
			ExpectedModuleSets: ModuleSetMap{
				"mod-set-1": ModuleSet{
					Version: "v1.2.3-RC1+meta",
					Modules: []ModulePath{
						"go.opentelemetry.io/test/test1",
						"go.opentelemetry.io/test/test2",
					},
				},
				"mod-set-2": ModuleSet{
					Version: "v0.1.0",
					Modules: []ModulePath{
						"go.opentelemetry.io/test3",
					},
				},
			},
			ExpectedExcludedModules: []ModulePath{
				"go.opentelemetry.io/excluded1",
			},
		},
		{
			name:                    "invalid version file syntax",
			versioningFilename:      "./test_data/read_versioning_filename/versions_invalid_syntax.yaml",
			ShouldError:             true,
			ExpectedModuleSets:      nil,
			ExpectedExcludedModules: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual, err := readVersioningFile(tc.versioningFilename)

			if tc.ShouldError {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			assert.IsType(t, versionConfig{}, actual)
			assert.Equal(t, tc.ExpectedModuleSets, actual.ModuleSets)
			assert.Equal(t, tc.ExpectedExcludedModules, actual.ExcludedModules)
		})
	}
}

func TestBuildModuleSetsMap(t *testing.T) {
	vCfg := versionConfig{
		ModuleSets: ModuleSetMap{
			"mod-set-1": ModuleSet{
				Version: "v1.2.3-RC1+meta",
				Modules: []ModulePath{
					"go.opentelemetry.io/test/test1",
					"go.opentelemetry.io/test/test2",
				},
			},
			"mod-set-2": ModuleSet{
				Version: "v0.1.0",
				Modules: []ModulePath{
					"go.opentelemetry.io/test3",
				},
			},
		},
		ExcludedModules: []ModulePath{
			"go.opentelemetry.io/excluded1",
		},
	}

	expected := ModuleSetMap{
		"mod-set-1": ModuleSet{
			Version: "v1.2.3-RC1+meta",
			Modules: []ModulePath{
				"go.opentelemetry.io/test/test1",
				"go.opentelemetry.io/test/test2",
			},
		},
		"mod-set-2": ModuleSet{
			Version: "v0.1.0",
			Modules: []ModulePath{
				"go.opentelemetry.io/test3",
			},
		},
	}

	actual, err := vCfg.buildModuleSetsMap()

	require.NoError(t, err)
	assert.Equal(t, expected, actual)
}

func TestBuildModuleMap(t *testing.T) {
	testCases := []struct {
		name        string
		vCfg        versionConfig
		shouldError bool
		expected    ModuleInfoMap
	}{
		{
			name: "valid",
			vCfg: versionConfig{
				ModuleSets: ModuleSetMap{
					"mod-set-1": ModuleSet{
						Version: "v1.2.3-RC1+meta",
						Modules: []ModulePath{
							"go.opentelemetry.io/test/test1",
							"go.opentelemetry.io/test/test2",
						},
					},
					"mod-set-2": ModuleSet{
						Version: "v0.1.0",
						Modules: []ModulePath{
							"go.opentelemetry.io/test3",
						},
					},
				},
				ExcludedModules: []ModulePath{
					"go.opentelemetry.io/excluded1",
				},
			},
			shouldError: false,
			expected: ModuleInfoMap{
				"go.opentelemetry.io/test/test1": ModuleInfo{
					ModuleSetName: "mod-set-1",
					Version:       "v1.2.3-RC1+meta",
				},
				"go.opentelemetry.io/test/test2": ModuleInfo{
					ModuleSetName: "mod-set-1",
					Version:       "v1.2.3-RC1+meta",
				},
				"go.opentelemetry.io/test3": ModuleInfo{
					ModuleSetName: "mod-set-2",
					Version:       "v0.1.0",
				},
			},
		},
		{
			name: "module duplicated",
			vCfg: versionConfig{
				ModuleSets: ModuleSetMap{
					"mod-set-1": ModuleSet{
						Version: "v1.2.3-RC1+meta",
						Modules: []ModulePath{
							"go.opentelemetry.io/duplicate",
						},
					},
					"mod-set-2": ModuleSet{
						Version: "v0.1.0",
						Modules: []ModulePath{
							"go.opentelemetry.io/duplicate",
						},
					},
				},
				ExcludedModules: []ModulePath{
					"go.opentelemetry.io/excluded1",
				},
			},
			shouldError: true,
			expected:    nil,
		},
		{
			name: "module listed in set and excluded",
			vCfg: versionConfig{
				ModuleSets: ModuleSetMap{
					"mod-set-1": ModuleSet{
						Version: "v1.2.3-RC1+meta",
						Modules: []ModulePath{
							"go.opentelemetry.io/excluded",
						},
					},
				},
				ExcludedModules: []ModulePath{
					"go.opentelemetry.io/excluded",
				},
			},
			shouldError: true,
			expected:    nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual, err := tc.vCfg.buildModuleMap()

			if tc.shouldError {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
			}
			assert.Equal(t, tc.expected, actual)
		})
	}
}

func TestShouldExcludeModule(t *testing.T) {
	vCfg := versionConfig{
		ModuleSets: ModuleSetMap{
			"mod-set-1": ModuleSet{
				Version: "v1.2.3-RC1+meta",
				Modules: []ModulePath{
					"go.opentelemetry.io/test/test1",
					"go.opentelemetry.io/test/test2",
				},
			},
			"mod-set-2": ModuleSet{
				Version: "v0.1.0",
				Modules: []ModulePath{
					"go.opentelemetry.io/test3",
				},
			},
		},
		ExcludedModules: []ModulePath{
			"go.opentelemetry.io/excluded1",
		},
	}

	testCases := []struct {
		ModPath  ModulePath
		Expected bool
	}{
		{ModPath: "go.opentelemetry.io/test/test1", Expected: false},
		{ModPath: "go.opentelemetry.io/test/test2", Expected: false},
		{ModPath: "go.opentelemetry.io/test3", Expected: false},
		{ModPath: "go.opentelemetry.io/excluded1", Expected: true},
		{ModPath: "go.opentelemetry.io/doesnotexist", Expected: false},
	}

	for _, tc := range testCases {
		actual := vCfg.shouldExcludeModule(tc.ModPath)

		assert.Equal(t, actual, tc.Expected)
	}
}

func TestGetExcludedModules(t *testing.T) {
	vCfg := versionConfig{
		ModuleSets: ModuleSetMap{
			"mod-set-1": ModuleSet{
				Version: "v1.2.3-RC1+meta",
				Modules: []ModulePath{
					"go.opentelemetry.io/test/test1",
					"go.opentelemetry.io/test/test2",
				},
			},
			"mod-set-2": ModuleSet{
				Version: "v0.1.0",
				Modules: []ModulePath{
					"go.opentelemetry.io/test3",
				},
			},
		},
		ExcludedModules: []ModulePath{
			"go.opentelemetry.io/excluded1",
		},
	}

	expected := excludedModulesSet{
		"go.opentelemetry.io/excluded1": struct{}{},
	}

	actual := vCfg.getExcludedModules()

	assert.Equal(t, expected, actual)
}

func TestBuildModulePathMap(t *testing.T) {
	vCfg := versionConfig{
		ModuleSets: ModuleSetMap{
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
					"go.opentelemetry.io/testroot",
				},
			},
		},
		ExcludedModules: []ModulePath{
			"go.opentelemetry.io/test/testexcluded",
		},
	}

	repoRoot := "./test_data/build_module_path_map"

	expected := ModulePathMap{
		"go.opentelemetry.io/test/test1": ModuleFilePath(filepath.Join(repoRoot, "test", "test1", "go.mod")),
		"go.opentelemetry.io/test3":      ModuleFilePath(filepath.Join(repoRoot, "test", "go.mod")),
		"go.opentelemetry.io/testroot":   ModuleFilePath(filepath.Join(repoRoot, "go.mod")),
	}

	makeTree(t)

	actual, err := vCfg.BuildModulePathMap(repoRoot)

	require.NoError(t, err)
	assert.Equal(t, expected, actual)
}

func makeTree(t *testing.T) {

}

func TestCombineModuleTagNamesAndVersion(t *testing.T) {
	modTagNames := []ModuleTagName{
		"tag1",
		"tag2",
		"another/tag3",
		repoRootTag,
	}

	version := "v1.2.3-RC1+meta-RC1"

	expected := []string{
		"tag1/v1.2.3-RC1+meta-RC1",
		"tag2/v1.2.3-RC1+meta-RC1",
		"another/tag3/v1.2.3-RC1+meta-RC1",
		"v1.2.3-RC1+meta-RC1",
	}

	actual := combineModuleTagNamesAndVersion(modTagNames, version)

	assert.Equal(t, expected, actual)
}

func TestModulePathsToTagNames(t *testing.T) {
	modPaths := []ModulePath{
		"go.opentelemetry.io/test/test1",
		"go.opentelemetry.io/test/test2",
		"go.opentelemetry.io/test3",
		"go.opentelemetry.io/root",
	}

	modPathMap := ModulePathMap{
		"go.opentelemetry.io/test/test1": "root/path/to/mod/test/test1/go.mod",
		"go.opentelemetry.io/test/test2": "root/path/to/mod/test/test2/go.mod",
		"go.opentelemetry.io/test3":      "root/test3/go.mod",
		"go.opentelemetry.io/root":       "root/go.mod",
		"go.opentelemetry.io/not-used":   "path/to/mod/not-used/go.mod",
	}

	repoRoot := "root"

	expected := []ModuleTagName{
		"path/to/mod/test/test1",
		"path/to/mod/test/test2",
		"test3",
		repoRootTag,
	}

	actual, err := modulePathsToTagNames(modPaths, modPathMap, repoRoot)

	require.NoError(t, err)
	assert.Equal(t, expected, actual)
}

func TestModulePathsToFilePaths(t *testing.T) {
	testCases := []struct {
		name        string
		modPaths    []ModulePath
		modPathMap  ModulePathMap
		shouldError bool
		expected    []ModuleFilePath
	}{
		{
			name: "valid",
			modPaths: []ModulePath{
				"go.opentelemetry.io/test/test1",
				"go.opentelemetry.io/test/test2",
				"go.opentelemetry.io/test3",
				"go.opentelemetry.io/root",
			},
			modPathMap: ModulePathMap{
				"go.opentelemetry.io/test/test1": "root/path/to/mod/test/test1/go.mod",
				"go.opentelemetry.io/test/test2": "root/path/to/mod/test/test2/go.mod",
				"go.opentelemetry.io/test3":      "root/test3/go.mod",
				"go.opentelemetry.io/root":       "root/go.mod",
				"go.opentelemetry.io/not-used":   "path/to/mod/not-used/go.mod",
			},
			shouldError: false,
			expected: []ModuleFilePath{
				"root/path/to/mod/test/test1/go.mod",
				"root/path/to/mod/test/test2/go.mod",
				"root/test3/go.mod",
				"root/go.mod",
			},
		},
		{
			name: "module not in map",
			modPaths: []ModulePath{
				"go.opentelemetry.io/in_map",
				"go.opentelemetry.io/not_in_map",
			},
			modPathMap: ModulePathMap{
				"go.opentelemetry.io/in_map": "root/path/go.mod",
			},
			shouldError: true,
			expected:    []ModuleFilePath{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual, err := modulePathsToFilePaths(tc.modPaths, tc.modPathMap)

			if tc.shouldError {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
			}
			assert.Equal(t, tc.expected, actual)
		})
	}
}

func TestModuleFilePathToTagName(t *testing.T) {
	repoRoot := "root"

	testCases := []struct {
		name        string
		ModFilePath ModuleFilePath
		ShouldError bool
		Expected    ModuleTagName
	}{
		{
			name:        "go mod file in inner dir",
			ModFilePath: "root/path/to/mod/test/test1/go.mod",
			ShouldError: false,
			Expected:    ModuleTagName("path/to/mod/test/test1"),
		},
		{
			name:        "go mod file in root",
			ModFilePath: "root/go.mod",
			ShouldError: false,
			Expected:    repoRootTag,
		},
		{
			name:        "no go mod in path",
			ModFilePath: "no/go/mod/in/path",
			ShouldError: true,
			Expected:    "",
		},
		{
			name:        "go mod not contained within root",
			ModFilePath: "not/in/root/go.mod",
			ShouldError: true,
			Expected:    "",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual, err := moduleFilePathToTagName(tc.ModFilePath, repoRoot)

			if tc.ShouldError {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.Expected, actual)
			}
		})
	}
}

func TestModuleFilePathsToTagNames(t *testing.T) {
	testCases := []struct {
		name         string
		modFilePaths []ModuleFilePath
		repoRoot     string
		shouldError  bool
		expected     []ModuleTagName
	}{
		{
			name: "valid",
			modFilePaths: []ModuleFilePath{
				"root/path/to/mod/test/test1/go.mod",
				"root/path/to/mod/test/test2/go.mod",
				"root/test3/go.mod",
				"root/go.mod",
			},
			repoRoot:    "root",
			shouldError: false,
			expected: []ModuleTagName{
				"path/to/mod/test/test1",
				"path/to/mod/test/test2",
				"test3",
				repoRootTag,
			},
		},
		{
			name: "no go mod in path",
			modFilePaths: []ModuleFilePath{
				"no/go/mod/in/path",
			},
			repoRoot:    "root",
			shouldError: true,
			expected:    nil,
		},
		{
			name: "go mod not contained in root",
			modFilePaths: []ModuleFilePath{
				"not/in/root/go.mod",
			},
			repoRoot:    "root",
			shouldError: true,
			expected:    nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual, err := moduleFilePathsToTagNames(tc.modFilePaths, tc.repoRoot)

			if tc.shouldError {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
			}
			assert.Equal(t, tc.expected, actual)
		})
	}
}

func TestIsStableVersion(t *testing.T) {
	testCases := []struct {
		Version  string
		Expected bool
	}{
		{Version: "v1.0.0", Expected: true},
		{Version: "v1.2.3", Expected: true},
		{Version: "v1.0.0-RC1", Expected: true},
		{Version: "v1.0.0-RC2+MetaData", Expected: true},
		{Version: "v10.10.10", Expected: true},
		{Version: "v0.0.0", Expected: false},
		{Version: "v0.1.2", Expected: false},
		{Version: "v0.20.0", Expected: false},
		{Version: "v0.0.0-RC1", Expected: false},
		{Version: "not-valid-semver", Expected: false},
	}

	for _, tc := range testCases {
		actual := IsStableVersion(tc.Version)

		assert.Equal(t, tc.Expected, actual)
	}
}

func TestChangeToRepoRoot(t *testing.T) {
	expected, _ := filepath.Abs("../../../")

	actual, err := ChangeToRepoRoot()

	require.NoError(t, err)
	assert.Equal(t, expected, actual)

	newDir, err := os.Getwd()
	if err != nil {
		t.Logf("could not get current working directory: %v", err)
	}
	assert.Equal(t, expected, newDir)
}
