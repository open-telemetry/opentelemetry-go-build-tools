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
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testDataDir = "./test_data"
)
//
//func TestMockModuleVersioning(t *testing.T) {
//	modSetMap := common.ModuleSetMap{
//		"mod-set-1": common.ModuleSet{
//			Version: "v1.2.3-RC1+meta",
//			Modules: []common.ModulePath{
//				"go.opentelemetry.io/test/test1",
//				"go.opentelemetry.io/test/test2",
//			},
//		},
//		"mod-set-2": common.ModuleSet{
//			Version: "v0.1.0",
//			Modules: []common.ModulePath{
//				"go.opentelemetry.io/test3",
//			},
//		},
//	}
//
//	modPathMap := common.ModulePathMap{
//		"go.opentelemetry.io/test/test1": "root/path/to/mod/test/test1/go.mod",
//		"go.opentelemetry.io/test/test2": "root/path/to/mod/test/test2/go.mod",
//		"go.opentelemetry.io/test3":      "root/test3/go.mod",
//	}
//
//	expected := common.ModuleVersioning{
//		ModSetMap:  modSetMap,
//		ModPathMap: modPathMap,
//		ModInfoMap: common.ModuleInfoMap{
//			"go.opentelemetry.io/test/test1": common.ModuleInfo{
//				ModuleSetName: "mod-set-1",
//				Version:       "v1.2.3-RC1+meta",
//			},
//			"go.opentelemetry.io/test/test2": common.ModuleInfo{
//				ModuleSetName: "mod-set-1",
//				Version:       "v1.2.3-RC1+meta",
//			},
//			"go.opentelemetry.io/test3": common.ModuleInfo{
//				ModuleSetName: "mod-set-2",
//				Version:       "v0.1.0",
//			},
//		},
//	}
//
//	actual, err := MockModuleVersioning(modSetMap, modPathMap)
//
//	require.NoError(t, err)
//	assert.Equal(t, expected, actual)
//}

func TestWriteGoModFiles(t *testing.T) {
	tmpRootDir, err := os.MkdirTemp(testDataDir, "WriteGoModFiles")
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

	if err := WriteGoModFiles(modFiles); err != nil {
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
