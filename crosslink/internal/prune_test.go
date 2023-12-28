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
	"os"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"golang.org/x/mod/modfile"
)

// TODO: Add cleanup to os.removeall

func TestPrune(t *testing.T) {
	lg, _ := zap.NewDevelopment()
	tests := []struct {
		testName string
		mockDir  string
		config   RunConfig
		expected map[string][]byte
	}{
		{
			testName: "testSimplePrune",
			mockDir:  "testSimplePrune",
			config: RunConfig{
				Logger:  lg,
				Prune:   true,
				Verbose: true,
			},
			expected: map[string][]byte{
				"go.mod": []byte("module go.opentelemetry.io/build-tools/crosslink/testroot\n\n" +
					"go 1.20\n\n" +
					"require (\n\t" +
					"go.opentelemetry.io/build-tools/crosslink/testroot/testA v1.0.0\n" +
					")\n" +
					"replace go.opentelemetry.io/build-tools/crosslink/testroot/testA => ./testA\n\n" +
					"replace go.opentelemetry.io/build-tools/crosslink/testroot/testB => ./testB"),
				filepath.Join("testA", "go.mod"): []byte("module go.opentelemetry.io/build-tools/crosslink/testroot/testA\n\n" +
					"go 1.20\n\n" +
					"require (\n\t" +
					"go.opentelemetry.io/build-tools/crosslink/testroot/testB v1.0.0\n" +
					")\n" +
					"replace go.opentelemetry.io/build-tools/crosslink/testroot/testB => ../testB"),
				filepath.Join("testB", "go.mod"): []byte("module go.opentelemetry.io/build-tools/crosslink/testroot/testB\n\n" +
					"go 1.20\n\n"),
			},
		},
	}

	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			tmpRootDir, err := createTempTestDir(test.mockDir)
			if err != nil {
				t.Fatal("creating temp dir:", err)
			}
			t.Cleanup(func() { os.RemoveAll(tmpRootDir) })

			err = renameGoMod(tmpRootDir)
			if err != nil {
				t.Errorf("error renaming gomod files: %v", err)
			}

			test.config.RootPath = tmpRootDir
			err = Prune(test.config)

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

func TestPruneReplace(t *testing.T) {
	testName := "testPrune"

	tmpRootDir, err := createTempTestDir(testName)
	if err != nil {
		t.Fatal("creating temp dir:", err)
	}

	t.Cleanup(func() { os.RemoveAll(tmpRootDir) })

	err = renameGoMod(tmpRootDir)
	if err != nil {
		t.Errorf("error renaming gomod files: %v", err)
	}

	modContents, err := os.ReadFile(filepath.Clean(filepath.Join(tmpRootDir, "go.mod")))
	if err != nil {
		t.Errorf("failed to read mock gomod file: %v", err)
	}

	modFile, err := modfile.Parse("go.mod", modContents, nil)
	if err != nil {
		t.Errorf("failed to parse mock gomod file: %v", err)
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

	mockModInfo := newModuleInfo(*modFile)
	mockModInfo.requiredReplaceStatements = mockRequiredReplaceStatements
	lg, _ := zap.NewDevelopment()
	pruneReplace("go.opentelemetry.io/build-tools/crosslink/testroot", mockModInfo, RunConfig{Prune: true, Verbose: true, Logger: lg})

	expectedModFile := []byte("module go.opentelemetry.io/build-tools/crosslink/testroot\n\n" +
		"go 1.20\n\n" +
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

	actual := mockModInfo.moduleContents
	actual.Cleanup()

	// replace structs need to be assorted to avoid flaky fails in test
	replaceSortFunc := func(x, y *modfile.Replace) bool {
		return x.Old.Path < y.Old.Path
	}

	if diff := cmp.Diff(*expModParse, actual, cmpopts.IgnoreFields(modfile.Replace{}, "Syntax"),
		cmpopts.IgnoreFields(modfile.File{}, "Require", "Exclude", "Retract", "Syntax", "Module"),
		cmpopts.SortSlices(replaceSortFunc),
	); diff != "" {
		t.Errorf("Replace{} mismatch (-want +got):\n%s", diff)
	}

}
