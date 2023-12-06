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
	"fmt"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/mod/modfile"
)

// TestRunVerifyErrors tests the error paths of verify.
func TestRunVerifyErrors(t *testing.T) {
	assert.ErrorIs(t, verify(nil, nil), errNotEnoughArg)
	assert.ErrorContains(t, verify([]string{"", ""}, nil), "only single path argument allowed, received")
}

// TestRunVerify tests the happy path of verify.
func TestRunVerify(t *testing.T) {
	root := filepath.ToSlash(t.TempDir())

	// We need to override the functions that finds the files
	t.Cleanup(func(f func([]string) (string, []*modfile.File, error)) func() {
		return func() { allModsFunc = f }
	}(allModsFunc))
	allModsFunc = func([]string) (string, []*modfile.File, error) {
		return root, []*modfile.File{
			{Syntax: &modfile.FileSyntax{Name: fmt.Sprintf("%s/go.mod", root)}},
			{Syntax: &modfile.FileSyntax{Name: fmt.Sprintf("%s/a/go.mod", root)}},
			{Syntax: &modfile.FileSyntax{Name: fmt.Sprintf("%s/b/go.mod", root)}},
		}, nil
	}

	t.Cleanup(func(f func(string, []string) ([]string, error)) func() {
		return func() { allDockerFunc = f }
	}(allDockerFunc))
	allDockerFunc = func(string, []string) ([]string, error) {
		return []string{
			fmt.Sprintf("%s/", root),
			fmt.Sprintf("%s/a/", root),
			fmt.Sprintf("%s/b/", root),
		}, nil
	}

	t.Cleanup(func(f func(string, []string) ([]string, error)) func() {
		return func() { allPipFunc = f }
	}(allPipFunc))
	allPipFunc = func(string, []string) ([]string, error) {
		return []string{
			fmt.Sprintf("%s/requirements.txt", root),
		}, nil
	}

	// Finally, we need to override the function that finds the updates
	t.Cleanup(func(f func(string) (updates, error)) func() {
		return func() { configuredUpdatesFunc = f }
	}(configuredUpdatesFunc))
	configuredUpdatesFunc = func(string) (updates, error) {
		return updates{
			mods: map[string]struct{}{
				"/":  {},
				"/a": {},
				"/b": {},
			},
			docker: map[string]struct{}{
				"/":  {},
				"/a": {},
				"/b": {},
			},
			pip: map[string]struct{}{
				"/": {},
			},
		}, nil
	}

	// Assert that verify returns no error
	assert.NoError(t, verify([]string{""}, nil))
}

// TestRunVerifyMissingRepo tests when there are missing mods
func TestRunVerifyMissingMods(t *testing.T) {
	root := filepath.ToSlash(t.TempDir())

	t.Cleanup(func(f func([]string) (string, []*modfile.File, error)) func() {
		return func() { allModsFunc = f }
	}(allModsFunc))
	allModsFunc = func([]string) (string, []*modfile.File, error) {
		return root, []*modfile.File{
			{Syntax: &modfile.FileSyntax{Name: fmt.Sprintf("%s/go.mod", root)}},
		}, nil
	}

	t.Cleanup(func(f func(string, []string) ([]string, error)) func() {
		return func() { allDockerFunc = f }
	}(allDockerFunc))
	allDockerFunc = func(string, []string) ([]string, error) {
		return []string{}, nil
	}

	t.Cleanup(func(f func(string, []string) ([]string, error)) func() {
		return func() { allPipFunc = f }
	}(allPipFunc))
	allPipFunc = func(string, []string) ([]string, error) {
		return []string{}, nil
	}

	t.Cleanup(func(f func(string) (updates, error)) func() {
		return func() { configuredUpdatesFunc = f }
	}(configuredUpdatesFunc))
	configuredUpdatesFunc = func(string) (updates, error) {
		return updates{}, nil
	}

	assert.ErrorContains(t, verify([]string{""}, nil), "missing update check(s)")
}

// TestRunVerifyMissingDocker tests when there are missing docker files
func TestRunVerifyMissingDocker(t *testing.T) {
	root := filepath.ToSlash(t.TempDir())

	t.Cleanup(func(f func([]string) (string, []*modfile.File, error)) func() {
		return func() { allModsFunc = f }
	}(allModsFunc))
	allModsFunc = func([]string) (string, []*modfile.File, error) {
		return root, []*modfile.File{}, nil
	}

	t.Cleanup(func(f func(string, []string) ([]string, error)) func() {
		return func() { allDockerFunc = f }
	}(allDockerFunc))
	allDockerFunc = func(string, []string) ([]string, error) {
		return []string{fmt.Sprintf("%s/", root)}, nil
	}

	t.Cleanup(func(f func(string, []string) ([]string, error)) func() {
		return func() { allPipFunc = f }
	}(allPipFunc))
	allPipFunc = func(string, []string) ([]string, error) {
		return []string{}, nil
	}

	t.Cleanup(func(f func(string) (updates, error)) func() {
		return func() { configuredUpdatesFunc = f }
	}(configuredUpdatesFunc))
	configuredUpdatesFunc = func(string) (updates, error) {
		return updates{}, nil
	}

	assert.ErrorContains(t, verify([]string{""}, nil), "missing update check(s)")
}

// TestRunVerifyMissingPip tests when there are missing Pip files
func TestRunVerifyMissingPip(t *testing.T) {
	root := filepath.ToSlash(t.TempDir())

	t.Cleanup(func(f func([]string) (string, []*modfile.File, error)) func() {
		return func() { allModsFunc = f }
	}(allModsFunc))
	allModsFunc = func([]string) (string, []*modfile.File, error) {
		return root, []*modfile.File{}, nil
	}

	t.Cleanup(func(f func(string, []string) ([]string, error)) func() {
		return func() { allDockerFunc = f }
	}(allDockerFunc))
	allDockerFunc = func(string, []string) ([]string, error) {
		return []string{}, nil
	}

	t.Cleanup(func(f func(string, []string) ([]string, error)) func() {
		return func() { allPipFunc = f }
	}(allPipFunc))
	allPipFunc = func(string, []string) ([]string, error) {
		return []string{fmt.Sprintf("%s/requirements.txt", root)}, nil
	}

	t.Cleanup(func(f func(string) (updates, error)) func() {
		return func() { configuredUpdatesFunc = f }
	}(configuredUpdatesFunc))
	configuredUpdatesFunc = func(string) (updates, error) {
		return updates{}, nil
	}

	assert.ErrorContains(t, verify([]string{""}, nil), "missing update check(s)")
}

func TestRunVerifyReturnAllModsError(t *testing.T) {
	t.Cleanup(func(f func([]string) (string, []*modfile.File, error)) func() {
		return func() { allModsFunc = f }
	}(allModsFunc))
	allModsFunc = func([]string) (string, []*modfile.File, error) {
		return "", []*modfile.File{}, assert.AnError
	}
	assert.ErrorIs(t, verify([]string{""}, nil), assert.AnError)
}

func TestRunVerifyReturnAllDockerError(t *testing.T) {
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
		return []string{}, assert.AnError
	}

	assert.ErrorIs(t, verify([]string{""}, nil), assert.AnError)
}

func TestRunVerifyReturnAllPipError(t *testing.T) {
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
		return []string{}, nil
	}

	t.Cleanup(func(f func(string, []string) ([]string, error)) func() {
		return func() { allPipFunc = f }
	}(allPipFunc))
	allPipFunc = func(string, []string) ([]string, error) {
		return []string{}, assert.AnError
	}

	assert.ErrorIs(t, verify([]string{""}, nil), assert.AnError)
}

func TestRunVerifyReturnConfiguredUpdatesError(t *testing.T) {
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
		return []string{}, nil
	}

	t.Cleanup(func(f func(string, []string) ([]string, error)) func() {
		return func() { allPipFunc = f }
	}(allPipFunc))
	allPipFunc = func(string, []string) ([]string, error) {
		return []string{}, nil
	}

	t.Cleanup(func(f func(string) (updates, error)) func() {
		return func() { configuredUpdatesFunc = f }
	}(configuredUpdatesFunc))
	configuredUpdatesFunc = func(string) (updates, error) {
		return updates{}, assert.AnError
	}
	assert.ErrorIs(t, verify([]string{""}, nil), assert.AnError)
}

func TestConfiguredUpdates(t *testing.T) {
	updates, err := configuredUpdates("./testdata/dependabot.yml")
	require.NoError(t, err)

	assert.Equal(t, map[string]struct{}{
		"/":    {},
		"/a":   {},
		"/a/b": {},
	}, updates.mods)
	assert.Equal(t, map[string]struct{}{
		"/":            {},
		"/a/b/example": {},
	}, updates.docker)
	assert.Equal(t, map[string]struct{}{
		"/": {},
	}, updates.pip)
}

func TestConfiguredUpdatesBadPath(t *testing.T) {
	const path = "./testdata/file-does-not-exist"
	_, err := configuredUpdates(path)
	errMsg := fmt.Sprintf("dependabot configuration file does not exist: %s", path)
	assert.EqualError(t, err, errMsg)
}

func TestConfiguredUpdatesInvalidYAML(t *testing.T) {
	_, err := configuredUpdates("./testdata/invalid.yml")
	assert.ErrorContains(t, err, "invalid dependabot configuration")
}
