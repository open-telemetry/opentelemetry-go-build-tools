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

package repo

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
	expected, err := filepath.Abs("./../..")
	require.NoError(t, err)

	actual, err := FindRoot()
	require.NoError(t, err)
	assert.Equal(t, expected, actual)
}

func prepend(root string, paths ...string) []string {
	out := make([]string, len(paths))
	for i, p := range paths {
		out[i] = filepath.Join(root, p)
	}
	return out
}

func setupGoMod(t *testing.T, dirs []string) string {
	t.Helper()

	root := t.TempDir()

	paths := append([]string{root}, prepend(root, dirs...)...)
	for _, d := range paths {
		require.NoError(t, os.MkdirAll(d, os.ModePerm))
		goMod := filepath.Join(d, "go.mod")
		f, err := os.Create(filepath.Clean(goMod))
		require.NoError(t, err)
		modName := strings.Replace(d, root, "fake.multi.mod.project", 1)
		fmt.Fprintf(f, "module %s\n", modName)
		require.NoError(t, f.Close())
	}
	return root
}

func TestFindModules(t *testing.T) {
	dirs := []string{"a", "a/b", "c"}
	root := setupGoMod(t, dirs)
	// Add a non-module dir.
	require.NoError(t, os.MkdirAll(filepath.Join(root, "tools"), os.ModePerm))

	got, err := FindModules(root, nil)
	require.NoError(t, err)
	require.Len(t, got, len(dirs)+1, "number of found modules")
	for i, d := range append([]string{root}, prepend(root, dirs...)...) {
		want := filepath.Join(d, "go.mod")
		assert.Equal(t, want, got[i].Syntax.Name)
	}
}

func TestFindModulesIgnore(t *testing.T) {
	dirs := []string{"a", "a/b", "a/b/c", "a/b/c/d", "aa", "aa/b0", "aa/b1", "aa/b2", "c"}
	root := setupGoMod(t, dirs)

	got, err := FindModules(root, []string{"a/b", "aa/b?", "c"})
	require.NoError(t, err)
	require.Len(t, got, 3, "number of found modules")
	for i, d := range append([]string{root}, prepend(root, "a", "aa")...) {
		want := filepath.Join(d, "go.mod")
		assert.Equal(t, want, got[i].Syntax.Name)
	}
}

func TestFindModulesReturnsErrorForInvalidGoModFile(t *testing.T) {
	root := t.TempDir()
	goMod := filepath.Join(root, "go.mod")

	require.NoError(t, os.WriteFile(filepath.Clean(goMod), []byte("invalid file format"), 0600))

	_, err := FindModules(root, nil)
	errList := modfile.ErrorList{}
	require.ErrorAs(t, err, &errList)
	require.Len(t, errList, 1, "unexpected errors")
	assert.EqualError(t, errList[0].Err, "unknown directive: invalid")
}

type fPath struct {
	dir  string
	file string
}

func setupDocker(t *testing.T, layout []*fPath) string {
	root := t.TempDir()
	for i, fp := range layout {
		layout[i].dir = prepend(root, fp.dir)[0]

	}
	for _, path := range layout {
		require.NoError(t, os.MkdirAll(path.dir, os.ModePerm))

		dFile := filepath.Join(path.dir, path.file)
		f, err := os.Create(filepath.Clean(dFile))
		require.NoError(t, err)
		fmt.Fprint(f, "FROM golang:1.20-alpine\n")
		require.NoError(t, f.Close())
	}
	return root
}

func TestFindDockerfiles(t *testing.T) {
	layout := []*fPath{
		{"", "Dockerfile"},
		{"a/b", "Dockerfile.test"},
		{"a", "test.Dockerfile"},
		{"c", "Dockerfile"},
	}
	root := setupDocker(t, layout)
	// Add an empty dir.
	require.NoError(t, os.MkdirAll(filepath.Join(root, "tools"), os.ModePerm))

	got, err := FindFilePatternDirs(root, "*Dockerfile*", nil)
	require.NoError(t, err)
	require.Len(t, got, len(layout), "number of found Dockerfile")
	for i, path := range layout {
		assert.Equal(t, filepath.Join(path.dir, path.file), got[i])
	}
}

func TestFindDockerfilesIgnore(t *testing.T) {
	layout := []*fPath{
		{"", "Dockerfile"},
		{"a", "Dockerfile"},
		{"aa", "Dockerfile"},
		{"a/b", "Dockerfile"},
		{"a/b/c", "Dockerfile"},
		{"a/b/c/d", "Dockerfile"},
		{"aa/b0", "Dockerfile"},
		{"aa/b1", "Dockerfile"},
		{"aa/b2", "Dockerfile"},
		{"c", "Dockerfile"},
	}
	root := setupDocker(t, layout)
	// Add an empty dir.
	require.NoError(t, os.MkdirAll(filepath.Join(root, "tools"), os.ModePerm))

	got, err := FindFilePatternDirs(root, "*Dockerfile*", []string{"a/b", "aa/b?", "c"})
	require.NoError(t, err)
	require.Len(t, got, 3, "number of found Dockerfile")
	want := []*fPath{layout[0], layout[1], layout[2]}
	for i, path := range want {
		assert.Equal(t, filepath.Join(path.dir, path.file), got[i])
	}
}
