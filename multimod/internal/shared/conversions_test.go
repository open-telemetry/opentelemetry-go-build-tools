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

func TestCombineModuleTagNamesAndVersionVNSubdirectory(t *testing.T) {
	testCases := []struct {
		name        string
		modTagNames []ModuleTagName
		version     string
		expected    []string
	}{
		{
			name:        "regular modules without vN suffix",
			modTagNames: []ModuleTagName{"foo/bar", "another/module"},
			version:     "v1.2.3",
			expected:    []string{"foo/bar/v1.2.3", "another/module/v1.2.3"},
		},
		{
			name:        "modules with vN in middle of path",
			modTagNames: []ModuleTagName{"foo/v2/bar", "another/v3/module/test"},
			version:     "v1.0.0",
			expected:    []string{"foo/v2/bar/v1.0.0", "another/v3/module/test/v1.0.0"},
		},
		{
			name:        "modules ending with v but not vN pattern",
			modTagNames: []ModuleTagName{"foo/barv", "test/modulev2x"},
			version:     "v1.5.0",
			expected:    []string{"foo/barv/v1.5.0", "test/modulev2x/v1.5.0"},
		},
		{
			name:        "root repo tag behavior unchanged",
			modTagNames: []ModuleTagName{RepoRootTag},
			version:     "v2.0.0",
			expected:    []string{"v2.0.0"},
		},
		{
			name:        "v2 module in subdirectory",
			modTagNames: []ModuleTagName{"detectors/aws/ec2/v2"},
			version:     "v2.0.0",
			expected:    []string{"detectors/aws/ec2/v2.0.0"},
		},
		{
			name:        "v3 module in nested subdirectory",
			modTagNames: []ModuleTagName{"foo/bar/baz/v3"},
			version:     "v3.1.2",
			expected:    []string{"foo/bar/baz/v3.1.2"},
		},
		{
			name:        "root level vN module creates direct tag",
			modTagNames: []ModuleTagName{"v5"},
			version:     "v5.0.0",
			expected:    []string{"v5.0.0"},
		},
		{
			name:        "v9 module in simple subdirectory",
			modTagNames: []ModuleTagName{"module/v9"},
			version:     "v9.2.1",
			expected:    []string{"module/v9.2.1"},
		},
		{
			name: "mixed v2 and regular modules",
			modTagNames: []ModuleTagName{
				"detectors/aws/ec2/v2",
				"detectors/aws/ecs",
				"contrib/v3",
			},
			version: "v2.0.0",
			expected: []string{
				"detectors/aws/ec2/v2.0.0",
				"detectors/aws/ecs/v2.0.0",
				"contrib/v3/v2.0.0", // v3 module but v2.0.0 version - should use regular behavior
			},
		},
		{
			name:        "exact issue case - detectors/aws/ec2/v2 with v2.0.0",
			modTagNames: []ModuleTagName{"detectors/aws/ec2/v2"},
			version:     "v2.0.0",
			expected:    []string{"detectors/aws/ec2/v2.0.0"},
		},
		{
			name:        "version mismatch should use regular behavior",
			modTagNames: []ModuleTagName{"foo/bar/v3"},
			version:     "v2.1.0", // v3 module but v2 version
			expected:    []string{"foo/bar/v3/v2.1.0"},
		},
		{
			name:        "root level version mismatch uses regular behavior",
			modTagNames: []ModuleTagName{"v3"},
			version:     "v2.0.0", // v3 module but v2 version
			expected:    []string{"v3/v2.0.0"},
		},
		{
			name:        "multi-digit version mismatch uses regular behavior",
			modTagNames: []ModuleTagName{"foo/bar/v10"},
			version:     "v11.0.0", // v10 module but v11 version
			expected:    []string{"foo/bar/v10/v11.0.0"},
		},
		{
			name:        "v0 and v1 modules use regular behavior",
			modTagNames: []ModuleTagName{"foo/bar/v0", "foo/bar/v1"},
			version:     "v0.1.0",
			expected:    []string{"foo/bar/v0/v0.1.0", "foo/bar/v1/v0.1.0"}, // v0/v1 don't use vN subdirectory convention
		},
		{
			name:        "double digit version numbers are supported",
			modTagNames: []ModuleTagName{"foo/bar/v10"},
			version:     "v10.0.0",
			expected:    []string{"foo/bar/v10.0.0"}, // multi-digit vN now supported
		},
		{
			name:        "root level v2 module",
			modTagNames: []ModuleTagName{"v2"},
			version:     "v2.1.5",
			expected:    []string{"v2.1.5"},
		},
		{
			name:        "root level v9 module",
			modTagNames: []ModuleTagName{"v9"},
			version:     "v9.0.0-beta.1",
			expected:    []string{"v9.0.0-beta.1"},
		},
		{
			name:        "root level v10 module",
			modTagNames: []ModuleTagName{"v10"},
			version:     "v10.1.0",
			expected:    []string{"v10.1.0"},
		},
		{
			name:        "root level v123 module",
			modTagNames: []ModuleTagName{"v123"},
			version:     "v123.0.0",
			expected:    []string{"v123.0.0"},
		},
		{
			name:        "subdirectory v11 module",
			modTagNames: []ModuleTagName{"contrib/instrumentation/v11"},
			version:     "v11.2.3",
			expected:    []string{"contrib/instrumentation/v11.2.3"},
		},
		{
			name:        "nested subdirectory v100 module",
			modTagNames: []ModuleTagName{"foo/bar/baz/qux/v100"},
			version:     "v100.0.0-rc.1",
			expected:    []string{"foo/bar/baz/qux/v100.0.0-rc.1"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual := combineModuleTagNamesAndVersion(tc.modTagNames, tc.version)
			assert.Equal(t, tc.expected, actual)
		})
	}
}
