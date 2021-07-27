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

package verify

import (
	"bytes"
	"go.opentelemetry.io/build-tools/releaser/internal/common/commontest"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.opentelemetry.io/build-tools/releaser/internal/common"
)

const (
	testDataDir = "./test_data"
)

// TestMain performs setup for the tests and suppress printing logs.
func TestMain(m *testing.M) {
	log.SetOutput(ioutil.Discard)
	os.Exit(m.Run())
}

func captureOutput(f func()) string {
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer func() {
		log.SetOutput(ioutil.Discard)
	}()

	f()
	return buf.String()
}

// MockVerification creates a verification struct for testing purposes.
func MockVerification(modSetMap common.ModuleSetMap, modPathMap common.ModulePathMap, dependencies dependencyMap) verification {
	modVersioning, err := commontest.MockModuleVersioning(modSetMap, modPathMap)
	if err != nil {
		log.Printf("error getting MockModuleVersioning: %v", err)
		return verification{}
	}

	return verification{
		ModuleVersioning: modVersioning,
		dependencies:     dependencies,
	}
}

// Positive-only test
func TestMockVerification(t *testing.T) {
	modSetMap := common.ModuleSetMap{
		"mod-set-1": common.ModuleSet{
			Version: "v1.2.3-RC1+meta",
			Modules: []common.ModulePath{
				"go.opentelemetry.io/test/test1",
				"go.opentelemetry.io/test/test2",
			},
		},
		"mod-set-2": common.ModuleSet{
			Version: "v0.1.0",
			Modules: []common.ModulePath{
				"go.opentelemetry.io/test3",
			},
		},
	}

	modPathMap := common.ModulePathMap{
		"go.opentelemetry.io/test/test1": "root/path/to/mod/test/test1/go.mod",
		"go.opentelemetry.io/test/test2": "root/path/to/mod/test/test2/go.mod",
		"go.opentelemetry.io/test3":      "root/test3/go.mod",
	}

	depMap := dependencyMap{
		"go.opentelemetry.io/test/test1": []common.ModulePath{"go.opentelemetry.io/test/test2"},
		"go.opentelemetry.io/test3":      []common.ModulePath{"go.opentelemetry.io/test/test1"},
	}

	expected := verification{
		ModuleVersioning: common.ModuleVersioning{
			ModSetMap:  modSetMap,
			ModPathMap: modPathMap,
			ModInfoMap: common.ModuleInfoMap{
				"go.opentelemetry.io/test/test1": common.ModuleInfo{
					ModuleSetName: "mod-set-1",
					Version:       "v1.2.3-RC1+meta",
				},
				"go.opentelemetry.io/test/test2": common.ModuleInfo{
					ModuleSetName: "mod-set-1",
					Version:       "v1.2.3-RC1+meta",
				},
				"go.opentelemetry.io/test3": common.ModuleInfo{
					ModuleSetName: "mod-set-2",
					Version:       "v0.1.0",
				},
			},
		},
		dependencies: depMap,
	}

	actual := MockVerification(modSetMap, modPathMap, depMap)

	assert.IsType(t, verification{}, actual)
	assert.Equal(t, expected, actual)
}

func TestNewVerification(t *testing.T) {
	tmpRootDir, err := os.MkdirTemp(testDataDir, "NewVerification")
	if err != nil {
		t.Fatal("creating temp dir:", err)
	}

	defer os.RemoveAll(tmpRootDir)

	modFiles := map[common.ModuleFilePath][]byte{
		common.ModuleFilePath(filepath.Join(tmpRootDir, "test", "test1", "go.mod")): []byte("module \"go.opentelemetry.io/test/test1\"\n\ngo 1.16\n\n" +
			"require (\n\t\"go.opentelemetry.io/testroot/v2\" v2.0.0\n)\n"),
		common.ModuleFilePath(filepath.Join(tmpRootDir, "test", "go.mod")):          []byte("module go.opentelemetry.io/test3\n\ngo 1.16\n"),
		common.ModuleFilePath(filepath.Join(tmpRootDir, "go.mod")):                  []byte("module go.opentelemetry.io/testroot/v2\n\ngo 1.16\n"),
		common.ModuleFilePath(filepath.Join(tmpRootDir, "test", "test2", "go.mod")): []byte("module \"go.opentelemetry.io/test/testexcluded\"\n\ngo 1.16\n"),
	}

	if err := commontest.WriteGoModFiles(modFiles); err != nil {
		t.Fatal("could not create go mod file tree", err)
	}

	testCases := []struct {
		name                  string
		versioningFilename    string
		repoRoot              string
		shouldError           bool
		expectedModuleSetMap  common.ModuleSetMap
		expectedModulePathMap common.ModulePathMap
		expectedModuleInfoMap common.ModuleInfoMap
	}{
		{
			name:               "valid versioning",
			versioningFilename: filepath.Join(testDataDir, "new_verification/versions_valid.yaml"),
			repoRoot:           tmpRootDir,
			shouldError:        false,
			expectedModuleSetMap: common.ModuleSetMap{
				"mod-set-1": common.ModuleSet{
					Version: "v1.2.3-RC1+meta",
					Modules: []common.ModulePath{
						"go.opentelemetry.io/test/test1",
						"go.opentelemetry.io/test/test2",
					},
				},
				"mod-set-2": common.ModuleSet{
					Version: "v0.1.0",
					Modules: []common.ModulePath{
						"go.opentelemetry.io/test3",
					},
				},
			},
			expectedModulePathMap: common.ModulePathMap{
				"go.opentelemetry.io/test/test1":  common.ModuleFilePath(filepath.Join(tmpRootDir, "test", "test1", "go.mod")),
				"go.opentelemetry.io/test3":       common.ModuleFilePath(filepath.Join(tmpRootDir, "test", "go.mod")),
				"go.opentelemetry.io/testroot/v2": common.ModuleFilePath(filepath.Join(tmpRootDir, "go.mod")),
			},
			expectedModuleInfoMap: common.ModuleInfoMap{
				"go.opentelemetry.io/test/test1": common.ModuleInfo{
					ModuleSetName: "mod-set-1",
					Version:       "v1.2.3-RC1+meta",
				},
				"go.opentelemetry.io/test/test2": common.ModuleInfo{
					ModuleSetName: "mod-set-1",
					Version:       "v1.2.3-RC1+meta",
				},
				"go.opentelemetry.io/test3": common.ModuleInfo{
					ModuleSetName: "mod-set-2",
					Version:       "v0.1.0",
				},
			},
		},
		{
			name:                  "invalid version file syntax",
			versioningFilename:    filepath.Join(testDataDir, "new_verification/versions_invalid_syntax.yaml"),
			repoRoot:              tmpRootDir,
			shouldError:           true,
			expectedModuleSetMap:  nil,
			expectedModulePathMap: nil,
			expectedModuleInfoMap: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual, err := newVerification(tc.versioningFilename, tc.repoRoot)

			if tc.shouldError {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
			}
			assert.IsType(t, verification{}, actual)
			assert.Equal(t, tc.expectedModuleSetMap, actual.ModuleVersioning.ModSetMap)
			assert.Equal(t, tc.expectedModulePathMap, actual.ModuleVersioning.ModPathMap)
			assert.Equal(t, tc.expectedModuleInfoMap, actual.ModuleVersioning.ModInfoMap)
		})
	}
}

// Positive-only test
func TestGetDependencies(t *testing.T) {
	tmpRootDir, err := os.MkdirTemp(testDataDir, "GetDependencies")
	if err != nil {
		t.Fatal("creating temp dir:", err)
	}

	defer os.RemoveAll(tmpRootDir)

	modFiles := map[common.ModuleFilePath][]byte{
		common.ModuleFilePath(filepath.Join(tmpRootDir, "test", "test1", "go.mod")): []byte("module go.opentelemetry.io/build-tools/releaser/internal/verify/test/test1\n\n" +
			"go 1.16\n\n" +
			"require (\n\t" +
			"go.opentelemetry.io/build-tools/releaser/internal/verify/test/test2 v1.2.3-RC1+meta\n\t" +
			"go.opentelemetry.io/build-tools/releaser/internal/verify/test3 v0.1.0\n" +
			")"),
		common.ModuleFilePath(filepath.Join(tmpRootDir, "test", "go.mod")): []byte("module go.opentelemetry.io/build-tools/releaser/internal/verify/test3\n\n" +
			"go 1.16\n\n" +
			"require (\n\t" +
			"go.opentelemetry.io/build-tools/releaser/internal/verify/test/test1 v1.2.3-RC1+meta\n\t" +
			"go.opentelemetry.io/build-tools/releaser/internal/verify/test/test2 v1.2.3-RC1+meta\n\n" +
			")"),
		common.ModuleFilePath(filepath.Join(tmpRootDir, "go.mod")): []byte("module go.opentelemetry.io/build-tools/releaser/internal/verify/testroot\n\n" +
			"go 1.16\n\n" +
			"require (\n\t" +
			"go.opentelemetry.io/build-tools/releaser/internal/verify/test/test1 v1.2.3-RC1+meta\n\t" +
			"go.opentelemetry.io/build-tools/releaser/internal/verify/test/test2 v1.2.3-RC1+meta\n\t" +
			"go.opentelemetry.io/build-tools/releaser/internal/verify/test3 v0.1.0\n" +
			")"),
		common.ModuleFilePath(filepath.Join(tmpRootDir, "test", "test2", "go.mod")): []byte("module go.opentelemetry.io/build-tools/releaser/internal/verify/test/test2\n\n" +
			"go 1.16\n\n" +
			"require (\n\t" +
			"go.opentelemetry.io/build-tools/releaser/internal/verify/test/test1 v1.2.3-RC1+meta\n\t" +
			"go.opentelemetry.io/build-tools/releaser/internal/verify/test3 v0.1.0\n\t" +
			"go.opentelemetry.io/build-tools/releaser/internal/verify/testroot v0.2.0\n" +
			")"),
	}

	if err := commontest.WriteGoModFiles(modFiles); err != nil {
		t.Fatal("could not create go mod file tree", err)
	}
	modVersioning, _ := commontest.MockModuleVersioning(
		common.ModuleSetMap{
			"mod-set-1": common.ModuleSet{
				Version: "v1.2.3-RC1+meta",
				Modules: []common.ModulePath{
					"go.opentelemetry.io/build-tools/releaser/internal/verify/test/test1",
					"go.opentelemetry.io/build-tools/releaser/internal/verify/test/test2",
				},
			},
			"mod-set-2": common.ModuleSet{
				Version: "v0.1.0",
				Modules: []common.ModulePath{
					"go.opentelemetry.io/build-tools/releaser/internal/verify/test3",
				},
			},
			"mod-set-3": common.ModuleSet{
				Version: "v0.2.0",
				Modules: []common.ModulePath{
					"go.opentelemetry.io/build-tools/releaser/internal/verify/testroot",
				},
			},
		},
		common.ModulePathMap{
			"go.opentelemetry.io/build-tools/releaser/internal/verify/test/test1": common.ModuleFilePath(filepath.Join(tmpRootDir, "test/test1/go.mod")),
			"go.opentelemetry.io/build-tools/releaser/internal/verify/test/test2": common.ModuleFilePath(filepath.Join(tmpRootDir, "test/test2/go.mod")),
			"go.opentelemetry.io/build-tools/releaser/internal/verify/test3":      common.ModuleFilePath(filepath.Join(tmpRootDir, "test/go.mod")),
			"go.opentelemetry.io/build-tools/releaser/internal/verify/testroot":   common.ModuleFilePath(filepath.Join(tmpRootDir, "go.mod")),
		},
	)

	expected := dependencyMap{
		"go.opentelemetry.io/build-tools/releaser/internal/verify/test/test1": []common.ModulePath{
			"go.opentelemetry.io/build-tools/releaser/internal/verify/test/test2",
			"go.opentelemetry.io/build-tools/releaser/internal/verify/test3",
		},
		"go.opentelemetry.io/build-tools/releaser/internal/verify/test/test2": []common.ModulePath{
			"go.opentelemetry.io/build-tools/releaser/internal/verify/test/test1",
			"go.opentelemetry.io/build-tools/releaser/internal/verify/test3",
			"go.opentelemetry.io/build-tools/releaser/internal/verify/testroot",
		},
		"go.opentelemetry.io/build-tools/releaser/internal/verify/test3": []common.ModulePath{
			"go.opentelemetry.io/build-tools/releaser/internal/verify/test/test1",
			"go.opentelemetry.io/build-tools/releaser/internal/verify/test/test2",
		},
		"go.opentelemetry.io/build-tools/releaser/internal/verify/testroot": []common.ModulePath{
			"go.opentelemetry.io/build-tools/releaser/internal/verify/test/test1",
			"go.opentelemetry.io/build-tools/releaser/internal/verify/test/test2",
			"go.opentelemetry.io/build-tools/releaser/internal/verify/test3",
		},
	}

	actual, err := getDependencies(modVersioning)

	require.NoError(t, err)
	assert.Equal(t, len(expected), len(actual))
	for modPath, expectedDepPaths := range expected {
		actualDepPaths, ok := actual[modPath]
		require.True(t, ok, "modPath %v is not in actual dependencyMap.", modPath)

		assert.ElementsMatch(t, expectedDepPaths, actualDepPaths)
	}
}

func TestVerifyAllModulesInSet(t *testing.T) {
	testCases := []struct {
		name          string
		v             verification
		expectedError error
	}{
		{
			name: "valid",
			v: MockVerification(
				common.ModuleSetMap{
					"mod-set-1": common.ModuleSet{
						Version: "v1.2.3-RC1+meta",
						Modules: []common.ModulePath{
							"go.opentelemetry.io/test/test1",
							"go.opentelemetry.io/test/test2",
						},
					},
					"mod-set-2": common.ModuleSet{
						Version: "v0.1.0",
						Modules: []common.ModulePath{
							"go.opentelemetry.io/test3",
						},
					},
				},
				common.ModulePathMap{
					"go.opentelemetry.io/test/test1": "root/path/to/mod/test/test1/go.mod",
					"go.opentelemetry.io/test/test2": "root/path/to/mod/test/test2/go.mod",
					"go.opentelemetry.io/test3":      "root/test3/go.mod",
				},
				dependencyMap{},
			),
			expectedError: nil,
		},
		{
			name: "module not listed",
			v: MockVerification(
				common.ModuleSetMap{
					"mod-set-1": common.ModuleSet{
						Version: "v1.2.3-RC1+meta",
						Modules: []common.ModulePath{
							"go.opentelemetry.io/test/test1",
						},
					},
					"mod-set-2": common.ModuleSet{
						Version: "v0.1.0",
						Modules: []common.ModulePath{
							"go.opentelemetry.io/test3",
						},
					},
				},
				common.ModulePathMap{
					"go.opentelemetry.io/test/test1": "root/path/to/mod/test/test1/go.mod",
					"go.opentelemetry.io/test/test2": "root/path/to/mod/test/test2/go.mod",
					"go.opentelemetry.io/test3":      "root/test3/go.mod",
				},
				dependencyMap{},
			),
			expectedError: &errModuleNotInSet{
				modPath:     "go.opentelemetry.io/test/test2",
				modFilePath: "root/path/to/mod/test/test2/go.mod",
			},
		},
		{
			name: "module not in repo",
			v: MockVerification(
				common.ModuleSetMap{
					"mod-set-1": common.ModuleSet{
						Version: "v1.2.3-RC1+meta",
						Modules: []common.ModulePath{
							"go.opentelemetry.io/test/test1",
							"go.opentelemetry.io/test/test2",
						},
					},
					"mod-set-2": common.ModuleSet{
						Version: "v0.1.0",
						Modules: []common.ModulePath{
							"go.opentelemetry.io/test3",
						},
					},
				},
				common.ModulePathMap{
					"go.opentelemetry.io/test/test1": "root/path/to/mod/test/test1/go.mod",
					"go.opentelemetry.io/test/test2": "root/path/to/mod/test/test2/go.mod",
				},
				dependencyMap{},
			),
			expectedError: &errModuleNotInRepo{
				modPath:    "go.opentelemetry.io/test3",
				modSetName: "mod-set-2",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual := tc.v.verifyAllModulesInSet()

			assert.Equal(t, tc.expectedError, actual)
		})
	}
}

func TestVerifyVersions(t *testing.T) {
	testCases := []struct {
		name          string
		v             verification
		expectedError error
	}{
		{
			name: "valid",
			v: MockVerification(
				common.ModuleSetMap{
					"mod-set-1": common.ModuleSet{
						Version: "v1.2.3-RC1+meta",
						Modules: []common.ModulePath{
							"go.opentelemetry.io/test/test1",
							"go.opentelemetry.io/test/test2",
						},
					},
					"mod-set-2": common.ModuleSet{
						Version: "v0.1.0",
						Modules: []common.ModulePath{
							"go.opentelemetry.io/test3",
						},
					},
					"mod-set-3": common.ModuleSet{
						Version: "v0.1.0",
						Modules: []common.ModulePath{
							"go.opentelemetry.io/test4",
						},
					},
					"mod-set-4": common.ModuleSet{
						Version: "v2.2.2-RC1+meta",
						Modules: []common.ModulePath{
							"go.opentelemetry.io/test5",
						},
					},
				},
				common.ModulePathMap{
					"go.opentelemetry.io/test/test1": "root/path/to/mod/test/test1/go.mod",
					"go.opentelemetry.io/test/test2": "root/path/to/mod/test/test2/go.mod",
					"go.opentelemetry.io/test3":      "root/test3/go.mod",
					"go.opentelemetry.io/test4":      "root/test4/go.mod",
					"go.opentelemetry.io/test5":      "root/test5/go.mod",
				},
				dependencyMap{},
			),
			expectedError: nil,
		},
		{
			name: "invalid version",
			v: MockVerification(
				common.ModuleSetMap{
					"mod-set-1": common.ModuleSet{
						Version: "invalid-version-v.02.0.",
						Modules: []common.ModulePath{
							"go.opentelemetry.io/test/test1",
							"go.opentelemetry.io/test/test2",
						},
					},
					"mod-set-2": common.ModuleSet{
						Version: "v0.1.0",
						Modules: []common.ModulePath{
							"go.opentelemetry.io/test3",
						},
					},
				},
				common.ModulePathMap{
					"go.opentelemetry.io/test/test1": "root/path/to/mod/test/test1/go.mod",
					"go.opentelemetry.io/test/test2": "root/path/to/mod/test/test2/go.mod",
					"go.opentelemetry.io/test3":      "root/test3/go.mod",
				},
				dependencyMap{},
			),
			expectedError: &errInvalidVersion{
				modSetName:    "mod-set-1",
				modSetVersion: "invalid-version-v.02.0.",
			},
		},
		{
			name: "multiple sets with same major version",
			v: MockVerification(
				common.ModuleSetMap{
					"mod-set-1": common.ModuleSet{
						Version: "v1.2.3-RC1+meta",
						Modules: []common.ModulePath{
							"go.opentelemetry.io/test/test1",
							"go.opentelemetry.io/test/test2",
						},
					},
					"mod-set-2": common.ModuleSet{
						Version: "v0.1.0",
						Modules: []common.ModulePath{
							"go.opentelemetry.io/test3",
						},
					},
					"mod-set-3": common.ModuleSet{
						Version: "v1.1.0-duplicatedmajor",
						Modules: []common.ModulePath{
							"go.opentelemetry.io/test4",
						},
					},
					"mod-set-4": common.ModuleSet{
						Version: "v1.9.0-anotherduplicatedmajor",
						Modules: []common.ModulePath{
							"go.opentelemetry.io/test5",
						},
					},
				},
				common.ModulePathMap{
					"go.opentelemetry.io/test/test1": "root/path/to/mod/test/test1/go.mod",
					"go.opentelemetry.io/test/test2": "root/path/to/mod/test/test2/go.mod",
					"go.opentelemetry.io/test3":      "root/test3/go.mod",
					"go.opentelemetry.io/test4":      "root/test4/go.mod",
					"go.opentelemetry.io/test5":      "root/test5/go.mod",
				},
				dependencyMap{},
			),
			expectedError: &errMultipleSetSameVersion{
				modSetNames:   []string{"mod-set-1", "mod-set-3", "mod-set-4"},
				modSetVersion: "v1",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual := tc.v.verifyVersions()
			if tc.expectedError != nil {
				// Check if expectedError is of type errMultipleSetSameVersion
				if expectedErr, ok := tc.expectedError.(*errMultipleSetSameVersion); ok {
					actualErr, ok := actual.(*errMultipleSetSameVersion)
					assert.True(t, ok)
					assert.IsType(t, expectedErr, actualErr)

					// compare that modSetNames elements match (order should not matter)
					assert.ElementsMatch(t, expectedErr.modSetNames, actualErr.modSetNames)
				}
			} else {
				assert.Equal(t, tc.expectedError, actual)
			}
		})
	}
}

func TestVerifyDependencies(t *testing.T) {
	testCases := []struct {
		name         string
		v            verification
		expectedLogs []string
	}{
		{
			name: "valid",
			v: MockVerification(
				common.ModuleSetMap{
					"mod-set-1": common.ModuleSet{
						Version: "v1.2.3-RC1+meta",
						Modules: []common.ModulePath{
							"go.opentelemetry.io/test/test1",
							"go.opentelemetry.io/test/test2",
						},
					},
					"mod-set-2": common.ModuleSet{
						Version: "v0.1.0",
						Modules: []common.ModulePath{
							"go.opentelemetry.io/test3",
						},
					},
					"mod-set-3": common.ModuleSet{
						Version: "v0.1.0",
						Modules: []common.ModulePath{
							"go.opentelemetry.io/test4",
						},
					},
					"mod-set-4": common.ModuleSet{
						Version: "v2.2.2-RC1+meta",
						Modules: []common.ModulePath{
							"go.opentelemetry.io/test5",
						},
					},
				},
				common.ModulePathMap{
					"go.opentelemetry.io/test/test1": "root/path/to/mod/test/test1/go.mod",
					"go.opentelemetry.io/test/test2": "root/path/to/mod/test/test2/go.mod",
					"go.opentelemetry.io/test3":      "root/test3/go.mod",
					"go.opentelemetry.io/test4":      "root/test4/go.mod",
					"go.opentelemetry.io/test5":      "root/test5/go.mod",
				},
				dependencyMap{
					"go.opentelemetry.io/test/test1": []common.ModulePath{
						"go.opentelemetry.io/test/test2",
						"go.opentelemetry.io/test5",
					},
					"go.opentelemetry.io/test/test2": []common.ModulePath{
						"go.opentelemetry.io/test/test1",
						"go.opentelemetry.io/test5",
					},
					"go.opentelemetry.io/test3": []common.ModulePath{
						"go.opentelemetry.io/test/test1",
						"go.opentelemetry.io/test/test2",
						"go.opentelemetry.io/test4",
						"go.opentelemetry.io/test5",
					},
					"go.opentelemetry.io/test4": []common.ModulePath{
						"go.opentelemetry.io/test3",
						"go.opentelemetry.io/test5",
					},
					"go.opentelemetry.io/test5": []common.ModulePath{
						"go.opentelemetry.io/test/test1",
					},
				},
			),
			expectedLogs: []string{
				"Finished checking all stable modules' dependencies.\n",
			},
		},
		{
			name: "stable depends on unstable",
			v: MockVerification(
				common.ModuleSetMap{
					"mod-set-1": common.ModuleSet{
						Version: "v1.2.3-RC1+meta",
						Modules: []common.ModulePath{
							"go.opentelemetry.io/test/test1",
							"go.opentelemetry.io/test/test2",
						},
					},
					"mod-set-2": common.ModuleSet{
						Version: "v0.1.0",
						Modules: []common.ModulePath{
							"go.opentelemetry.io/test3",
						},
					},
				},
				common.ModulePathMap{
					"go.opentelemetry.io/test/test1": "root/path/to/mod/test/test1/go.mod",
					"go.opentelemetry.io/test/test2": "root/path/to/mod/test/test2/go.mod",
					"go.opentelemetry.io/test3":      "root/test3/go.mod",
				},
				dependencyMap{
					"go.opentelemetry.io/test/test1": []common.ModulePath{
						"go.opentelemetry.io/test2",
						"go.opentelemetry.io/test3",
					},
					"go.opentelemetry.io/test/test2": []common.ModulePath{
						"go.opentelemetry.io/test1",
						"go.opentelemetry.io/test3",
					},
					"go.opentelemetry.io/test/test3": []common.ModulePath{
						"go.opentelemetry.io/test1",
						"go.opentelemetry.io/test2",
					},
				},
			),
			expectedLogs: []string{
				(&errDependency{
					modPath:    "go.opentelemetry.io/test/test1",
					modVersion: "v1.2.3-RC1+meta",
					depPath:    "go.opentelemetry.io/test3",
					depVersion: "v0.1.0",
				}).Error(),
				(&errDependency{
					modPath:    "go.opentelemetry.io/test/test2",
					modVersion: "v1.2.3-RC1+meta",
					depPath:    "go.opentelemetry.io/test3",
					depVersion: "v0.1.0",
				}).Error(),
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual := captureOutput(func() {
				tc.v.verifyDependencies()
			})

			for _, expectedLog := range tc.expectedLogs {
				assert.Contains(t, actual, expectedLog)
			}
		})
	}
}
