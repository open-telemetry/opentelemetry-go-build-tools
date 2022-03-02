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

package tools

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/mod/modfile"
)

func TestFindRepoRoot(t *testing.T) {
	expected, err := filepath.Abs("./")
	require.NoError(t, err)

	actual, err := FindRepoRoot()
	require.NoError(t, err)
	assert.Equal(t, expected, actual)
}

func TestFindModules(t *testing.T) {
	root := t.TempDir()
	dirs := []string{
		root,
		filepath.Join(root, "a"),
		filepath.Join(root, "a/b"),
		filepath.Join(root, "c"),
	}
	for _, d := range dirs {
		require.NoError(t, os.MkdirAll(d, os.ModePerm))
		goMod := filepath.Join(d, "go.mod")
		f, err := os.Create(goMod)
		require.NoError(t, err)
		modName := strings.Replace(d, root, "fake.multi.mod.project", 1)
		fmt.Fprintf(f, "module %s\n", modName)
		require.NoError(t, f.Close())
	}
	// Add a non-module dir.
	require.NoError(t, os.MkdirAll(filepath.Join(root, "tools"), os.ModePerm))

	got, err := FindModules(root)
	require.NoError(t, err)
	require.Len(t, got, len(dirs), "number of found modules")
	for i, d := range dirs {
		assert.Equal(t, filepath.Join(d, "go.mod"), got[i].Syntax.Name)
	}
}

func TestFindModulesReturnsErrorForInvalidGoModFile(t *testing.T) {
	root := t.TempDir()
	goMod := filepath.Join(root, "go.mod")
	f, err := os.Create(goMod)
	require.NoError(t, err)
	fmt.Fprintln(f, "invalid file format")
	require.NoError(t, f.Close())

	_, err = FindModules(root)
	require.IsType(t, modfile.ErrorList{}, err)
	errList := err.(modfile.ErrorList)
	require.Len(t, errList, 1, "unexpected errors")
	assert.EqualError(t, errList[0].Err, "unknown directive: invalid")
}
