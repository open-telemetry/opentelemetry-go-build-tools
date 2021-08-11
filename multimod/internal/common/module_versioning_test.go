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
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.opentelemetry.io/build-tools/multimod/internal/common/commontest"
)

func TestNewModuleVersioning(t *testing.T) {
	tmpRootDir, err := os.MkdirTemp(testDataDir, "NewModuleVersioning")
	if err != nil {
		t.Fatal("creating temp dir:", err)
	}

	defer commontest.RemoveAll(t, tmpRootDir)

	modFiles := map[string][]byte{
		filepath.Join(tmpRootDir, "test", "test1", "go.mod"): []byte("module \"go.opentelemetry.io/test/test1\"\n\ngo 1.16\n\n" +
			"require (\n\t\"go.opentelemetry.io/testroot/v2\" v2.0.0\n)\n"),
		filepath.Join(tmpRootDir, "test", "go.mod"):          []byte("module go.opentelemetry.io/test3\n\ngo 1.16\n"),
		filepath.Join(tmpRootDir, "go.mod"):                  []byte("module go.opentelemetry.io/testroot/v2\n\ngo 1.16\n"),
		filepath.Join(tmpRootDir, "test", "test2", "go.mod"): []byte("module \"go.opentelemetry.io/test/testexcluded\"\n\ngo 1.16\n"),
	}

	if err := commontest.WriteTempFiles(modFiles); err != nil {
		t.Fatal("could not create go mod file tree", err)
	}

	testCases := []struct {
		name                  string
		versioningFilename    string
		repoRoot              string
		shouldError           bool
		expectedModuleSetMap  ModuleSetMap
		expectedModulePathMap ModulePathMap
		expectedModuleInfoMap ModuleInfoMap
	}{
		{
			name:               "valid versioning",
			versioningFilename: filepath.Join(testDataDir, "new_module_versioning/versions_valid.yaml"),
			repoRoot:           tmpRootDir,
			shouldError:        false,
			expectedModuleSetMap: ModuleSetMap{
				"mod-set-1": ModuleSet{
					Version: "v1.2.3-RC1+meta",
					Modules: []ModulePath{
						"go.opentelemetry.io/test/test1",
					},
				},
				"mod-set-2": ModuleSet{
					Version: "v0.1.0",
					Modules: []ModulePath{
						"go.opentelemetry.io/test3",
					},
				},
				"mod-set-3": ModuleSet{
					Version: "v2.2.2",
					Modules: []ModulePath{
						"go.opentelemetry.io/testroot/v2",
					},
				},
			},
			expectedModulePathMap: ModulePathMap{
				"go.opentelemetry.io/test/test1":  ModuleFilePath(filepath.Join(tmpRootDir, "test", "test1", "go.mod")),
				"go.opentelemetry.io/test3":       ModuleFilePath(filepath.Join(tmpRootDir, "test", "go.mod")),
				"go.opentelemetry.io/testroot/v2": ModuleFilePath(filepath.Join(tmpRootDir, "go.mod")),
			},
			expectedModuleInfoMap: ModuleInfoMap{
				"go.opentelemetry.io/test/test1": ModuleInfo{
					ModuleSetName: "mod-set-1",
					Version:       "v1.2.3-RC1+meta",
				},
				"go.opentelemetry.io/testroot/v2": ModuleInfo{
					ModuleSetName: "mod-set-3",
					Version:       "v2.2.2",
				},
				"go.opentelemetry.io/test3": ModuleInfo{
					ModuleSetName: "mod-set-2",
					Version:       "v0.1.0",
				},
			},
		},
		{
			name:                  "invalid version file syntax",
			versioningFilename:    filepath.Join(testDataDir, "new_module_versioning/versions_invalid_syntax.yaml"),
			repoRoot:              tmpRootDir,
			shouldError:           true,
			expectedModuleSetMap:  nil,
			expectedModulePathMap: nil,
			expectedModuleInfoMap: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual, err := NewModuleVersioning(tc.versioningFilename, tc.repoRoot)

			if tc.shouldError {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
			}
			assert.IsType(t, ModuleVersioning{}, actual)
			assert.Equal(t, tc.expectedModuleSetMap, actual.ModSetMap)
			assert.Equal(t, tc.expectedModulePathMap, actual.ModPathMap)
			assert.Equal(t, tc.expectedModuleInfoMap, actual.ModInfoMap)
		})
	}
}
