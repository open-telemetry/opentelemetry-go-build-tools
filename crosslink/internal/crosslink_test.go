package crosslink

import (
	"io/ioutil"
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
func TestExecuteSimple(t *testing.T) {
	testName := "testSimple"

	tmpRootDir, err := os.MkdirTemp(testDataDir, testName)
	if err != nil {
		t.Fatal("creating temp dir:", err)
	}

	mockDataDir := filepath.Join(mockDataDir, testName)
	cp.Copy(mockDataDir, tmpRootDir)

	defer os.RemoveAll(tmpRootDir)

	config := DefaultRunConfig()
	config.rootPath = tmpRootDir

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
func TestExecuteCyclic(t *testing.T) {
	testName := "testCyclic"

	tmpRootDir, err := os.MkdirTemp(testDataDir, testName)
	if err != nil {
		t.Fatal("creating temp dir:", err)
	}

	mockDataDir := filepath.Join(mockDataDir, testName)
	cp.Copy(mockDataDir, tmpRootDir)

	defer os.RemoveAll(tmpRootDir)

	runConfig := DefaultRunConfig()
	runConfig.rootPath = tmpRootDir

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
		verbose:       true,
		overwrite:     true,
		excludedPaths: map[string]struct{}{},
		rootPath:      tmpRootDir,
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
		excludedPaths: map[string]struct{}{},
		rootPath:      tmpRootDir,
		verbose:       true,
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
				prune: true,
				excludedPaths: map[string]struct{}{
					"go.opentelemetry.io/build-tools/crosslink/testroot/testB": {},
					"go.opentelemetry.io/build-tools/excludeme":                {},
					"go.opentelemetry.io/build-tools/crosslink/testroot/testA": {},
				},
				verbose: true,
				logger:  lg,
			},
		},
		{
			testCase: "Overwrite on",
			config: runConfig{
				overwrite: true,
				prune:     true,
				excludedPaths: map[string]struct{}{
					"go.opentelemetry.io/build-tools/crosslink/testroot/testB": {},
					"go.opentelemetry.io/build-tools/excludeme":                {},
					"go.opentelemetry.io/build-tools/crosslink/testroot/testA": {},
				},
				logger:  lg,
				verbose: true,
			},
		},
	}

	for _, tc := range tests {
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
	}
}

// test prune
func TestExecutePrune(t *testing.T) {
	testName := "testPrune"

	tmpRootDir, err := os.MkdirTemp(testDataDir, testName)
	if err != nil {
		t.Fatal("creating temp dir:", err)
	}

	mockDataDir := filepath.Join(mockDataDir, testName)
	cp.Copy(mockDataDir, tmpRootDir)

	defer os.RemoveAll(tmpRootDir)
	modContents, err := ioutil.ReadFile(filepath.Join(tmpRootDir, "go.mod"))
	if err != nil {
		t.Errorf("failed to read mock go.mod file: %v", err)
	}

	mockRequiredReplaceStatements := map[string]struct{}{
		"go.opentelemetry.io/build-tools/crosslink/testroot/testA": {},
		"go.opentelemetry.io/build-tools/crosslink/testroot/testB": {},
		"go.opentelemetry.io/build-tools/crosslink/testroot/testC": {},
		"go.opentelemetry.io/build-tools/crosslink/testroot/testD": {},
		"go.opentelemetry.io/build-tools/crosslink/testroot/testE": {},
		"go.opentelemetry.io/build-tools/crosslink/testroot/testF": {},
		"go.opentelemetry.io/build-tools/crosslink/testroot/testG": {},
		"go.opentelemetry.io/build-tools/crosslink/testroot/testH": {},
		"go.opentelemetry.io/build-tools/crosslink/testroot/testK": {},
	}

	mockModInfo := moduleInfo{
		moduleFilePath:            tmpRootDir,
		moduleContents:            modContents,
		requiredReplaceStatements: mockRequiredReplaceStatements,
	}
	lg, _ := zap.NewProduction()
	assert.NoError(t, pruneReplace("go.opentelemetry.io/build-tools/crosslink/testroot", &mockModInfo, runConfig{prune: true, verbose: true, logger: lg}))

	expectedModFile := []byte("module go.opentelemetry.io/build-tools/crosslink/testroot\n\n" +
		"go 1.17\n\n" +
		"require (\n\t" +
		"go.opentelemetry.io/build-tools/crosslink/testroot/testA v1.0.0\n" +
		"go.opentelemetry.io/build-tools/crosslink/testroot/testB v1.0.0\n" +
		"go.opentelemetry.io/build-tools/crosslink/testroot/testC v1.0.0\n" +
		"go.opentelemetry.io/build-tools/crosslink/testroot/testD v1.0.0\n" +
		"go.opentelemetry.io/build-tools/crosslink/testroot/testE v1.0.0\n" +
		"go.opentelemetry.io/build-tools/crosslink/testroot/testF v1.0.0\n" +
		")\n" +
		"replace go.opentelemetry.io/build-tools/crosslink/testroot/testA => ./testA\n\n" +
		"replace go.opentelemetry.io/build-tools/crosslink/testroot/testB => ./testB\n\n" +
		"replace go.opentelemetry.io/build-tools/crosslink/testroot/testC => ./testC\n\n" +
		"replace go.opentelemetry.io/build-tools/crosslink/testroot/testD => ./testD\n\n" +
		"replace go.opentelemetry.io/build-tools/crosslink/testroot/testE => ./testE\n\n" +
		"replace go.opentelemetry.io/build-tools/crosslink/testroot/testF => ./testF\n\n" +
		"replace go.opentelemetry.io/build-tools/crosslink/testroot/testG => ./testG\n\n" +
		"replace go.opentelemetry.io/build-tools/crosslink/testroot/testH => ./testH\n\n" +
		"replace go.opentelemetry.io/not-a-real-module/testFoo => ./testFoo\n\n" +
		"replace go.opentelemetry.io/fake-module/ => ./fake-module\n\n" +
		"replace go.opentelemetry.io/build-tools/multimod => ../multimod\n\n" +
		"replace foo.opentelemetery.io/bar => ../bar\n\n" +
		"replace go.opentelemetry.io/build-tools/crosslink/testroot/testK => ../crosslinkcopy/testK\n\n")

	expModParse, err := modfile.Parse("go.mod", expectedModFile, nil)
	if err != nil {
		t.Errorf("error parsing expected mod file: %v", err)
	}
	expModParse.Cleanup()

	actual, err := modfile.Parse("go.mod", mockModInfo.moduleContents, nil)
	if err != nil {
		t.Fatalf("error decoding original mod files: %v", err)
	}
	actual.Cleanup()

	// replace structs need to be assorted to avoid flaky fails in test
	replaceSortFunc := func(x, y *modfile.Replace) bool {
		return x.Old.Path < y.Old.Path
	}

	if diff := cmp.Diff(expModParse, actual, cmpopts.IgnoreFields(modfile.Replace{}, "Syntax"),
		cmpopts.IgnoreFields(modfile.File{}, "Require", "Exclude", "Retract", "Syntax"),
		cmpopts.SortSlices(replaceSortFunc),
	); diff != "" {
		t.Errorf("Replace{} mismatch (-want +got):\n%s", diff)
	}

}
