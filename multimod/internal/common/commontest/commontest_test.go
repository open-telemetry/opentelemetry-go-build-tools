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

package commontest

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	testDataDir, _ = filepath.Abs("./test_data")
)

// TestMain performs setup for the tests and suppress printing logs.
func TestMain(m *testing.M) {
	log.SetOutput(ioutil.Discard)
	os.Exit(m.Run())
}

func TestWriteTempFiles(t *testing.T) {
	tmpRootDir, err := os.MkdirTemp(testDataDir, "WriteTempFiles")
	if err != nil {
		t.Fatal("creating temp dir:", err)
	}

	defer RemoveAll(t, tmpRootDir)

	modFiles := map[string][]byte{
		filepath.Join(tmpRootDir, "test", "test1", "go.mod"): []byte("module \"go.opentelemetry.io/test/test1\"\n\ngo 1.16\n\n" +
		"require (\n\t\"go.opentelemetry.io/testroot/v2\" v2.0.0\n)\n"),
		filepath.Join(tmpRootDir, "test", "go.mod"):          []byte("module go.opentelemetry.io/test3\n\ngo 1.16\n"),
		filepath.Join(tmpRootDir, "go.mod"):                  []byte("module go.opentelemetry.io/testroot/v2\n\ngo 1.16\n"),
		filepath.Join(tmpRootDir, "test", "test2", "go.mod"): []byte("module \"go.opentelemetry.io/test/testexcluded\"\n\ngo 1.16\n"),
	}

	if err := WriteTempFiles(modFiles); err != nil {
		t.Fatal("could not create go mod file tree", err)
	}

	// check all mod files have been written correctly
	for modPath, expectedModFile := range modFiles {
		actual, err := ioutil.ReadFile(modPath)

		require.NoError(t, err)
		assert.Equal(t, expectedModFile, actual)
	}
}

func TestRemoveAll(t *testing.T) {
	tmpRootDir, err := os.MkdirTemp(testDataDir, "RemoveAll")
	if err != nil {
		t.Fatal("creating temp dir:", err)
	}

	tmpNestedDir, err := os.MkdirTemp(tmpRootDir, "RemoveAllNested")
	if err != nil {
		t.Fatal("creating temp dir:", err)
	}

	RemoveAll(t, tmpRootDir)

	_, rootStatErr := os.Stat(tmpRootDir)
	assert.True(t, os.IsNotExist(rootStatErr))
	_, nestedStatErr := os.Stat(tmpNestedDir)
	assert.True(t, os.IsNotExist(nestedStatErr))
}
