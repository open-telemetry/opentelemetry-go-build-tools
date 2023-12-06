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
	"fmt"
	"io"
	"path/filepath"
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
	require.NoError(t, generate(nil))

	got := b.String()
	assert.True(t, strings.HasPrefix(got, header), "missing header")
}

func TestRunGenerateYAML(t *testing.T) {
	// Should output parsable yaml.

	var b bytes.Buffer
	t.Cleanup(func(w io.Writer) func() { return func() { output = w } }(output))
	output = &b
	require.NoError(t, generate(nil))

	var c dependabotConfig
	assert.NoError(t, yaml.NewDecoder(&b).Decode(&c))
}

func newUpdate(pkgEco, dir string, labels []string) update {
	return update{
		PackageEcosystem: pkgEco,
		Directory:        dir,
		Labels:           labels,
		Schedule:         weeklySchedule,
	}
}

func TestBuildConfig(t *testing.T) {
	root := filepath.ToSlash(t.TempDir())

	mods := []*modfile.File{
		{Syntax: &modfile.FileSyntax{Name: fmt.Sprintf("%s/go.mod", root)}},
		{Syntax: &modfile.FileSyntax{Name: fmt.Sprintf("%s/a/go.mod", root)}},
		{Syntax: &modfile.FileSyntax{Name: fmt.Sprintf("%s/b/go.mod", root)}},
	}
	dockerFiles := []string{
		fmt.Sprintf("%s/", root),
		fmt.Sprintf("%s/a/", root),
		fmt.Sprintf("%s/b/", root),
	}
	pipFiles := []string{
		fmt.Sprintf("%s/requirements.txt", root),
	}

	got, err := buildConfig(root, mods, dockerFiles, pipFiles)
	require.NoError(t, err)
	assert.Equal(t, &dependabotConfig{
		Version: version2,
		Updates: []update{
			newUpdate(ghPkgEco, "/", actionLabels),
			newUpdate(dockerPkgEco, "/", dockerLabels),
			newUpdate(dockerPkgEco, "/a", dockerLabels),
			newUpdate(dockerPkgEco, "/b", dockerLabels),
			newUpdate(gomodPkgEco, "/", goLabels),
			newUpdate(gomodPkgEco, "/a", goLabels),
			newUpdate(gomodPkgEco, "/b", goLabels),
			newUpdate(pipPkgEco, "/", pipLabels),
		},
	}, got)
}

func TestRunGenerateReturnAllModsError(t *testing.T) {
	t.Cleanup(func(f func([]string) (string, []*modfile.File, error)) func() {
		return func() { allModsFunc = f }
	}(allModsFunc))
	allModsFunc = func([]string) (string, []*modfile.File, error) {
		return "", []*modfile.File{}, assert.AnError
	}
	assert.ErrorIs(t, generate(nil), assert.AnError)
}

func TestRunGenerateReturnAllDockerError(t *testing.T) {
	t.Cleanup(func(f func(string, []string) ([]string, error)) func() {
		return func() { allDockerFunc = f }
	}(allDockerFunc))
	allDockerFunc = func(string, []string) ([]string, error) {
		return nil, assert.AnError
	}
	assert.ErrorIs(t, generate(nil), assert.AnError)
}

func TestRunGenerateReturnBuildConfigError(t *testing.T) {
	t.Cleanup(func(f func([]string) (string, []*modfile.File, error)) func() {
		return func() { allModsFunc = f }
	}(allModsFunc))
	allModsFunc = func([]string) (string, []*modfile.File, error) {
		return "", []*modfile.File{}, nil
	}
	t.Cleanup(func(f func(string, []string) ([]string, error)) func() {
		return func() { allDockerFunc = f }
	}(allDockerFunc))
	allDockerFunc = func(string, []string) ([]string, error) {
		return nil, nil
	}
	t.Cleanup(func(f func(string, []string) ([]string, error)) func() {
		return func() { allPipFunc = f }
	}(allPipFunc))
	allPipFunc = func(string, []string) ([]string, error) {
		return nil, nil
	}

	t.Cleanup(func(f func(string, []*modfile.File, []string, []string) (*dependabotConfig, error)) func() {
		return func() { buildConfigFunc = f }
	}(buildConfigFunc))
	buildConfigFunc = func(string, []*modfile.File, []string, []string) (*dependabotConfig, error) {
		return nil, assert.AnError
	}
	assert.ErrorIs(t, generate(nil), assert.AnError)
}
