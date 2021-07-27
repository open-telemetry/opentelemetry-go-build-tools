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
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testDataDir = "./test_data"
)

// TestMain performs setup for the tests and suppress printing logs.
func TestMain(m *testing.M) {
	log.SetOutput(ioutil.Discard)
	os.Exit(m.Run())
}

func TestNewModuleVersioning(t *testing.T) {
	tmpRootDir, err := os.MkdirTemp(testDataDir, "NewModuleVersioning")
	if err != nil {
		t.Fatal("creating temp dir:", err)
	}

	defer os.RemoveAll(tmpRootDir)

	modFiles := map[ModuleFilePath][]byte{
		ModuleFilePath(filepath.Join(tmpRootDir, "test", "test1", "go.mod")): []byte("module \"go.opentelemetry.io/test/test1\"\n\ngo 1.16\n\n" +
			"require (\n\t\"go.opentelemetry.io/testroot/v2\" v2.0.0\n)\n"),
		ModuleFilePath(filepath.Join(tmpRootDir, "test", "go.mod")):          []byte("module go.opentelemetry.io/test3\n\ngo 1.16\n"),
		ModuleFilePath(filepath.Join(tmpRootDir, "go.mod")):                  []byte("module go.opentelemetry.io/testroot/v2\n\ngo 1.16\n"),
		ModuleFilePath(filepath.Join(tmpRootDir, "test", "test2", "go.mod")): []byte("module \"go.opentelemetry.io/test/testexcluded\"\n\ngo 1.16\n"),
	}

	if err := writeGoModFiles(modFiles); err != nil {
		t.Fatal("could not create go mod file tree", err)
	}

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
			versioningFilename: filepath.Join(testDataDir, "new_module_versioning/versions_valid.yaml"),
			repoRoot:           tmpRootDir,
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
				"go.opentelemetry.io/test/test1":  ModuleFilePath(filepath.Join(tmpRootDir, "test", "test1", "go.mod")),
				"go.opentelemetry.io/test3":       ModuleFilePath(filepath.Join(tmpRootDir, "test", "go.mod")),
				"go.opentelemetry.io/testroot/v2": ModuleFilePath(filepath.Join(tmpRootDir, "go.mod")),
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
			versioningFilename:    filepath.Join(testDataDir, "new_module_versioning/versions_invalid_syntax.yaml"),
			repoRoot:              tmpRootDir,
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
			versioningFilename: filepath.Join(testDataDir, "read_versioning_filename/versions_valid.yaml"),
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
			versioningFilename:      filepath.Join(testDataDir, "read_versioning_filename/versions_invalid_syntax.yaml"),
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

	tmpRootDir, err := os.MkdirTemp(testDataDir, "BuildModulePathMap")
	if err != nil {
		t.Fatal("creating temp dir:", err)
	}

	defer os.RemoveAll(tmpRootDir)

	modFiles := map[ModuleFilePath][]byte{
		ModuleFilePath(filepath.Join(tmpRootDir, "test", "test1", "go.mod")): []byte("module \"go.opentelemetry.io/test/test1\"\n\ngo 1.16\n\n" +
			"require (\n\t\"go.opentelemetry.io/testroot/v2\" v2.0.0\n)\n"),
		ModuleFilePath(filepath.Join(tmpRootDir, "test", "go.mod")):          []byte("module go.opentelemetry.io/test3\n\ngo 1.16\n"),
		ModuleFilePath(filepath.Join(tmpRootDir, "go.mod")):                  []byte("module go.opentelemetry.io/testroot/v2\n\ngo 1.16\n"),
		ModuleFilePath(filepath.Join(tmpRootDir, "test", "test2", "go.mod")): []byte("module \"go.opentelemetry.io/test/testexcluded\"\n\ngo 1.16\n"),
	}

	if err := writeGoModFiles(modFiles); err != nil {
		t.Fatal("could not create go mod file tree", err)
	}

	expected := ModulePathMap{
		"go.opentelemetry.io/test/test1":  ModuleFilePath(filepath.Join(tmpRootDir, "test", "test1", "go.mod")),
		"go.opentelemetry.io/test3":       ModuleFilePath(filepath.Join(tmpRootDir, "test", "go.mod")),
		"go.opentelemetry.io/testroot/v2": ModuleFilePath(filepath.Join(tmpRootDir, "go.mod")),
	}

	actual, err := vCfg.BuildModulePathMap(tmpRootDir)

	require.NoError(t, err)
	assert.Equal(t, expected, actual)
}

// writeGoModFiles is a helper function to dynamically write go.mod files used for testing.
// This func is duplicated from the commontest package to avoid a cyclic dependency.
func writeGoModFiles(modFiles map[ModuleFilePath][]byte) error {
	perm := os.FileMode(0700)

	for modFilePath, file := range modFiles {
		path := filepath.Dir(string(modFilePath))
		os.MkdirAll(path, perm)

		if err := ioutil.WriteFile(string(modFilePath), file, perm); err != nil {
			return fmt.Errorf("could not write temporary mod file %v", err)
		}
	}

	return nil
}

func TestWriteGoModFiles(t *testing.T) {
	tmpRootDir, err := os.MkdirTemp(testDataDir, "NewModuleVersioning")
	if err != nil {
		t.Fatal("creating temp dir:", err)
	}

	defer os.RemoveAll(tmpRootDir)

	modFiles := map[ModuleFilePath][]byte{
		ModuleFilePath(filepath.Join(tmpRootDir, "test", "test1", "go.mod")): []byte("module \"go.opentelemetry.io/test/test1\"\n\ngo 1.16\n\n" +
			"require (\n\t\"go.opentelemetry.io/testroot/v2\" v2.0.0\n)\n"),
		ModuleFilePath(filepath.Join(tmpRootDir, "test", "go.mod")):          []byte("module go.opentelemetry.io/test3\n\ngo 1.16\n"),
		ModuleFilePath(filepath.Join(tmpRootDir, "go.mod")):                  []byte("module go.opentelemetry.io/testroot/v2\n\ngo 1.16\n"),
		ModuleFilePath(filepath.Join(tmpRootDir, "test", "test2", "go.mod")): []byte("module \"go.opentelemetry.io/test/testexcluded\"\n\ngo 1.16\n"),
	}

	if err := writeGoModFiles(modFiles); err != nil {
		t.Fatal("could not create go mod file tree", err)
	}

	// check all mod files have been written correctly
	for modPath, expectedModFile := range modFiles {
		actual, err := ioutil.ReadFile(string(modPath))

		require.NoError(t, err)
		assert.Equal(t, expectedModFile, actual)
	}
}
