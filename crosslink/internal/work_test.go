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

func TestWorkUpdate(t *testing.T) {
	lg, _ := zap.NewDevelopment()
	config := RunConfig{Logger: lg}

	mockDir := "testWork"
	want := `go 1.19
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
			replace foo.opentelemetery.io/bar => ../bar`

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

	err = Work(config)
	require.NoError(t, err)
	assertGoWork(t, want, tmpRootDir)
}

func TestWorkNew(t *testing.T) {
	lg, _ := zap.NewDevelopment()
	config := RunConfig{Logger: lg, GoVersion: "1.20"}

	mockDir := "testWork"
	want := `go 1.20
			use ./
			use ./testA
			use ./testB`

	tmpRootDir, err := createTempTestDir(mockDir)
	if err != nil {
		t.Fatal("creating temp dir:", err)
	}

	// remove the go.work to make sure new one gets created
	err = os.Remove(filepath.Join(tmpRootDir, "go.work"))
	require.NoError(t, err)

	err = renameGoMod(tmpRootDir)
	if err != nil {
		t.Errorf("error renaming gomod files: %v", err)
	}
	t.Cleanup(func() { os.RemoveAll(tmpRootDir) })

	config.RootPath = tmpRootDir

	err = Work(config)
	require.NoError(t, err)
	assertGoWork(t, want, tmpRootDir)
}

func assertGoWork(t *testing.T, expected string, tmpRootDir string) {
	t.Helper()

	goWorkContent, err := os.ReadFile(filepath.Clean(filepath.Join(tmpRootDir, "go.work")))
	require.NoError(t, err)

	actual, err := modfile.ParseWork("go.work", goWorkContent, nil)
	require.NoError(t, err)
	actual.Cleanup()

	want, err := modfile.ParseWork("go.work", []byte(expected), nil)
	require.NoError(t, err)
	want.Cleanup()

	// replace structs need to be assorted to avoid flaky fails in test
	replaceSortFunc := func(x, y *modfile.Replace) bool {
		return x.Old.Path < y.Old.Path
	}

	// use structs need to be assorted to avoid flaky fails in test
	useSortFunc := func(x, y *modfile.Use) bool {
		return x.Path < y.Path
	}

	if diff := cmp.Diff(want, actual,
		cmpopts.IgnoreFields(modfile.Use{}, "Syntax", "ModulePath"),
		cmpopts.IgnoreFields(modfile.Replace{}, "Syntax"),
		cmpopts.IgnoreFields(modfile.WorkFile{}, "Syntax"),
		cmpopts.SortSlices(replaceSortFunc),
		cmpopts.SortSlices(useSortFunc),
	); diff != "" {
		t.Errorf("go.work mismatch (-want +got):\n%s", diff)
	}
}
