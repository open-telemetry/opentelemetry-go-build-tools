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
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"golang.org/x/mod/modfile"
)

func TestWork(t *testing.T) {
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
			// new statement added by crosslink
			use ./
			// existing valid use statements under root should remain
			use ./testA
		
			// new statement added by crossling
			use ./testB
			
			// invalid use statements under root should be removed
			// use ./testC
			
			// use statements outside the root should remain
			use ../other-module
			
			// replace statements should remain
			replace foo.opentelemetery.io/bar => ../bar`,
		},
		{
			testName: "excluded",
			config: RunConfig{Logger: lg, ExcludedPaths: map[string]struct{}{
				"go.opentelemetry.io/build-tools/crosslink/testroot/testB": {},
				"go.opentelemetry.io/build-tools/crosslink/testroot/testC": {},
			}},
			expected: `go 1.19
			// new statement added by crosslink
			use ./
			// existing valid use statements under root should remain
			use ./testA

			// do not add EXCLUDED modules
			// use ./testB
			
			// do not add remove EXCLUDED modules
			use ./testC
			
			// use statements outside the root should remain
			use ../other-module
			
			// replace statements should remain
			replace foo.opentelemetery.io/bar => ../bar`,
		},
	}

	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			mockDir := "testWork"
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

			err = Work(test.config)
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
