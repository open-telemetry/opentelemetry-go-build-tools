package crosslink

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	cp "github.com/otiai10/copy"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"golang.org/x/mod/modfile"
)

var (
	testDataDir, _ = filepath.Abs("../test_data")
	mockDataDir, _ = filepath.Abs("../mock_test_data")
)

// simple test case is to create a mock repository with file structure listed below
// no overwrites will be necceessary and only inserts will be performed
// ./go.mod root requires  a which needs to add a replace statement for a and b
// ./a/go.mod a requires  b which needs a replace statement for b
// ./b/go.mod
// TODO: add a go.mod that does not match standard naming conventions but is still intra-repository
func TestSimple(t *testing.T) {
	testName := "testSimple"

	tmpRootDir, err := os.MkdirTemp(testDataDir, testName)
	if err != nil {
		t.Fatal("creating temp dir:", err)
	}

	mockDataDir := filepath.Join(mockDataDir, testName)
	cp.Copy(mockDataDir, tmpRootDir)

	defer os.RemoveAll(tmpRootDir)

	config := DefaultRunConfig()
	config.RootPath = tmpRootDir

	assert.NotPanics(t, func() { Crosslink(config) })

	if assert.NoError(t, err, "error message on execution %s") {
		// a mock_test_data_expected folder could be built instead of building expected files by hand.
		modFilesExpected := map[string][]byte{
			filepath.Join(tmpRootDir, "go.mod"): []byte("module go.opentelemetry.io/build-tools/crosslink/testroot\n\n" +
				"go 1.17\n\n" +
				"require (\n\t" +
				"go.opentelemetry.io/build-tools/crosslink/testroot/testA v1.0.0\n" +
				")\n" +
				"replace go.opentelemetry.io/build-tools/crosslink/testroot/testA => ./testA\n\n" +
				"replace go.opentelemetry.io/build-tools/crosslink/testroot/testB => ./testB"),
			filepath.Join(tmpRootDir, "testA", "go.mod"): []byte("module go.opentelemetry.io/build-tools/crosslink/testroot/testA\n\n" +
				"go 1.17\n\n" +
				"require (\n\t" +
				"go.opentelemetry.io/build-tools/crosslink/testroot/testB v1.0.0\n" +
				")\n" +
				"replace go.opentelemetry.io/build-tools/crosslink/testroot/testB => ../testB"),
			filepath.Join(tmpRootDir, "testB", "go.mod"): []byte("module go.opentelemetry.io/build-tools/crosslink/testroot/testB\n\n" +
				"go 1.17\n\n"),
		}

		for modFilePath, modFilesExpected := range modFilesExpected {
			modFileActual, err := os.ReadFile(modFilePath)

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

}

// Also test cyclic
// ./go.mod requires on a see above
// ./a/go.mod requires on a see above and also root reference to a due to b's dependency
// ./b/go.mod requires on root which needs replace statements for root and a
func TestCyclic(t *testing.T) {
	testName := "testCyclic"

	tmpRootDir, err := os.MkdirTemp(testDataDir, testName)
	if err != nil {
		t.Fatal("creating temp dir:", err)
	}

	mockDataDir := filepath.Join(mockDataDir, testName)
	cp.Copy(mockDataDir, tmpRootDir)

	defer os.RemoveAll(tmpRootDir)

	runConfig := DefaultRunConfig()
	runConfig.RootPath = tmpRootDir

	assert.NotPanics(t, func() { Crosslink(runConfig) })

	if assert.NoError(t, err, "error message on execution %s") {
		// a mock_test_data_expected folder could be built instead of building expected files by hand.
		modFilesExpected := map[string][]byte{
			filepath.Join(tmpRootDir, "go.mod"): []byte("module go.opentelemetry.io/build-tools/crosslink/testroot\n\n" +
				"go 1.17\n\n" +
				"require (\n\t" +
				"go.opentelemetry.io/build-tools/crosslink/testroot/testA v1.0.0\n" +
				")\n" +
				"replace go.opentelemetry.io/build-tools/crosslink/testroot/testA => ./testA\n\n" +
				"replace go.opentelemetry.io/build-tools/crosslink/testroot/testB => ./testB"),
			filepath.Join(tmpRootDir, "testA", "go.mod"): []byte("module go.opentelemetry.io/build-tools/crosslink/testroot/testA\n\n" +
				"go 1.17\n\n" +
				"require (\n\t" +
				"go.opentelemetry.io/build-tools/crosslink/testroot/testB v1.0.0\n" +
				")\n" +
				"replace go.opentelemetry.io/build-tools/crosslink/testroot/testB => ../testB\n\n" +
				"replace go.opentelemetry.io/build-tools/crosslink/testroot => ../"),
			// b has req on root but not neccessary to write out with current comparison logic
			filepath.Join(tmpRootDir, "testB", "go.mod"): []byte("module go.opentelemetry.io/build-tools/crosslink/testroot/testB\n\n" +
				"go 1.17\n\n" +
				"replace go.opentelemetry.io/build-tools/crosslink/testroot/testA => ../testA\n\n" +
				"replace go.opentelemetry.io/build-tools/crosslink/testroot => ../\n\n"),
		}

		for modFilePath, modFilesExpected := range modFilesExpected {
			modFileActual, err := os.ReadFile(modFilePath)

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
}

func TestOverwrite(t *testing.T) {
	testName := "testOverwrite"

	tmpRootDir, err := os.MkdirTemp(testDataDir, testName)
	if err != nil {
		t.Fatal("creating temp dir:", err)
	}

	mockDataDir := filepath.Join(mockDataDir, testName)
	cp.Copy(mockDataDir, tmpRootDir)

	defer os.RemoveAll(tmpRootDir)
	lg, _ := zap.NewProduction()
	rc := runConfig{
		Verbose:       true,
		Overwrite:     true,
		ExcludedPaths: map[string]struct{}{},
		RootPath:      tmpRootDir,
		logger:        lg,
	}

	assert.NotPanics(t, func() { Crosslink(rc) })

	if assert.NoError(t, err, "error message on execution %s") {
		// a mock_test_data_expected folder could be built instead of building expected files by hand.
		modFilesExpected := map[string][]byte{
			filepath.Join(tmpRootDir, "go.mod"): []byte("module go.opentelemetry.io/build-tools/crosslink/testroot\n\n" +
				"go 1.17\n\n" +
				"require (\n\t" +
				"go.opentelemetry.io/build-tools/crosslink/testroot/testA v1.0.0\n" +
				")\n" +
				"replace go.opentelemetry.io/build-tools/crosslink/testroot/testA => ./testA\n\n" +
				"replace go.opentelemetry.io/build-tools/crosslink/testroot/testB => ./testB"),
			filepath.Join(tmpRootDir, "testA", "go.mod"): []byte("module go.opentelemetry.io/build-tools/crosslink/testroot/testA\n\n" +
				"go 1.17\n\n" +
				"require (\n\t" +
				"go.opentelemetry.io/build-tools/crosslink/testroot/testB v1.0.0\n" +
				")\n" +
				"replace go.opentelemetry.io/build-tools/crosslink/testroot/testB => ../testB"),
			filepath.Join(tmpRootDir, "testB", "go.mod"): []byte("module go.opentelemetry.io/build-tools/crosslink/testroot/testB\n\n" +
				"go 1.17\n\n"),
		}

		for modFilePath, modFilesExpected := range modFilesExpected {
			modFileActual, err := os.ReadFile(modFilePath)

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
}

func TestNoOverwrite(t *testing.T) {
	testName := "testNoOverwrite"

	tmpRootDir, err := os.MkdirTemp(testDataDir, testName)
	if err != nil {
		t.Fatal("creating temp dir:", err)
	}

	mockDataDir := filepath.Join(mockDataDir, testName)
	cp.Copy(mockDataDir, tmpRootDir)

	defer os.RemoveAll(tmpRootDir)
	lg, _ := zap.NewProduction()
	rc := runConfig{
		ExcludedPaths: map[string]struct{}{},
		RootPath:      tmpRootDir,
		Verbose:       true,
		logger:        lg,
	}

	assert.NotPanics(t, func() { Crosslink(rc) })

	if assert.NoError(t, err, "error message on execution %s") {
		// a mock_test_data_expected folder could be built instead of building expected files by hand.
		modFilesExpected := map[string][]byte{
			filepath.Join(tmpRootDir, "go.mod"): []byte("module go.opentelemetry.io/build-tools/crosslink/testroot\n\n" +
				"go 1.17\n\n" +
				"require (\n\t" +
				"go.opentelemetry.io/build-tools/crosslink/testroot/testA v1.0.0\n" +
				")\n" +
				"replace go.opentelemetry.io/build-tools/crosslink/testroot/testA => ../testA\n\n" +
				"replace go.opentelemetry.io/build-tools/crosslink/testroot/testB => ./testB"),
			filepath.Join(tmpRootDir, "testA", "go.mod"): []byte("module go.opentelemetry.io/build-tools/crosslink/testroot/testA\n\n" +
				"go 1.17\n\n" +
				"require (\n\t" +
				"go.opentelemetry.io/build-tools/crosslink/testroot/testB v1.0.0\n" +
				")\n" +
				"replace go.opentelemetry.io/build-tools/crosslink/testroot/testB => ../testB"),
			filepath.Join(tmpRootDir, "testB", "go.mod"): []byte("module go.opentelemetry.io/build-tools/crosslink/testroot/testB\n\n" +
				"go 1.17\n\n"),
		}

		for modFilePath, modFilesExpected := range modFilesExpected {
			modFileActual, err := os.ReadFile(modFilePath)

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

}

// Testing exclude functionality for prune, overwrite, and no overwrite.
func TestExclude(t *testing.T) {
	testName := "testExclude"
	lg, _ := zap.NewProduction()
	tests := []struct {
		testCase string
		config   runConfig
	}{
		{
			testCase: "Overwrite off",
			config: runConfig{
				Prune: true,
				ExcludedPaths: map[string]struct{}{
					"go.opentelemetry.io/build-tools/crosslink/testroot/testB": {},
					"go.opentelemetry.io/build-tools/excludeme":                {},
					"go.opentelemetry.io/build-tools/crosslink/testroot/testA": {},
				},
				Verbose: true,
				logger:  lg,
			},
		},
		{
			testCase: "Overwrite on",
			config: runConfig{
				Overwrite: true,
				Prune:     true,
				ExcludedPaths: map[string]struct{}{
					"go.opentelemetry.io/build-tools/crosslink/testroot/testB": {},
					"go.opentelemetry.io/build-tools/excludeme":                {},
					"go.opentelemetry.io/build-tools/crosslink/testroot/testA": {},
				},
				logger:  lg,
				Verbose: true,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.testCase, func(t *testing.T) {
			tmpRootDir, err := os.MkdirTemp(testDataDir, testName)
			if err != nil {
				t.Fatal("creating temp dir:", err)
			}

			mockDataDir := filepath.Join(mockDataDir, testName)
			cp.Copy(mockDataDir, tmpRootDir)

			assert.NotPanics(t, func() { Crosslink(tc.config) })
			if assert.NoError(t, err, "error message on execution %s") {
				// a mock_test_data_expected folder could be built instead of building expected files by hand.
				modFilesExpected := map[string][]byte{
					filepath.Join(tmpRootDir, "go.mod"): []byte("module go.opentelemetry.io/build-tools/crosslink/testroot\n\n" +
						"go 1.17\n\n" +
						"require (\n\t" +
						"go.opentelemetry.io/build-tools/crosslink/testroot/testA v1.0.0\n" +
						")\n" +
						"replace go.opentelemetry.io/build-tools/crosslink/testroot/testA => ../testA\n\n" +
						"replace go.opentelemetry.io/build-tools/excludeme => ../excludeme\n\n"),
					filepath.Join(tmpRootDir, "testA", "go.mod"): []byte("module go.opentelemetry.io/build-tools/crosslink/testroot/testA\n\n" +
						"go 1.17\n\n" +
						"require (\n\t" +
						"go.opentelemetry.io/build-tools/crosslink/testroot/testB v1.0.0\n" +
						")\n" +
						"replace go.opentelemetry.io/build-tools/crosslink/testroot/testB => ../testB"),
					filepath.Join(tmpRootDir, "testB", "go.mod"): []byte("module go.opentelemetry.io/build-tools/crosslink/testroot/testB\n\n" +
						"go 1.17\n\n"),
				}

				for modFilePath, modFilesExpected := range modFilesExpected {
					modFileActual, err := os.ReadFile(modFilePath)

					if err != nil {
						t.Fatalf("TestCase: %s, error reading actual mod files: %v", tc.testCase, err)
					}

					actual, err := modfile.Parse("go.mod", modFileActual, nil)
					if err != nil {
						t.Fatalf("error decoding original mod files: %v", err)
					}
					actual.Cleanup()

					expected, err := modfile.Parse("go.mod", modFilesExpected, nil)
					if err != nil {
						t.Fatalf("TestCase: %s ,error decoding expected mod file: %v", tc.testCase, err)
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
						t.Errorf("TestCase: %s \n Replace{} mismatch (-want +got):\n%s", tc.testCase, diff)
					}
				}
			}
			os.RemoveAll(tmpRootDir)
		})

	}
}
