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

package crosslink

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"golang.org/x/mod/modfile"
)

func TestCrosslink(t *testing.T) {
	lg, _ := zap.NewDevelopment()

	tests := []struct {
		testName string
		mockDir  string
		config   RunConfig
		expected map[string][]byte
	}{
		{
			testName: "testSimple",
			mockDir:  "testSimple",
			config:   DefaultRunConfig(),
			expected: map[string][]byte{
				"go.mod": []byte("module go.opentelemetry.io/build-tools/crosslink/testroot\n\n" +
					"go 1.19\n\n" +
					"require (\n\t" +
					"go.opentelemetry.io/build-tools/crosslink/testroot/testA v1.0.0\n" +
					")\n" +
					"replace go.opentelemetry.io/build-tools/crosslink/testroot/testA => ./testA\n\n" +
					"replace go.opentelemetry.io/build-tools/crosslink/testroot/testY => ./testY\n\n" +
					"replace go.opentelemetry.io/build-tools/crosslink/testroot/testZ => ./testZ\n\n" +
					"replace go.opentelemetry.io/build-tools/crosslink/testroot/testB => ./testB"),
				filepath.Join("testA", "go.mod"): []byte("module go.opentelemetry.io/build-tools/crosslink/testroot/testA\n\n" +
					"go 1.19\n\n" +
					"require (\n\t" +
					"go.opentelemetry.io/build-tools/crosslink/testroot/testB v1.0.0\n" +
					")\n" +
					"replace go.opentelemetry.io/build-tools/crosslink/testroot/testB => ../testB"),
				filepath.Join("testB", "go.mod"): []byte("module go.opentelemetry.io/build-tools/crosslink/testroot/testB\n\n" +
					"go 1.19\n\n"),
			},
		},
		{
			testName: "testCyclic",
			mockDir:  "testCyclic",
			config:   DefaultRunConfig(),
			expected: map[string][]byte{
				"go.mod": []byte("module go.opentelemetry.io/build-tools/crosslink/testroot\n\n" +
					"go 1.19\n\n" +
					"require (\n\t" +
					"go.opentelemetry.io/build-tools/crosslink/testroot/testA v1.0.0\n" +
					")\n" +
					"replace go.opentelemetry.io/build-tools/crosslink/testroot/testA => ./testA\n\n" +
					"replace go.opentelemetry.io/build-tools/crosslink/testroot/testB => ./testB"),
				filepath.Join("testA", "go.mod"): []byte("module go.opentelemetry.io/build-tools/crosslink/testroot/testA\n\n" +
					"go 1.19\n\n" +
					"require (\n\t" +
					"go.opentelemetry.io/build-tools/crosslink/testroot/testB v1.0.0\n" +
					")\n" +
					"replace go.opentelemetry.io/build-tools/crosslink/testroot/testB => ../testB\n\n" +
					"replace go.opentelemetry.io/build-tools/crosslink/testroot => ../"),
				// b has req on root but not necessary to write out with current comparison logic
				filepath.Join("testB", "go.mod"): []byte("module go.opentelemetry.io/build-tools/crosslink/testroot/testB\n\n" +
					"go 1.19\n\n" +
					"replace go.opentelemetry.io/build-tools/crosslink/testroot/testA => ../testA\n\n" +
					"replace go.opentelemetry.io/build-tools/crosslink/testroot => ../\n\n"),
			},
		},
		{
			testName: "testSimpleWithPrune",
			mockDir:  "testSimple",
			config: RunConfig{
				Prune:  true,
				Logger: lg,
			},
			expected: map[string][]byte{
				"go.mod": []byte("module go.opentelemetry.io/build-tools/crosslink/testroot\n\n" +
					"go 1.19\n\n" +
					"require (\n\t" +
					"go.opentelemetry.io/build-tools/crosslink/testroot/testA v1.0.0\n" +
					")\n" +
					"replace go.opentelemetry.io/build-tools/crosslink/testroot/testA => ./testA\n\n" +
					"replace go.opentelemetry.io/build-tools/crosslink/testroot/testB => ./testB"),
				filepath.Join("testA", "go.mod"): []byte("module go.opentelemetry.io/build-tools/crosslink/testroot/testA\n\n" +
					"go 1.19\n\n" +
					"require (\n\t" +
					"go.opentelemetry.io/build-tools/crosslink/testroot/testB v1.0.0\n" +
					")\n" +
					"replace go.opentelemetry.io/build-tools/crosslink/testroot/testB => ../testB"),
				filepath.Join("testB", "go.mod"): []byte("module go.opentelemetry.io/build-tools/crosslink/testroot/testB\n\n" +
					"go 1.19\n\n"),
			},
		},
	}

	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			tmpRootDir, err := createTempTestDir(test.mockDir)
			if err != nil {
				t.Fatal("creating temp dir:", err)
			}

			err = renameGoMod(tmpRootDir)
			if err != nil {
				t.Errorf("error renaming gomod files: %v", err)
			}
			t.Cleanup(func() { os.RemoveAll(tmpRootDir) })

			test.config.RootPath = tmpRootDir

			err = Crosslink(test.config)

			if assert.NoError(t, err, "error message on execution %s") {

				for modFilePath, modFilesExpected := range test.expected {
					modFileActual, err := os.ReadFile(filepath.Clean(filepath.Join(tmpRootDir, modFilePath)))

					if err != nil {
						t.Fatalf("error reading actual mod files: %v", err)
					}

					actual, err := modfile.Parse("go.mod", modFileActual, nil)
					if err != nil {
						t.Fatalf("error decoding original mod files: %v", err)
					}
					actual.Cleanup()

					expected, err := modfile.Parse("go.mod", modFilesExpected, nil)
					if err != nil {
						t.Fatalf("error decoding expected mod file: %v", err)
					}
					expected.Cleanup()

					// replace structs need to be assorted to avoid flaky fails in test
					replaceSortFunc := func(x, y *modfile.Replace) bool {
						return x.Old.Path < y.Old.Path
					}

					if diff := cmp.Diff(expected, actual, cmpopts.IgnoreFields(modfile.Replace{}, "Syntax"),
						cmpopts.IgnoreFields(modfile.File{}, "Require", "Exclude", "Retract", "Syntax"),
						cmpopts.SortSlices(replaceSortFunc),
					); diff != "" {
						t.Errorf("Replace{} mismatch (-want +got):\n%s", diff)
					}
				}
			}

		})
	}
}

func TestOverwrite(t *testing.T) {
	lg, _ := zap.NewDevelopment()

	tests := []struct {
		testName string
		config   RunConfig
		expected map[string][]byte
	}{
		{
			testName: "testOverwrite",
			config: RunConfig{
				Verbose:       true,
				Overwrite:     true,
				ExcludedPaths: map[string]struct{}{},
				Logger:        lg,
			},
			expected: map[string][]byte{
				"go.mod": []byte("module go.opentelemetry.io/build-tools/crosslink/testroot\n\n" +
					"go 1.19\n\n" +
					"require (\n\t" +
					"go.opentelemetry.io/build-tools/crosslink/testroot/testA v1.0.0\n" +
					")\n" +
					"replace go.opentelemetry.io/build-tools/crosslink/testroot/testA => ./testA\n\n" +
					"replace go.opentelemetry.io/build-tools/crosslink/testroot/testB => ./testB"),
				filepath.Join("testA", "go.mod"): []byte("module go.opentelemetry.io/build-tools/crosslink/testroot/testA\n\n" +
					"go 1.19\n\n" +
					"require (\n\t" +
					"go.opentelemetry.io/build-tools/crosslink/testroot/testB v1.0.0\n" +
					")\n" +
					"replace go.opentelemetry.io/build-tools/crosslink/testroot/testB => ../testB"),
				filepath.Join("testB", "go.mod"): []byte("module go.opentelemetry.io/build-tools/crosslink/testroot/testB\n\n" +
					"go 1.19\n\n"),
			},
		},
		{
			testName: "testNoOverwrite",
			config: RunConfig{
				ExcludedPaths: map[string]struct{}{},
				Verbose:       true,
				Logger:        lg,
			},
			expected: map[string][]byte{
				"go.mod": []byte("module go.opentelemetry.io/build-tools/crosslink/testroot\n\n" +
					"go 1.19\n\n" +
					"require (\n\t" +
					"go.opentelemetry.io/build-tools/crosslink/testroot/testA v1.0.0\n" +
					")\n" +
					"replace go.opentelemetry.io/build-tools/crosslink/testroot/testA => ../testA\n\n" +
					"replace go.opentelemetry.io/build-tools/crosslink/testroot/testB => ./testB"),
				filepath.Join("testA", "go.mod"): []byte("module go.opentelemetry.io/build-tools/crosslink/testroot/testA\n\n" +
					"go 1.19\n\n" +
					"require (\n\t" +
					"go.opentelemetry.io/build-tools/crosslink/testroot/testB v1.0.0\n" +
					")\n" +
					"replace go.opentelemetry.io/build-tools/crosslink/testroot/testB => ../testB"),
				filepath.Join("testB", "go.mod"): []byte("module go.opentelemetry.io/build-tools/crosslink/testroot/testB\n\n" +
					"go 1.19\n\n"),
			},
		},
	}

	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			tmpRootDir, err := createTempTestDir(test.testName)
			if err != nil {
				t.Fatal("creating temp dir:", err)
			}

			t.Cleanup(func() { os.RemoveAll(tmpRootDir) })

			err = renameGoMod(tmpRootDir)
			if err != nil {
				t.Errorf("error renaming gomod files: %v", err)
			}

			test.config.RootPath = tmpRootDir

			err = Crosslink(test.config)

			if assert.NoError(t, err, "error message on execution %s") {
				// a mock_test_data_expected folder could be built instead of building expected files by hand.

				for modFilePath, modFilesExpected := range test.expected {
					modFileActual, err := os.ReadFile(filepath.Clean(filepath.Join(tmpRootDir, modFilePath)))

					if err != nil {
						t.Fatalf("error reading actual mod files: %v", err)
					}

					actual, err := modfile.Parse("go.mod", modFileActual, nil)
					if err != nil {
						t.Fatalf("error decoding original mod files: %v", err)
					}
					actual.Cleanup()

					expected, err := modfile.Parse("go.mod", modFilesExpected, nil)
					if err != nil {
						t.Fatalf("error decoding expected mod file: %v", err)
					}
					expected.Cleanup()

					// replace structs need to be assorted to avoid flaky fails in test
					replaceSortFunc := func(x, y *modfile.Replace) bool {
						return x.Old.Path < y.Old.Path
					}

					if diff := cmp.Diff(expected, actual, cmpopts.IgnoreFields(modfile.Replace{}, "Syntax"),
						cmpopts.IgnoreFields(modfile.File{}, "Require", "Exclude", "Retract", "Syntax"),
						cmpopts.SortSlices(replaceSortFunc),
					); diff != "" {
						t.Errorf("Replace{} mismatch (-want +got):\n%s", diff)
					}
				}
			}

		})
	}
	err := lg.Sync()
	if err != nil {
		fmt.Printf("failed to sync logger:  %v", err)
	}

}

// Testing exclude functionality for prune, overwrite, and no overwrite.
func TestExclude(t *testing.T) {
	testName := "testExclude"
	lg, _ := zap.NewDevelopment()
	tests := []struct {
		testCase string
		config   RunConfig
	}{
		{
			testCase: "Overwrite off",
			config: RunConfig{
				Prune: true,
				ExcludedPaths: map[string]struct{}{
					"go.opentelemetry.io/build-tools/crosslink/testroot/testB": {},
					"go.opentelemetry.io/build-tools/excludeme":                {},
					"go.opentelemetry.io/build-tools/crosslink/testroot/testA": {},
				},
				Verbose: true,
				Logger:  lg,
			},
		},
		{
			testCase: "Overwrite on",
			config: RunConfig{
				Overwrite: true,
				Prune:     true,
				ExcludedPaths: map[string]struct{}{
					"go.opentelemetry.io/build-tools/crosslink/testroot/testB": {},
					"go.opentelemetry.io/build-tools/excludeme":                {},
					"go.opentelemetry.io/build-tools/crosslink/testroot/testA": {},
				},
				Logger:  lg,
				Verbose: true,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.testCase, func(t *testing.T) {
			tmpRootDir, err := createTempTestDir(testName)
			if err != nil {
				t.Fatal("creating temp dir:", err)
			}
			t.Cleanup(func() { os.RemoveAll(tmpRootDir) })

			err = renameGoMod(tmpRootDir)
			if err != nil {
				t.Errorf("error renaming gomod files: %v", err)
			}

			test.config.RootPath = tmpRootDir

			err = Crosslink(test.config)

			if assert.NoError(t, err, "error message on execution %s") {
				// a mock_test_data_expected folder could be built instead of building expected files by hand.
				modFilesExpected := map[string][]byte{
					filepath.Join(tmpRootDir, "go.mod"): []byte("module go.opentelemetry.io/build-tools/crosslink/testroot\n\n" +
						"go 1.19\n\n" +
						"require (\n\t" +
						"go.opentelemetry.io/build-tools/crosslink/testroot/testA v1.0.0\n" +
						")\n" +
						"replace go.opentelemetry.io/build-tools/crosslink/testroot/testA => ../testA\n\n" +
						"replace go.opentelemetry.io/build-tools/excludeme => ../excludeme\n\n"),
					filepath.Join(tmpRootDir, "testA", "go.mod"): []byte("module go.opentelemetry.io/build-tools/crosslink/testroot/testA\n\n" +
						"go 1.19\n\n" +
						"require (\n\t" +
						"go.opentelemetry.io/build-tools/crosslink/testroot/testB v1.0.0\n" +
						")\n" +
						"replace go.opentelemetry.io/build-tools/crosslink/testroot/testB => ../testB"),
					filepath.Join(tmpRootDir, "testB", "go.mod"): []byte("module go.opentelemetry.io/build-tools/crosslink/testroot/testB\n\n" +
						"go 1.19\n\n"),
				}

				for modFilePath, modFilesExpected := range modFilesExpected {
					modFileActual, err := os.ReadFile(filepath.Clean(modFilePath))

					if err != nil {
						t.Fatalf("TestCase: %s, error reading actual mod files: %v", test.testCase, err)
					}

					actual, err := modfile.Parse("go.mod", modFileActual, nil)
					if err != nil {
						t.Fatalf("error decoding original mod files: %v", err)
					}
					actual.Cleanup()

					expected, err := modfile.Parse("go.mod", modFilesExpected, nil)
					if err != nil {
						t.Fatalf("TestCase: %s ,error decoding expected mod file: %v", test.testCase, err)
					}
					expected.Cleanup()

					// replace structs need to be assorted to avoid flaky fails in test
					replaceSortFunc := func(x, y *modfile.Replace) bool {
						return x.Old.Path < y.Old.Path
					}

					if diff := cmp.Diff(expected, actual, cmpopts.IgnoreFields(modfile.Replace{}, "Syntax"),
						cmpopts.IgnoreFields(modfile.File{}, "Require", "Exclude", "Retract", "Syntax"),
						cmpopts.SortSlices(replaceSortFunc),
					); diff != "" {
						t.Errorf("TestCase: %s \n Replace{} mismatch (-want +got):\n%s", test.testCase, diff)
					}
				}
			}
		})
	}
}

func TestBadRootPath(t *testing.T) {
	lg, _ := zap.NewDevelopment()
	tests := []struct {
		testName      string
		mockDir       string
		setConfigPath bool
		config        RunConfig
	}{
		{
			testName:      "noGoMod",
			mockDir:       "noGoMod",
			setConfigPath: true,
			config: RunConfig{
				Logger: lg,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			tmpRootDir, err := os.MkdirTemp(testDataDir, test.mockDir)
			if err != nil {
				t.Fatalf("Failed to create temp dir %v", err)
			}
			if test.setConfigPath {
				test.config.RootPath = tmpRootDir
			}

			t.Cleanup(func() { os.RemoveAll(tmpRootDir) })
			err = Crosslink(test.config)
			assert.Error(t, err)
			err = Prune(test.config)
			assert.Error(t, err)
		})
	}

}

func TestGoWork(t *testing.T) {
	lg, _ := zap.NewDevelopment()

	tests := []struct {
		testName string
		config   RunConfig
		expected string
	}{
		{
			testName: "default",
			config:   RunConfig{Logger: lg},
			expected: `go 1.19

			// existing valid use statements under root should remain
			use ./testA
		
			// new statement added by crossling
			use ./testB
			
			// invalid use statements under root should be removed ONLY if prune is used
			use ./testC
			
			// use statements outside the root should remain
			use ../other-module
			
			// replace statements should remain
			replace foo.opentelemetery.io/bar => ../bar`,
		},
		{
			testName: "prune",
			config:   RunConfig{Logger: lg, Prune: true},
			expected: `go 1.19

			// existing valid use statements under root should remain
			use ./testA
		
			// new statement added by crossling
			use ./testB
			
			// invalid use statements under root is REMOVED when prune is used
			// use ./testC
			
			// use statements outside the root should remain
			use ../other-module
			
			// replace statements should remain
			replace foo.opentelemetery.io/bar => ../bar`,
		},
	}

	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			mockDir := "testGoWork"
			tmpRootDir, err := createTempTestDir(mockDir)
			if err != nil {
				t.Fatal("creating temp dir:", err)
			}

			err = renameGoMod(tmpRootDir)
			if err != nil {
				t.Errorf("error renaming gomod files: %v", err)
			}
			t.Cleanup(func() { os.RemoveAll(tmpRootDir) })

			test.config.RootPath = tmpRootDir

			err = Crosslink(test.config)
			require.NoError(t, err)
			goWorkContent, err := os.ReadFile(filepath.Clean(filepath.Join(tmpRootDir, "go.work")))
			require.NoError(t, err)

			actual, err := modfile.ParseWork("go.work", goWorkContent, nil)
			require.NoError(t, err)
			actual.Cleanup()

			expected, err := modfile.ParseWork("go.work", []byte(test.expected), nil)
			require.NoError(t, err)
			expected.Cleanup()

			// replace structs need to be assorted to avoid flaky fails in test
			replaceSortFunc := func(x, y *modfile.Replace) bool {
				return x.Old.Path < y.Old.Path
			}

			// use structs need to be assorted to avoid flaky fails in test
			useSortFunc := func(x, y *modfile.Use) bool {
				return x.Path < y.Path
			}

			if diff := cmp.Diff(expected, actual,
				cmpopts.IgnoreFields(modfile.Use{}, "Syntax", "ModulePath"),
				cmpopts.IgnoreFields(modfile.Replace{}, "Syntax"),
				cmpopts.IgnoreFields(modfile.WorkFile{}, "Syntax"),
				cmpopts.SortSlices(replaceSortFunc),
				cmpopts.SortSlices(useSortFunc),
			); diff != "" {
				t.Errorf("go.work mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestNoGoWork(t *testing.T) {
	// go.work file is not created by crosslink. It is only modified if it exists.
	mockDir := "testSimple"
	config := DefaultRunConfig()

	tmpRootDir, err := createTempTestDir(mockDir)
	if err != nil {
		t.Fatal("creating temp dir:", err)
	}

	err = renameGoMod(tmpRootDir)
	if err != nil {
		t.Errorf("error renaming gomod files: %v", err)
	}
	t.Cleanup(func() { os.RemoveAll(tmpRootDir) })

	config.RootPath = tmpRootDir

	err = Crosslink(config)

	require.NoError(t, err)
	assert.NoFileExists(t, filepath.Join(tmpRootDir, "go.work"))
}
