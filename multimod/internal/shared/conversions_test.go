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

package shared

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCombineModuleTagNamesAndVersion(t *testing.T) {
	modTagNames := []ModuleTagName{
		"tag1",
		"tag2",
		"another/tag3",
		RepoRootTag,
	}

	version := "v1.2.3-RC1+meta-RC1"

	expected := []string{
		"tag1/v1.2.3-RC1+meta-RC1",
		"tag2/v1.2.3-RC1+meta-RC1",
		"another/tag3/v1.2.3-RC1+meta-RC1",
		"v1.2.3-RC1+meta-RC1",
	}

	actual := combineModuleTagNamesAndVersion(modTagNames, version)

	assert.Equal(t, expected, actual)
}

func TestModulePathsToTagNames(t *testing.T) {
	modPaths := []ModulePath{
		"go.opentelemetry.io/test/test1",
		"go.opentelemetry.io/test/test2",
		"go.opentelemetry.io/test3",
		"go.opentelemetry.io/root",
	}

	modPathMap := ModulePathMap{
		"go.opentelemetry.io/test/test1": "root/path/to/mod/test/test1/go.mod",
		"go.opentelemetry.io/test/test2": "root/path/to/mod/test/test2/go.mod",
		"go.opentelemetry.io/test3":      "root/test3/go.mod",
		"go.opentelemetry.io/root":       "root/go.mod",
		"go.opentelemetry.io/not-used":   "path/to/mod/not-used/go.mod",
	}

	repoRoot := "root"

	expected := []ModuleTagName{
		"path/to/mod/test/test1",
		"path/to/mod/test/test2",
		"test3",
		RepoRootTag,
	}

	actual, err := ModulePathsToTagNames(modPaths, modPathMap, repoRoot)

	require.NoError(t, err)
	assert.Equal(t, expected, actual)
}

func TestModulePathsToFilePaths(t *testing.T) {
	testCases := []struct {
		name        string
		modPaths    []ModulePath
		modPathMap  ModulePathMap
		shouldError bool
		expected    []ModuleFilePath
	}{
		{
			name: "valid",
			modPaths: []ModulePath{
				"go.opentelemetry.io/test/test1",
				"go.opentelemetry.io/test/test2",
				"go.opentelemetry.io/test3",
				"go.opentelemetry.io/root",
			},
			modPathMap: ModulePathMap{
				"go.opentelemetry.io/test/test1": "root/path/to/mod/test/test1/go.mod",
				"go.opentelemetry.io/test/test2": "root/path/to/mod/test/test2/go.mod",
				"go.opentelemetry.io/test3":      "root/test3/go.mod",
				"go.opentelemetry.io/root":       "root/go.mod",
				"go.opentelemetry.io/not-used":   "path/to/mod/not-used/go.mod",
			},
			shouldError: false,
			expected: []ModuleFilePath{
				"root/path/to/mod/test/test1/go.mod",
				"root/path/to/mod/test/test2/go.mod",
				"root/test3/go.mod",
				"root/go.mod",
			},
		},
		{
			name: "module not in map",
			modPaths: []ModulePath{
				"go.opentelemetry.io/in_map",
				"go.opentelemetry.io/not_in_map",
			},
			modPathMap: ModulePathMap{
				"go.opentelemetry.io/in_map": "root/path/go.mod",
			},
			shouldError: true,
			expected:    []ModuleFilePath{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual, err := modulePathsToFilePaths(tc.modPaths, tc.modPathMap)

			if tc.shouldError {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
			}
			assert.Equal(t, tc.expected, actual)
		})
	}
}

func TestModuleFilePathToTagName(t *testing.T) {
	repoRoot := "root"

	testCases := []struct {
		name        string
		ModFilePath ModuleFilePath
		ShouldError bool
		Expected    ModuleTagName
	}{
		{
			name:        "go mod file in inner dir",
			ModFilePath: "root/path/to/mod/test/test1/go.mod",
			ShouldError: false,
			Expected:    ModuleTagName("path/to/mod/test/test1"),
		},
		{
			name:        "go mod file in root",
			ModFilePath: "root/go.mod",
			ShouldError: false,
			Expected:    RepoRootTag,
		},
		{
			name:        "no go mod in path",
			ModFilePath: "no/go/mod/in/path",
			ShouldError: true,
			Expected:    "",
		},
		{
			name:        "go mod not contained within root",
			ModFilePath: "not/in/root/go.mod",
			ShouldError: true,
			Expected:    "",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual, err := moduleFilePathToTagName(tc.ModFilePath, repoRoot)

			if tc.ShouldError {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.Expected, actual)
			}
		})
	}
}

func TestModuleFilePathsToTagNames(t *testing.T) {
	testCases := []struct {
		name         string
		modFilePaths []ModuleFilePath
		repoRoot     string
		shouldError  bool
		expected     []ModuleTagName
	}{
		{
			name: "valid",
			modFilePaths: []ModuleFilePath{
				"root/path/to/mod/test/test1/go.mod",
				"root/path/to/mod/test/test2/go.mod",
				"root/test3/go.mod",
				"root/go.mod",
			},
			repoRoot:    "root",
			shouldError: false,
			expected: []ModuleTagName{
				"path/to/mod/test/test1",
				"path/to/mod/test/test2",
				"test3",
				RepoRootTag,
			},
		},
		{
			name: "no go mod in path",
			modFilePaths: []ModuleFilePath{
				"no/go/mod/in/path",
			},
			repoRoot:    "root",
			shouldError: true,
			expected:    nil,
		},
		{
			name: "go mod not contained in root",
			modFilePaths: []ModuleFilePath{
				"not/in/root/go.mod",
			},
			repoRoot:    "root",
			shouldError: true,
			expected:    nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual, err := moduleFilePathsToTagNames(tc.modFilePaths, tc.repoRoot)

			if tc.shouldError {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
			}
			assert.Equal(t, tc.expected, actual)
		})
	}
}
