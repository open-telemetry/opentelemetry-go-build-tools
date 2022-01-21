package crosslink

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/mod/modfile"
)

var (
	testDataDir, _ = filepath.Abs("../test_data")
)

// WriteTempFiles is a helper function to dynamically write files such as go.mod or version.go used for testing.
// Duplicated from multimod build tool. Could possible be refactored into root repository common.go
func writeTempFiles(modFiles map[string][]byte) error {
	perm := os.FileMode(0700)

	for modFilePath, file := range modFiles {
		path := filepath.Dir(modFilePath)
		err := os.MkdirAll(path, perm)
		if err != nil {
			return fmt.Errorf("error calling os.MkdirAll(%v, %v): %v", path, perm, err)
		}

		if err := ioutil.WriteFile(modFilePath, file, perm); err != nil {
			return fmt.Errorf("could not write temporary file %v", err)
		}
	}

	return nil
}

// simple test case is to create a mock repository with file structure listed below
// ./go.mod root requires  a which needs to add a replace statement for a and b
// ./a/go.mod a requires  b which needs a replace statement for b
// ./b/go.mod
// TODO: add a go.mod that does not match standard naming conventions but is still intra-repository
func TestExecuteSimple(t *testing.T) {
	testName := "testExecute"

	tmpRootDir, err := os.MkdirTemp(testDataDir, testName)
	if err != nil {
		t.Fatal("creating temp dir:", err)
	}

	defer os.RemoveAll(tmpRootDir)

	modFiles := map[string][]byte{
		filepath.Join(tmpRootDir, "go.mod"): []byte("module go.opentelemetry.io/build-tools/crosslink/testroot\n\n" +
			"go 1.17\n\n" +
			"require (\n\t" +
			"go.opentelemetry.io/build-tools/crosslink/testroot/testA v1.0.0\n" +
			")"),
		filepath.Join(tmpRootDir, "testA", "go.mod"): []byte("module go.opentelemetry.io/build-tools/crosslink/testroot/testA\n\n" +
			"go 1.17\n\n" +
			"require (\n\t" +
			"go.opentelemetry.io/build-tools/crosslink/testroot/testB v1.0.0\n" +
			")"),
		filepath.Join(tmpRootDir, "testB", "go.mod"): []byte("module go.opentelemetry.io/build-tools/crosslink/testroot/testB\n\n" +
			"go 1.17\n\n"),
	}

	if err := writeTempFiles(modFiles); err != nil {
		t.Fatalf("Error writing mod files: %v", err)
	}

	//err = executeCrosslink(tmpRootDir)

	if assert.NoError(t, err, "error message on execution %s") {
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

			assert.Equal(t, len(actual.Replace), len(expected.Replace))
			// do some assertion magic with go-cmp to comapre replace fields
			// and ignore syntax line positioning garbage we do not care about.
		}
	}

}

// Also test cyclic
// ./go.mod requires on a see above
// ./a/go.mod requires on a see above and also root reference to a due to b's dependency
// ./b/go.mod requires on root which needs replace statements for root and a

func TestExecuteCyclic(t *testing.T) {

}

// test prune
func TestExecutePrune(t *testing.T) {

}
