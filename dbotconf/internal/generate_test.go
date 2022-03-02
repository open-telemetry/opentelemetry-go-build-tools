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

package internal

import (
	"bytes"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/mod/modfile"
	"gopkg.in/yaml.v3"
)

func TestRunGenerateHeader(t *testing.T) {
	var b bytes.Buffer
	t.Cleanup(func(w io.Writer) func() { return func() { output = w } }(output))
	output = &b
	require.NoError(t, generate())

	got := b.String()
	assert.True(t, strings.HasPrefix(got, header), "missing header")
}

func TestRunGenerateYAML(t *testing.T) {
	// Should output parsable yaml.

	var b bytes.Buffer
	t.Cleanup(func(w io.Writer) func() { return func() { output = w } }(output))
	output = &b
	require.NoError(t, generate())

	var c dependabotConfig
	assert.NoError(t, yaml.NewDecoder(&b).Decode(&c))
}

func newUpdate(pkgEco, dir string) update {
	return update{
		PackageEcosystem: pkgEco,
		Directory:        dir,
		Labels:           labels,
		Schedule:         weeklySchedule,
	}
}

func TestBuildConfig(t *testing.T) {
	root := "/home/user/repo"
	mods := []*modfile.File{
		{Syntax: &modfile.FileSyntax{Name: "/home/user/repo/go.mod"}},
		{Syntax: &modfile.FileSyntax{Name: "/home/user/repo/a/go.mod"}},
		{Syntax: &modfile.FileSyntax{Name: "/home/user/repo/b/go.mod"}},
	}

	got, err := buildConfig(root, mods)
	require.NoError(t, err)
	assert.Equal(t, &dependabotConfig{
		Version: version2,
		Updates: []update{
			newUpdate(ghPkgEco, "/"),
			newUpdate(gomodPkgEco, "/"),
			newUpdate(gomodPkgEco, "/a"),
			newUpdate(gomodPkgEco, "/b"),
		},
	}, got)
}

func TestRunGenerateReturnAllModsError(t *testing.T) {
	t.Cleanup(func(f func() (string, []*modfile.File, error)) func() {
		return func() { allModsFunc = f }
	}(allModsFunc))
	allModsFunc = func() (string, []*modfile.File, error) {
		return "", []*modfile.File{}, assert.AnError
	}
	assert.ErrorIs(t, generate(), assert.AnError)
}

func TestRunGenerateReturnBuildConfigError(t *testing.T) {
	t.Cleanup(func(f func() (string, []*modfile.File, error)) func() {
		return func() { allModsFunc = f }
	}(allModsFunc))
	allModsFunc = func() (string, []*modfile.File, error) {
		return "", []*modfile.File{}, nil
	}

	t.Cleanup(func(f func(string, []*modfile.File) (*dependabotConfig, error)) func() {
		return func() { buildConfigFunc = f }
	}(buildConfigFunc))
	buildConfigFunc = func(string, []*modfile.File) (*dependabotConfig, error) {
		return nil, assert.AnError
	}
	assert.ErrorIs(t, generate(), assert.AnError)
}
