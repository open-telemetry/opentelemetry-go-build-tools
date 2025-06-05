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
	"errors"
	"io"
	"log"
	"os"
	"path/filepath"
	"testing"

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

func captureOutput(f func()) string {
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer func() {
		log.SetOutput(io.Discard)
	}()

	f()
	return buf.String()
}

func TestNewVerification(t *testing.T) {
	testName := "new_verification"
	versionYamlDir := filepath.Join(testDataDir, testName)

	tmpRootDir := t.TempDir()
	modFiles := map[string][]byte{
		filepath.Join(tmpRootDir, "test", "test1", "go.mod"): []byte("module \"go.opentelemetry.io/test/test1\"\n\ngo 1.16\n\n" +
			"require (\n\t\"go.opentelemetry.io/testroot/v2\" v2.0.0\n)\n"),
		filepath.Join(tmpRootDir, "test", "go.mod"):             []byte("module go.opentelemetry.io/test2\n\ngo 1.16\n"),
		filepath.Join(tmpRootDir, "go.mod"):                     []byte("module go.opentelemetry.io/testroot/v2\n\ngo 1.16\n"),
		filepath.Join(tmpRootDir, "test", "excluded", "go.mod"): []byte("module \"go.opentelemetry.io/test/testexcluded\"\n\ngo 1.16\n"),
	}

	require.NoError(t, sharedtest.WriteTempFiles(modFiles), "could not create go mod file tree")

	testCases := []struct {
		name                  string
		versioningFilename    string
		repoRoot              string
		shouldError           bool
		expectedModuleSetMap  shared.ModuleSetMap
		expectedModulePathMap shared.ModulePathMap
		expectedModuleInfoMap shared.ModuleInfoMap
	}{
		{
			name:               "valid versioning",
			versioningFilename: filepath.Join(versionYamlDir, "versions_valid.yaml"),
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
				"go.opentelemetry.io/test2": shared.ModuleInfo{
					ModuleSetName: "mod-set-2",
					Version:       "v0.1.0",
				},
				"go.opentelemetry.io/testroot/v2": shared.ModuleInfo{
					ModuleSetName: "mod-set-3",
					Version:       "v2.2.2",
				},
			},
		},
		{
			name:                  "invalid version file syntax",
			versioningFilename:    filepath.Join(versionYamlDir, "versions_invalid_syntax.yaml"),
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
			assert.Equal(t, tc.expectedModuleSetMap, actual.ModSetMap)
			assert.Equal(t, tc.expectedModulePathMap, actual.ModPathMap)
			assert.Equal(t, tc.expectedModuleInfoMap, actual.ModInfoMap)
		})
	}
}

// Positive-only test
func TestGetDependencies(t *testing.T) {
	testName := "get_dependencies"

	versionYamlDir := filepath.Join(testDataDir, testName)

	tmpRootDir := t.TempDir()
	modFiles := map[string][]byte{
		filepath.Join(tmpRootDir, "test", "test1", "go.mod"): []byte("module go.opentelemetry.io/build-tools/multimod/internal/verify/test/test1\n\n" +
			"go 1.16\n\n" +
			"require (\n\t" +
			"go.opentelemetry.io/build-tools/multimod/internal/verify/test/test2 v1.2.3-RC1+meta\n\t" +
			"go.opentelemetry.io/build-tools/multimod/internal/verify/test3 v0.1.0\n" +
			")"),
		filepath.Join(tmpRootDir, "test", "go.mod"): []byte("module go.opentelemetry.io/build-tools/multimod/internal/verify/test3\n\n" +
			"go 1.16\n\n" +
			"require (\n\t" +
			"go.opentelemetry.io/build-tools/multimod/internal/verify/test/test1 v1.2.3-RC1+meta\n\t" +
			"go.opentelemetry.io/build-tools/multimod/internal/verify/test/test2 v1.2.3-RC1+meta\n\n" +
			")"),
		filepath.Join(tmpRootDir, "go.mod"): []byte("module go.opentelemetry.io/build-tools/multimod/internal/verify/testroot\n\n" +
			"go 1.16\n\n" +
			"require (\n\t" +
			"go.opentelemetry.io/build-tools/multimod/internal/verify/test/test1 v1.2.3-RC1+meta\n\t" +
			"go.opentelemetry.io/build-tools/multimod/internal/verify/test/test2 v1.2.3-RC1+meta\n\t" +
			"go.opentelemetry.io/build-tools/multimod/internal/verify/test3 v0.1.0\n" +
			")"),
		filepath.Join(tmpRootDir, "test", "test2", "go.mod"): []byte("module go.opentelemetry.io/build-tools/multimod/internal/verify/test/test2\n\n" +
			"go 1.16\n\n" +
			"require (\n\t" +
			"go.opentelemetry.io/build-tools/multimod/internal/verify/test/test1 v1.2.3-RC1+meta\n\t" +
			"go.opentelemetry.io/build-tools/multimod/internal/verify/test3 v0.1.0\n\t" +
			"go.opentelemetry.io/build-tools/multimod/internal/verify/testroot v0.2.0\n" +
			")"),
	}

	require.NoError(t, sharedtest.WriteTempFiles(modFiles), "could not create go mod file tree")
	v, _ := newVerification(
		filepath.Join(versionYamlDir, "versions_valid.yaml"),
		tmpRootDir,
	)

	expected := dependencyMap{
		"go.opentelemetry.io/build-tools/multimod/internal/verify/test/test1": []shared.ModulePath{
			"go.opentelemetry.io/build-tools/multimod/internal/verify/test/test2",
			"go.opentelemetry.io/build-tools/multimod/internal/verify/test3",
		},
		"go.opentelemetry.io/build-tools/multimod/internal/verify/test/test2": []shared.ModulePath{
			"go.opentelemetry.io/build-tools/multimod/internal/verify/test/test1",
			"go.opentelemetry.io/build-tools/multimod/internal/verify/test3",
			"go.opentelemetry.io/build-tools/multimod/internal/verify/testroot",
		},
		"go.opentelemetry.io/build-tools/multimod/internal/verify/test3": []shared.ModulePath{
			"go.opentelemetry.io/build-tools/multimod/internal/verify/test/test1",
			"go.opentelemetry.io/build-tools/multimod/internal/verify/test/test2",
		},
		"go.opentelemetry.io/build-tools/multimod/internal/verify/testroot": []shared.ModulePath{
			"go.opentelemetry.io/build-tools/multimod/internal/verify/test/test1",
			"go.opentelemetry.io/build-tools/multimod/internal/verify/test/test2",
			"go.opentelemetry.io/build-tools/multimod/internal/verify/test3",
		},
	}

	actual, err := v.getDependencies()

	require.NoError(t, err)
	require.Equal(t, len(expected), len(actual))
	for modPath, expectedDepPaths := range expected {
		actualDepPaths, ok := actual[modPath]
		require.True(t, ok, "modPath %v is not in actual dependencyMap.", modPath)

		assert.ElementsMatch(t, expectedDepPaths, actualDepPaths)
	}
}

func TestVerifyAllModulesInSet(t *testing.T) {
	testName := "verify_all_modules_in_set"
	versionYamlDir := filepath.Join(testDataDir, testName)

	tmpRootDir := t.TempDir()
	testCases := []struct {
		name               string
		versioningFilename string
		repoRoot           string
		modFiles           map[string][]byte
		expectedError      error
	}{
		{
			name:               "valid",
			versioningFilename: filepath.Join(versionYamlDir, "versions_valid.yaml"),
			repoRoot:           filepath.Join(tmpRootDir, "valid"),
			modFiles: map[string][]byte{
				filepath.Join(tmpRootDir, "valid", "test", "test1", "go.mod"):    []byte("module \"go.opentelemetry.io/test/test1\"\n\ngo 1.16\n"),
				filepath.Join(tmpRootDir, "valid", "test", "go.mod"):             []byte("module go.opentelemetry.io/test2\n\ngo 1.16\n"),
				filepath.Join(tmpRootDir, "valid", "go.mod"):                     []byte("module go.opentelemetry.io/testroot/v2\n\ngo 1.16\n"),
				filepath.Join(tmpRootDir, "valid", "test", "excluded", "go.mod"): []byte("module \"go.opentelemetry.io/test/testexcluded\"\n\ngo 1.16\n"),
			},
			expectedError: nil,
		},
		{
			name:               "module not listed",
			versioningFilename: filepath.Join(versionYamlDir, "module_not_listed.yaml"),
			repoRoot:           filepath.Join(tmpRootDir, "not_listed"),
			modFiles: map[string][]byte{
				filepath.Join(tmpRootDir, "not_listed", "test", "test1", "go.mod"):    []byte("module \"go.opentelemetry.io/test/test1\"\n\ngo 1.16\n"),
				filepath.Join(tmpRootDir, "not_listed", "test", "go.mod"):             []byte("module go.opentelemetry.io/test2\n\ngo 1.16\n"),
				filepath.Join(tmpRootDir, "not_listed", "go.mod"):                     []byte("module go.opentelemetry.io/testroot/v2\n\ngo 1.16\n"),
				filepath.Join(tmpRootDir, "not_listed", "test", "excluded", "go.mod"): []byte("module \"go.opentelemetry.io/test/testexcluded\"\n\ngo 1.16\n"),
			},
			expectedError: &errModuleNotInSet{
				modPath:     "go.opentelemetry.io/testroot/v2",
				modFilePath: shared.ModuleFilePath(filepath.Join(tmpRootDir, "not_listed", "go.mod")),
			},
		},
		{
			name:               "module not in repo",
			versioningFilename: filepath.Join(versionYamlDir, "module_not_in_repo.yaml"),
			repoRoot:           filepath.Join(tmpRootDir, "not_in_repo"),
			modFiles: map[string][]byte{
				filepath.Join(tmpRootDir, "not_in_repo", "test", "test1", "go.mod"):    []byte("module \"go.opentelemetry.io/test/test1\"\n\ngo 1.16\n"),
				filepath.Join(tmpRootDir, "not_in_repo", "test", "go.mod"):             []byte("module go.opentelemetry.io/test2\n\ngo 1.16\n"),
				filepath.Join(tmpRootDir, "not_in_repo", "test", "excluded", "go.mod"): []byte("module \"go.opentelemetry.io/test/testexcluded\"\n\ngo 1.16\n"),
			},
			expectedError: &errModuleNotInRepo{
				modPath:    "go.opentelemetry.io/testroot/v2",
				modSetName: "mod-set-3",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			require.NoError(t, sharedtest.WriteTempFiles(tc.modFiles), "could not create go mod file tree")

			v, err := newVerification(tc.versioningFilename, tc.repoRoot)
			require.NoError(t, err)

			actual := v.verifyAllModulesInSet()

			assert.Equal(t, tc.expectedError, actual)
		})
	}
}

func TestVerifyVersions(t *testing.T) {
	testName := "verify_versions"
	versionYamlDir := filepath.Join(testDataDir, testName)

	tmpRootDir := t.TempDir()
	testCases := []struct {
		name               string
		versioningFilename string
		repoRoot           string
		modFiles           map[string][]byte
		expectedError      error
	}{
		{
			name:               "valid",
			versioningFilename: filepath.Join(versionYamlDir, "versions_valid.yaml"),
			repoRoot:           filepath.Join(tmpRootDir, "valid"),
			modFiles: map[string][]byte{
				filepath.Join(tmpRootDir, "valid", "test", "test1", "go.mod"):    []byte("module \"go.opentelemetry.io/test/test1\"\n\ngo 1.16\n"),
				filepath.Join(tmpRootDir, "valid", "test", "go.mod"):             []byte("module go.opentelemetry.io/test2\n\ngo 1.16\n"),
				filepath.Join(tmpRootDir, "valid", "go.mod"):                     []byte("module go.opentelemetry.io/testroot/v2\n\ngo 1.16\n"),
				filepath.Join(tmpRootDir, "valid", "test", "excluded", "go.mod"): []byte("module \"go.opentelemetry.io/test/testexcluded\"\n\ngo 1.16\n"),
			},
			expectedError: nil,
		},
		{
			name:               "invalid version",
			versioningFilename: filepath.Join(versionYamlDir, "invalid_version.yaml"),
			repoRoot:           filepath.Join(tmpRootDir, "invalid_version"),
			modFiles: map[string][]byte{
				filepath.Join(tmpRootDir, "invalid_version", "test", "test1", "go.mod"):    []byte("module \"go.opentelemetry.io/test/test1\"\n\ngo 1.16\n"),
				filepath.Join(tmpRootDir, "invalid_version", "test", "go.mod"):             []byte("module go.opentelemetry.io/test2\n\ngo 1.16\n"),
				filepath.Join(tmpRootDir, "invalid_version", "go.mod"):                     []byte("module go.opentelemetry.io/testroot/v2\n\ngo 1.16\n"),
				filepath.Join(tmpRootDir, "invalid_version", "test", "excluded", "go.mod"): []byte("module \"go.opentelemetry.io/test/testexcluded\"\n\ngo 1.16\n"),
			},
			expectedError: &errInvalidVersion{
				modSetName:    "mod-set-1",
				modSetVersion: "invalid-version-v.02.0.",
			},
		},
		{
			name:               "multiple sets with same major version",
			versioningFilename: filepath.Join(versionYamlDir, "multiple_sets_same_major.yaml"),
			repoRoot:           filepath.Join(tmpRootDir, "multiple_sets_same_major"),
			modFiles: map[string][]byte{
				filepath.Join(tmpRootDir, "multiple_sets_same_major", "test", "test1", "go.mod"):    []byte("module \"go.opentelemetry.io/test/test1\"\n\ngo 1.16\n"),
				filepath.Join(tmpRootDir, "multiple_sets_same_major", "test", "go.mod"):             []byte("module go.opentelemetry.io/test2\n\ngo 1.16\n"),
				filepath.Join(tmpRootDir, "multiple_sets_same_major", "go.mod"):                     []byte("module go.opentelemetry.io/testroot/v2\n\ngo 1.16\n"),
				filepath.Join(tmpRootDir, "multiple_sets_same_major", "test", "excluded", "go.mod"): []byte("module \"go.opentelemetry.io/test/testexcluded\"\n\ngo 1.16\n"),
			},
			expectedError: &errMultipleSetSameVersionSlice{
				errs: []*errMultipleSetSameVersion{
					{
						modSetNames:   []string{"mod-set-1", "mod-set-3", "mod-set-4"},
						modSetVersion: "v1",
					},
					{
						modSetNames:   []string{"mod-set-5", "mod-set-6"},
						modSetVersion: "v1",
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			require.NoError(t, sharedtest.WriteTempFiles(tc.modFiles), "could not create go mod file tree")

			v, err := newVerification(tc.versioningFilename, tc.repoRoot)
			require.NoError(t, err)

			actual := v.verifyVersions()
			if tc.expectedError != nil {
				expectedErr := &errMultipleSetSameVersion{}
				// Check if expectedError is of type errMultipleSetSameVersion
				if errors.As(tc.expectedError, &expectedErr) {
					actualErr := &errMultipleSetSameVersion{}
					require.ErrorAs(t, actual, actualErr)

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
	testName := "verify_dependencies"
	versionYamlDir := filepath.Join(testDataDir, testName)

	tmpRootDir := t.TempDir()
	testCases := []struct {
		name               string
		versioningFilename string
		repoRoot           string
		modFiles           map[string][]byte
		expectWarnings     bool
		expectedLogs       []string
	}{
		{
			name:               "valid",
			versioningFilename: filepath.Join(versionYamlDir, "versions_valid.yaml"),
			repoRoot:           filepath.Join(tmpRootDir, "valid"),
			modFiles: map[string][]byte{
				filepath.Join(tmpRootDir, "valid", "test", "test1", "go.mod"): []byte("module go.opentelemetry.io/build-tools/multimod/internal/verify/test/test1\n\n" +
					"go 1.16\n\n" +
					"require (\n\t" +
					"go.opentelemetry.io/build-tools/multimod/internal/verify/test/test2 v1.2.3-RC1+meta\n" +
					")"),
				filepath.Join(tmpRootDir, "valid", "test", "test2", "go.mod"): []byte("module go.opentelemetry.io/build-tools/multimod/internal/verify/test/test2\n\n" +
					"go 1.16\n\n" +
					"require (\n\t" +
					"go.opentelemetry.io/build-tools/multimod/internal/verify/test/test1 v1.2.3-RC1+meta\n" +
					")"),
				filepath.Join(tmpRootDir, "valid", "test", "go.mod"): []byte("module go.opentelemetry.io/build-tools/multimod/internal/verify/test3\n\n" +
					"go 1.16\n\n" +
					"require (\n\t" +
					"go.opentelemetry.io/build-tools/multimod/internal/verify/test/test1 v1.2.3-RC1+meta\n\t" +
					"go.opentelemetry.io/build-tools/multimod/internal/verify/test/test2 v1.2.3-RC1+meta\n\n" +
					")"),
				filepath.Join(tmpRootDir, "valid", "go.mod"): []byte("module go.opentelemetry.io/build-tools/multimod/internal/verify/testroot\n\n" +
					"go 1.16\n\n" +
					"require (\n\t" +
					"go.opentelemetry.io/build-tools/multimod/internal/verify/test/test1 v1.2.3-RC1+meta\n\t" +
					"go.opentelemetry.io/build-tools/multimod/internal/verify/test/test2 v1.2.3-RC1+meta\n\t" +
					"go.opentelemetry.io/build-tools/multimod/internal/verify/test3 v0.1.0\n" +
					")"),
			},
			expectWarnings: false,
			expectedLogs: []string{
				"Finished checking all stable modules' dependencies.\n",
			},
		},
		{
			name:               "stable depends on unstable",
			versioningFilename: filepath.Join(versionYamlDir, "versions_valid.yaml"),
			repoRoot:           filepath.Join(tmpRootDir, "stable_unstable"),
			modFiles: map[string][]byte{
				filepath.Join(tmpRootDir, "stable_unstable", "test", "test1", "go.mod"): []byte("module go.opentelemetry.io/build-tools/multimod/internal/verify/test/test1\n\n" +
					"go 1.16\n\n" +
					"require (\n\t" +
					"go.opentelemetry.io/build-tools/multimod/internal/verify/test/test2 v1.2.3-RC1+meta\n\t" +
					"go.opentelemetry.io/build-tools/multimod/internal/verify/test3 v0.1.0\n" +
					")"),
				filepath.Join(tmpRootDir, "stable_unstable", "test", "go.mod"): []byte("module go.opentelemetry.io/build-tools/multimod/internal/verify/test3\n\n" +
					"go 1.16\n\n" +
					"require (\n\t" +
					"go.opentelemetry.io/build-tools/multimod/internal/verify/test/test1 v1.2.3-RC1+meta\n\t" +
					"go.opentelemetry.io/build-tools/multimod/internal/verify/test/test2 v1.2.3-RC1+meta\n\n" +
					")"),
				filepath.Join(tmpRootDir, "stable_unstable", "go.mod"): []byte("module go.opentelemetry.io/build-tools/multimod/internal/verify/testroot\n\n" +
					"go 1.16\n\n" +
					"require (\n\t" +
					"go.opentelemetry.io/build-tools/multimod/internal/verify/test/test1 v1.2.3-RC1+meta\n\t" +
					"go.opentelemetry.io/build-tools/multimod/internal/verify/test/test2 v1.2.3-RC1+meta\n\t" +
					"go.opentelemetry.io/build-tools/multimod/internal/verify/test3 v0.1.0\n" +
					")"),
				filepath.Join(tmpRootDir, "stable_unstable", "test", "test2", "go.mod"): []byte("module go.opentelemetry.io/build-tools/multimod/internal/verify/test/test2\n\n" +
					"go 1.16\n\n" +
					"require (\n\t" +
					"go.opentelemetry.io/build-tools/multimod/internal/verify/test/test1 v1.2.3-RC1+meta\n\t" +
					"go.opentelemetry.io/build-tools/multimod/internal/verify/test3 v0.1.0\n\t" +
					"go.opentelemetry.io/build-tools/multimod/internal/verify/testroot v0.2.0\n" +
					")"),
			},
			expectWarnings: true,
			expectedLogs: []string{
				(&errDependency{
					modPath:    "go.opentelemetry.io/build-tools/multimod/internal/verify/test/test1",
					modVersion: "v1.2.3-RC1+meta",
					depPath:    "go.opentelemetry.io/build-tools/multimod/internal/verify/test3",
					depVersion: "v0.1.0",
				}).Error(),
				(&errDependency{
					modPath:    "go.opentelemetry.io/build-tools/multimod/internal/verify/test/test2",
					modVersion: "v1.2.3-RC1+meta",
					depPath:    "go.opentelemetry.io/build-tools/multimod/internal/verify/testroot",
					depVersion: "v0.2.0",
				}).Error(),
				(&errDependency{
					modPath:    "go.opentelemetry.io/build-tools/multimod/internal/verify/test/test2",
					modVersion: "v1.2.3-RC1+meta",
					depPath:    "go.opentelemetry.io/build-tools/multimod/internal/verify/test3",
					depVersion: "v0.1.0",
				}).Error(),
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			require.NoError(t, sharedtest.WriteTempFiles(tc.modFiles), "could not create go mod file tree")

			v, err := newVerification(tc.versioningFilename, tc.repoRoot)
			require.NoError(t, err)

			actual := captureOutput(func() {
				err = v.verifyDependencies()
				require.NoError(t, err)
			})

			if tc.expectWarnings {
				for _, expectedLog := range tc.expectedLogs {
					assert.Contains(t, actual, expectedLog)
				}
			} else {
				assert.NotContains(t, actual, "WARNING")
			}
		})
	}
}
