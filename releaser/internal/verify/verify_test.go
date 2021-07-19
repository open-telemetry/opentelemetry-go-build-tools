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
	"io/ioutil"
	"log"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.opentelemetry.io/build-tools/releaser/internal/versions"
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
		log.SetOutput(os.Stderr)
	}()

	f()
	return buf.String()
}

// MockVerification creates a Verification struct for testing purposes.
func MockVerification(modSetMap versions.ModuleSetMap, modPathMap versions.ModulePathMap, dependencies dependencyMap) verification {
	modVersioning, err := versions.MockModuleVersioning(modSetMap, modPathMap)
	if err != nil {
		log.Printf("error getting MockModuleVersioning: %v", err)
		return verification{}
	}

	return verification{
		ModuleVersioning: modVersioning,
		dependencies:     dependencies,
	}
}

func TestMockVerification(t *testing.T) {
	modSetMap := versions.ModuleSetMap{
		"mod-set-1": versions.ModuleSet{
			Version: "v1.2.3-RC1+meta",
			Modules: []versions.ModulePath{
				"go.opentelemetry.io/test/test1",
				"go.opentelemetry.io/test/test2",
			},
		},
		"mod-set-2": versions.ModuleSet{
			Version: "v0.1.0",
			Modules: []versions.ModulePath{
				"go.opentelemetry.io/test3",
			},
		},
	}

	modPathMap := versions.ModulePathMap{
		"go.opentelemetry.io/test/test1": "root/path/to/mod/test/test1/go.mod",
		"go.opentelemetry.io/test/test2": "root/path/to/mod/test/test2/go.mod",
		"go.opentelemetry.io/test3":      "root/test3/go.mod",
	}

	depMap := dependencyMap{
		"go.opentelemetry.io/test/test1": []versions.ModulePath{"go.opentelemetry.io/test/test2"},
		"go.opentelemetry.io/test3":      []versions.ModulePath{"go.opentelemetry.io/test/test1"},
	}

	expected := verification{
		ModuleVersioning: versions.ModuleVersioning{
			ModSetMap:  modSetMap,
			ModPathMap: modPathMap,
			ModInfoMap: versions.ModuleInfoMap{
				"go.opentelemetry.io/test/test1": versions.ModuleInfo{
					ModuleSetName: "mod-set-1",
					Version:       "v1.2.3-RC1+meta",
				},
				"go.opentelemetry.io/test/test2": versions.ModuleInfo{
					ModuleSetName: "mod-set-1",
					Version:       "v1.2.3-RC1+meta",
				},
				"go.opentelemetry.io/test3": versions.ModuleInfo{
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

// Positive-only test
func TestGetDependencies(t *testing.T) {
	modVersioning, _ := versions.MockModuleVersioning(
		versions.ModuleSetMap{
			"mod-set-1": versions.ModuleSet{
				Version: "v1.2.3-RC1+meta",
				Modules: []versions.ModulePath{
					"go.opentelemetry.io/build-tools/releaser/internal/verify/test/test1",
					"go.opentelemetry.io/build-tools/releaser/internal/verify/test/test2",
				},
			},
			"mod-set-2": versions.ModuleSet{
				Version: "v0.1.0",
				Modules: []versions.ModulePath{
					"go.opentelemetry.io/build-tools/releaser/internal/verify/test3",
				},
			},
			"mod-set-3": versions.ModuleSet{
				Version: "v0.2.0",
				Modules: []versions.ModulePath{
					"go.opentelemetry.io/build-tools/releaser/internal/verify/testroot",
				},
			},
		},
		versions.ModulePathMap{
			"go.opentelemetry.io/build-tools/releaser/internal/verify/test/test1": "./test_data/new_verification/test/test1/go.mod",
			"go.opentelemetry.io/build-tools/releaser/internal/verify/test/test2": "./test_data/new_verification/test/test2/go.mod",
			"go.opentelemetry.io/build-tools/releaser/internal/verify/test3":      "./test_data/new_verification/test/go.mod",
			"go.opentelemetry.io/build-tools/releaser/internal/verify/testroot":   "./test_data/new_verification/go.mod",
		},
	)

	expected := dependencyMap{
		"go.opentelemetry.io/build-tools/releaser/internal/verify/test/test1": []versions.ModulePath{
			"go.opentelemetry.io/build-tools/releaser/internal/verify/test/test2",
			"go.opentelemetry.io/build-tools/releaser/internal/verify/test3",
		},
		"go.opentelemetry.io/build-tools/releaser/internal/verify/test/test2": []versions.ModulePath{
			"go.opentelemetry.io/build-tools/releaser/internal/verify/test/test1",
			"go.opentelemetry.io/build-tools/releaser/internal/verify/test3",
			"go.opentelemetry.io/build-tools/releaser/internal/verify/testroot",
		},
		"go.opentelemetry.io/build-tools/releaser/internal/verify/test3": []versions.ModulePath{
			"go.opentelemetry.io/build-tools/releaser/internal/verify/test/test1",
			"go.opentelemetry.io/build-tools/releaser/internal/verify/test/test2",
		},
		"go.opentelemetry.io/build-tools/releaser/internal/verify/testroot": []versions.ModulePath{
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
				versions.ModuleSetMap{
					"mod-set-1": versions.ModuleSet{
						Version: "v1.2.3-RC1+meta",
						Modules: []versions.ModulePath{
							"go.opentelemetry.io/test/test1",
							"go.opentelemetry.io/test/test2",
						},
					},
					"mod-set-2": versions.ModuleSet{
						Version: "v0.1.0",
						Modules: []versions.ModulePath{
							"go.opentelemetry.io/test3",
						},
					},
				},
				versions.ModulePathMap{
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
				versions.ModuleSetMap{
					"mod-set-1": versions.ModuleSet{
						Version: "v1.2.3-RC1+meta",
						Modules: []versions.ModulePath{
							"go.opentelemetry.io/test/test1",
						},
					},
					"mod-set-2": versions.ModuleSet{
						Version: "v0.1.0",
						Modules: []versions.ModulePath{
							"go.opentelemetry.io/test3",
						},
					},
				},
				versions.ModulePathMap{
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
				versions.ModuleSetMap{
					"mod-set-1": versions.ModuleSet{
						Version: "v1.2.3-RC1+meta",
						Modules: []versions.ModulePath{
							"go.opentelemetry.io/test/test1",
							"go.opentelemetry.io/test/test2",
						},
					},
					"mod-set-2": versions.ModuleSet{
						Version: "v0.1.0",
						Modules: []versions.ModulePath{
							"go.opentelemetry.io/test3",
						},
					},
				},
				versions.ModulePathMap{
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
				versions.ModuleSetMap{
					"mod-set-1": versions.ModuleSet{
						Version: "v1.2.3-RC1+meta",
						Modules: []versions.ModulePath{
							"go.opentelemetry.io/test/test1",
							"go.opentelemetry.io/test/test2",
						},
					},
					"mod-set-2": versions.ModuleSet{
						Version: "v0.1.0",
						Modules: []versions.ModulePath{
							"go.opentelemetry.io/test3",
						},
					},
					"mod-set-3": versions.ModuleSet{
						Version: "v0.1.0",
						Modules: []versions.ModulePath{
							"go.opentelemetry.io/test4",
						},
					},
					"mod-set-4": versions.ModuleSet{
						Version: "v2.2.2-RC1+meta",
						Modules: []versions.ModulePath{
							"go.opentelemetry.io/test5",
						},
					},
				},
				versions.ModulePathMap{
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
				versions.ModuleSetMap{
					"mod-set-1": versions.ModuleSet{
						Version: "invalid-version-v.02.0.",
						Modules: []versions.ModulePath{
							"go.opentelemetry.io/test/test1",
							"go.opentelemetry.io/test/test2",
						},
					},
					"mod-set-2": versions.ModuleSet{
						Version: "v0.1.0",
						Modules: []versions.ModulePath{
							"go.opentelemetry.io/test3",
						},
					},
				},
				versions.ModulePathMap{
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
				versions.ModuleSetMap{
					"mod-set-1": versions.ModuleSet{
						Version: "v1.2.3-RC1+meta",
						Modules: []versions.ModulePath{
							"go.opentelemetry.io/test/test1",
							"go.opentelemetry.io/test/test2",
						},
					},
					"mod-set-2": versions.ModuleSet{
						Version: "v0.1.0",
						Modules: []versions.ModulePath{
							"go.opentelemetry.io/test3",
						},
					},
					"mod-set-3": versions.ModuleSet{
						Version: "v1.1.0-duplicatedmajor",
						Modules: []versions.ModulePath{
							"go.opentelemetry.io/test4",
						},
					},
					"mod-set-4": versions.ModuleSet{
						Version: "v1.9.0-anotherduplicatedmajor",
						Modules: []versions.ModulePath{
							"go.opentelemetry.io/test5",
						},
					},
				},
				versions.ModulePathMap{
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
				versions.ModuleSetMap{
					"mod-set-1": versions.ModuleSet{
						Version: "v1.2.3-RC1+meta",
						Modules: []versions.ModulePath{
							"go.opentelemetry.io/test/test1",
							"go.opentelemetry.io/test/test2",
						},
					},
					"mod-set-2": versions.ModuleSet{
						Version: "v0.1.0",
						Modules: []versions.ModulePath{
							"go.opentelemetry.io/test3",
						},
					},
					"mod-set-3": versions.ModuleSet{
						Version: "v0.1.0",
						Modules: []versions.ModulePath{
							"go.opentelemetry.io/test4",
						},
					},
					"mod-set-4": versions.ModuleSet{
						Version: "v2.2.2-RC1+meta",
						Modules: []versions.ModulePath{
							"go.opentelemetry.io/test5",
						},
					},
				},
				versions.ModulePathMap{
					"go.opentelemetry.io/test/test1": "root/path/to/mod/test/test1/go.mod",
					"go.opentelemetry.io/test/test2": "root/path/to/mod/test/test2/go.mod",
					"go.opentelemetry.io/test3":      "root/test3/go.mod",
					"go.opentelemetry.io/test4":      "root/test4/go.mod",
					"go.opentelemetry.io/test5":      "root/test5/go.mod",
				},
				dependencyMap{
					"go.opentelemetry.io/test/test1": []versions.ModulePath{
						"go.opentelemetry.io/test/test2",
						"go.opentelemetry.io/test5",
					},
					"go.opentelemetry.io/test/test2": []versions.ModulePath{
						"go.opentelemetry.io/test/test1",
						"go.opentelemetry.io/test5",
					},
					"go.opentelemetry.io/test3": []versions.ModulePath{
						"go.opentelemetry.io/test/test1",
						"go.opentelemetry.io/test/test2",
						"go.opentelemetry.io/test4",
						"go.opentelemetry.io/test5",
					},
					"go.opentelemetry.io/test4": []versions.ModulePath{
						"go.opentelemetry.io/test3",
						"go.opentelemetry.io/test5",
					},
					"go.opentelemetry.io/test5": []versions.ModulePath{
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
				versions.ModuleSetMap{
					"mod-set-1": versions.ModuleSet{
						Version: "v1.2.3-RC1+meta",
						Modules: []versions.ModulePath{
							"go.opentelemetry.io/test/test1",
							"go.opentelemetry.io/test/test2",
						},
					},
					"mod-set-2": versions.ModuleSet{
						Version: "v0.1.0",
						Modules: []versions.ModulePath{
							"go.opentelemetry.io/test3",
						},
					},
				},
				versions.ModulePathMap{
					"go.opentelemetry.io/test/test1": "root/path/to/mod/test/test1/go.mod",
					"go.opentelemetry.io/test/test2": "root/path/to/mod/test/test2/go.mod",
					"go.opentelemetry.io/test3":      "root/test3/go.mod",
				},
				dependencyMap{
					"go.opentelemetry.io/test/test1": []versions.ModulePath{
						"go.opentelemetry.io/test2",
						"go.opentelemetry.io/test3",
					},
					"go.opentelemetry.io/test/test2": []versions.ModulePath{
						"go.opentelemetry.io/test1",
						"go.opentelemetry.io/test3",
					},
					"go.opentelemetry.io/test/test3": []versions.ModulePath{
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
