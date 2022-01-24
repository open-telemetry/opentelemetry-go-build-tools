package crosslink

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	cp "github.com/otiai10/copy"
	"github.com/stretchr/testify/assert"
	"golang.org/x/mod/modfile"
)

var (
	testDataDir, _ = filepath.Abs("../test_data")
	mockDataDir, _ = filepath.Abs("../mock_test_data")
)

// simple test case is to create a mock repository with file structure listed below
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

	assert.NotPanics(t, func() { Crosslink(tmpRootDir) })

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

	assert.NotPanics(t, func() { Crosslink(tmpRootDir) })

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

// test prune
func TestExecutePrune(t *testing.T) {

}
