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

func TestGetGitRepo(t *testing.T) {
	pwd, err := filepath.Abs(".")
	require.NoError(t, err)

	_, err = GetGitRepo(pwd)
	require.NoError(t, err)

	_, err = GetGitRepo("/tmp")
	require.Error(t, err)
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
		f, err := os.Create(filepath.Clean(goMod))
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

	require.NoError(t, os.WriteFile(filepath.Clean(goMod), []byte("invalid file format"), 0600))

	_, err := FindModules(root)
	errList := modfile.ErrorList{}
	require.ErrorAs(t, err, &errList)
	require.Len(t, errList, 1, "unexpected errors")
	assert.EqualError(t, errList[0].Err, "unknown directive: invalid")
}

func TestFindDockerfiles(t *testing.T) {
	root := t.TempDir()
	layout := []struct {
		dir  string
		file string
	}{
		{root, "Dockerfile"},
		{filepath.Join(root, "a/b"), "Dockerfile.test"},
		{filepath.Join(root, "a"), "test.Dockerfile"},
		{filepath.Join(root, "c"), "Dockerfile"},
	}
	for _, path := range layout {
		require.NoError(t, os.MkdirAll(path.dir, os.ModePerm))

		dFile := filepath.Join(path.dir, path.file)
		f, err := os.Create(filepath.Clean(dFile))
		require.NoError(t, err)
		fmt.Fprint(f, "FROM golang:1.19-alpine\n")
		require.NoError(t, f.Close())
	}
	// Add an empty dir.
	require.NoError(t, os.MkdirAll(filepath.Join(root, "tools"), os.ModePerm))

	got, err := FindFilePatternDirs(root, "*Dockerfile*")
	require.NoError(t, err)
	require.Len(t, got, len(layout), "number of found Dockerfile")
	for i, path := range layout {
		assert.Equal(t, filepath.Join(path.dir, path.file), got[i])
	}
}
