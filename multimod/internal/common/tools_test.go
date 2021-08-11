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

package common

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.opentelemetry.io/build-tools/multimod/internal/common/commontest"
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
	testName := "update_go_mod_versions"

	tmpRootDir, err := os.MkdirTemp(testDataDir, testName)
	if err != nil {
		t.Fatal("error creating temp dir:", err)
	}

	defer commontest.RemoveAll(t, tmpRootDir)
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

	if err := commontest.WriteTempFiles(modFiles); err != nil {
		t.Fatal("could not create go mod file tree", err)
	}

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

	newModPaths := []ModulePath{
		"go.opentelemetry.io/build-tools/multimod/internal/prerelease/test/test1",
		"go.opentelemetry.io/build-tools/multimod/internal/prerelease/test/test2",
	}
	newVersion := "v1.2.3-RC1+meta"

	err = UpdateGoModFiles(modFilePaths, newModPaths, newVersion)
	require.NoError(t, err)

	for modFilePath, expectedByteOutput := range expectedModFiles {
		actual, err := ioutil.ReadFile(modFilePath)
		require.NoError(t, err)

		assert.Equal(t, expectedByteOutput, actual)
	}
}

func TestFilePathToRegex(t *testing.T) {
	testCases := []struct {
		fpath    string
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
		actual := filePathToRegex(tc.fpath)

		assert.Equal(t, tc.expected, actual)
	}
}
