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

package shared // nolint:revive // keeping generic package name until a proper refactoring is done

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.opentelemetry.io/build-tools/multimod/internal/shared/sharedtest"
)

func TestIsStableVersion(t *testing.T) {
	testCases := []struct {
		Version  string
		Expected bool
	}{
		{Version: "v1.0.0", Expected: true},
		{Version: "v1.2.3", Expected: true},
		{Version: "v1.0.0-RC1", Expected: true},
		{Version: "v1.0.0-RC2+MetaData", Expected: true},
		{Version: "v10.10.10", Expected: true},
		{Version: "v0.0.0", Expected: false},
		{Version: "v0.1.2", Expected: false},
		{Version: "v0.20.0", Expected: false},
		{Version: "v0.0.0-RC1", Expected: false},
		{Version: "not-valid-semver", Expected: false},
	}

	for _, tc := range testCases {
		actual := IsStableVersion(tc.Version)

		assert.Equal(t, tc.Expected, actual)
	}
}

func TestUpdateGoModVersions(t *testing.T) {
	tmpRootDir := t.TempDir()
	modFiles := map[string][]byte{
		filepath.Join(tmpRootDir, "test", "test1", "go.mod"): []byte("module go.opentelemetry.io/build-tools/multimod/internal/prerelease/test/test1\n\n" +
			"go 1.16\n\n" +
			"require (\n\t" +
			"go.opentelemetry.io/build-tools/multimod/internal/prerelease/test/test2 v1.2.3-OLD\n\t" +
			"go.opentelemetry.io/other/test/test1 v1.0.0\n\t" +
			"go.opentelemetry.io/other/testroot/v2 v2.2.2\n" +
			")"),
		filepath.Join(tmpRootDir, "test", "test2", "go.mod"): []byte("module go.opentelemetry.io/build-tools/multimod/internal/prerelease/test/test2\n\n" +
			"go 1.16\n\n" +
			"require (\n\t" +
			"go.opentelemetry.io/build-tools/multimod/internal/prerelease/test/test1 v1.2.3-OLD\n\t" +
			"go.opentelemetry.io/other/test/test1 v1.0.0\n\t" +
			"go.opentelemetry.io/other/testroot/v2 v2.2.2\n" +
			")"),
		filepath.Join(tmpRootDir, "test", "go.mod"): []byte("module go.opentelemetry.io/build-tools/multimod/internal/prerelease/test3\n\n" +
			"go 1.16\n\n" +
			"require (\n\t" +
			"go.opentelemetry.io/build-tools/multimod/internal/prerelease/test/test1 v1.2.3-OLD\n\t" +
			"go.opentelemetry.io/build-tools/multimod/internal/prerelease/test/test2 v1.2.3-OLD\n\t" +
			"go.opentelemetry.io/build-tools/multimod/internal/prerelease/testroot v0.1.0-shouldBe2\n\t" +
			"go.opentelemetry.io/other/test2 v0.1.0\n" +
			")"),
		filepath.Join(tmpRootDir, "go.mod"): []byte("module go.opentelemetry.io/build-tools/multimod/internal/prerelease/testroot\n\n" +
			"go 1.16\n\n" +
			"require (\n\t" +
			"go.opentelemetry.io/build-tools/multimod/internal/prerelease/test/test1 v1.2.3-OLD\n\t" +
			"go.opentelemetry.io/build-tools/multimod/internal/prerelease/test/test2 v1.2.3-OLD\n\t" +
			"go.opentelemetry.io/build-tools/multimod/internal/prerelease/test3 v0.1.0-OLD\n\t" +
			"go.opentelemetry.io/other/test/test1 v1.0.0\n\t" +
			")"),
	}

	require.NoError(t, sharedtest.WriteTempFiles(modFiles), "could not create go mod file tree")

	var modFilePaths []ModuleFilePath
	for modFilePath := range modFiles {
		modFilePaths = append(modFilePaths, ModuleFilePath(modFilePath))
	}

	expectedModFiles := map[string][]byte{
		filepath.Join(tmpRootDir, "test", "test1", "go.mod"): []byte("module go.opentelemetry.io/build-tools/multimod/internal/prerelease/test/test1\n\n" +
			"go 1.16\n\n" +
			"require (\n\t" +
			"go.opentelemetry.io/build-tools/multimod/internal/prerelease/test/test2 v1.2.3-RC1+meta\n\t" +
			"go.opentelemetry.io/other/test/test1 v1.0.0\n\t" +
			"go.opentelemetry.io/other/testroot/v2 v2.2.2\n" +
			")"),
		filepath.Join(tmpRootDir, "test", "test2", "go.mod"): []byte("module go.opentelemetry.io/build-tools/multimod/internal/prerelease/test/test2\n\n" +
			"go 1.16\n\n" +
			"require (\n\t" +
			"go.opentelemetry.io/build-tools/multimod/internal/prerelease/test/test1 v1.2.3-RC1+meta\n\t" +
			"go.opentelemetry.io/other/test/test1 v1.0.0\n\t" +
			"go.opentelemetry.io/other/testroot/v2 v2.2.2\n" +
			")"),
		filepath.Join(tmpRootDir, "test", "go.mod"): []byte("module go.opentelemetry.io/build-tools/multimod/internal/prerelease/test3\n\n" +
			"go 1.16\n\n" +
			"require (\n\t" +
			"go.opentelemetry.io/build-tools/multimod/internal/prerelease/test/test1 v1.2.3-RC1+meta\n\t" +
			"go.opentelemetry.io/build-tools/multimod/internal/prerelease/test/test2 v1.2.3-RC1+meta\n\t" +
			"go.opentelemetry.io/build-tools/multimod/internal/prerelease/testroot v0.1.0-shouldBe2\n\t" +
			"go.opentelemetry.io/other/test2 v0.1.0\n" +
			")"),
		filepath.Join(tmpRootDir, "go.mod"): []byte("module go.opentelemetry.io/build-tools/multimod/internal/prerelease/testroot\n\n" +
			"go 1.16\n\n" +
			"require (\n\t" +
			"go.opentelemetry.io/build-tools/multimod/internal/prerelease/test/test1 v1.2.3-RC1+meta\n\t" +
			"go.opentelemetry.io/build-tools/multimod/internal/prerelease/test/test2 v1.2.3-RC1+meta\n\t" +
			"go.opentelemetry.io/build-tools/multimod/internal/prerelease/test3 v0.1.0-OLD\n\t" +
			"go.opentelemetry.io/other/test/test1 v1.0.0\n\t" +
			")"),
	}

	newVersion := "v1.2.3-RC1+meta"
	newModRefs := []ModuleRef{
		{Path: "go.opentelemetry.io/build-tools/multimod/internal/prerelease/test/test1", Version: newVersion},
		{Path: "go.opentelemetry.io/build-tools/multimod/internal/prerelease/test/test2", Version: newVersion},
	}

	require.NoError(t, UpdateGoModFiles(modFilePaths, newModRefs))
	for modFilePath, expectedByteOutput := range expectedModFiles {
		actual, err := os.ReadFile(filepath.Clean(modFilePath))
		require.NoError(t, err)

		assert.Equal(t, expectedByteOutput, actual)
	}
}

func TestModulePathToRegex(t *testing.T) {
	testCases := []struct {
		fpath    ModulePath
		expected string
	}{
		{
			fpath:    "go.opentelemetry.io/test/test1",
			expected: `go\.opentelemetry\.io\/test\/test1`,
		},
		{
			fpath:    "go.opentelemetry.io/test/hyphen-test1",
			expected: `go\.opentelemetry\.io\/test\/hyphen-test1`,
		},
	}

	for _, tc := range testCases {
		actual := modulePathToRegex(tc.fpath)

		assert.Equal(t, tc.expected, actual)
	}
}

func TestReplaceModVersion(t *testing.T) {
	for _, s := range []struct {
		name     string
		input    []byte
		expected []byte
		err      bool
	}{
		{
			name: "simple",
			input: []byte(`module test
go 1.17

require (
	foo.bar/baz v1.2.3
)
`),
			expected: []byte(`module test
go 1.17

require (
	foo.bar/baz v1.2.4
)
`),
		},
		{
			name: "indirect",
			input: []byte(`module test
go 1.17

require (
	foo.bar/baz v1.2.3 // indirect
)
`),
			expected: []byte(`module test
go 1.17

require (
	foo.bar/baz v1.2.4 // indirect
)
`),
		},
		{
			name: "1.17 style",
			input: []byte(`module test
go 1.17

require (
	bar.baz/quux v0.1.2
)

require (
	foo.bar/baz v1.2.3 // indirect
)
`),
			expected: []byte(`module test
go 1.17

require (
	bar.baz/quux v0.1.2
)

require (
	foo.bar/baz v1.2.4 // indirect
)
`),
		},
	} {
		t.Run(s.name, func(t *testing.T) {
			got, err := replaceModVersion("foo.bar/baz", "v1.2.4", s.input)
			assert.Equal(t, string(s.expected), string(got))
			if s.err {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
