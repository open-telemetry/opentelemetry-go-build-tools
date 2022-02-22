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
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/mod/modfile"
)

func TestRunVerifyErrors(t *testing.T) {
	assert.ErrorIs(t, verify(nil), errNotEnoughArg)
	assert.ErrorIs(t, verify([]string{"", ""}), errTooManyArg)
}

func TestRunVerify(t *testing.T) {
	t.Cleanup(func(f func() (string, []*modfile.File, error)) func() {
		return func() { allModsFunc = f }
	}(allModsFunc))
	allModsFunc = func() (string, []*modfile.File, error) {
		return "/home/user/repo", []*modfile.File{
			{Syntax: &modfile.FileSyntax{Name: "/home/user/repo/go.mod"}},
			{Syntax: &modfile.FileSyntax{Name: "/home/user/repo/a/go.mod"}},
			{Syntax: &modfile.FileSyntax{Name: "/home/user/repo/b/go.mod"}},
		}, nil
	}

	t.Cleanup(func(f func(string) (map[string]struct{}, error)) func() {
		return func() { configuredUpdatesFunc = f }
	}(configuredUpdatesFunc))
	configuredUpdatesFunc = func(string) (map[string]struct{}, error) {
		return map[string]struct{}{
			"/":  struct{}{},
			"/a": struct{}{},
			"/b": struct{}{},
		}, nil
	}

	assert.NoError(t, verify([]string{""}))
}

func TestRunVerifyMissing(t *testing.T) {
	t.Cleanup(func(f func() (string, []*modfile.File, error)) func() {
		return func() { allModsFunc = f }
	}(allModsFunc))
	allModsFunc = func() (string, []*modfile.File, error) {
		return "/home/user/repo", []*modfile.File{
			{Syntax: &modfile.FileSyntax{Name: "/home/user/repo/go.mod"}},
		}, nil
	}

	t.Cleanup(func(f func(string) (map[string]struct{}, error)) func() {
		return func() { configuredUpdatesFunc = f }
	}(configuredUpdatesFunc))
	configuredUpdatesFunc = func(string) (map[string]struct{}, error) {
		return map[string]struct{}{}, nil
	}

	assert.ErrorIs(t, verify([]string{""}), errMissing)
}

func TestRunVerifyReturnAllModsError(t *testing.T) {
	t.Cleanup(func(f func() (string, []*modfile.File, error)) func() {
		return func() { allModsFunc = f }
	}(allModsFunc))
	allModsFunc = func() (string, []*modfile.File, error) {
		return "", []*modfile.File{}, assert.AnError
	}
	assert.ErrorIs(t, verify([]string{""}), assert.AnError)
}

func TestRunVerifyReturnConfiguredUpdatesError(t *testing.T) {
	t.Cleanup(func(f func() (string, []*modfile.File, error)) func() {
		return func() { allModsFunc = f }
	}(allModsFunc))
	allModsFunc = func() (string, []*modfile.File, error) {
		return "", []*modfile.File{}, nil
	}

	t.Cleanup(func(f func(string) (map[string]struct{}, error)) func() {
		return func() { configuredUpdatesFunc = f }
	}(configuredUpdatesFunc))
	configuredUpdatesFunc = func(string) (map[string]struct{}, error) {
		return map[string]struct{}{}, assert.AnError
	}
	assert.ErrorIs(t, verify([]string{""}), assert.AnError)
}

func TestConfiguredUpdates(t *testing.T) {
	updates, err := configuredUpdates("./testdata/dependabot.yml")
	require.NoError(t, err)

	assert.Equal(t, map[string]struct{}{
		"/":    struct{}{},
		"/a":   struct{}{},
		"/a/b": struct{}{},
	}, updates)
}

func TestConfiguredUpdatesBadPath(t *testing.T) {
	const path = "./testdata/file-does-not-exist"
	_, err := configuredUpdates(path)
	errMsg := fmt.Sprintf("dependabot configuration file does not exist: %s", path)
	assert.EqualError(t, err, errMsg)
}

func TestConfiguredUpdatesInvalidYAML(t *testing.T) {
	_, err := configuredUpdates("./testdata/invalid.yml")
	assert.ErrorIs(t, err, errInvalid)
}
