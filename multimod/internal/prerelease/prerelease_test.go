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

package prerelease

import (
	"log"
	"testing"

	"github.com/stretchr/testify/assert"

	"go.opentelemetry.io/build-tools/releaser/internal/common"
	"go.opentelemetry.io/build-tools/releaser/internal/common/commontest"
)

// MockPrerelease creates a prerelease struct for testing purposes.
func MockPrerelease(modSetMap common.ModuleSetMap, modPathMap common.ModulePathMap, modSetToUpdate string, repoRoot string) prerelease {
	modSetRelease, err := commontest.MockModuleSetRelease(modSetMap, modPathMap, modSetToUpdate, repoRoot)
	if err != nil {
		log.Printf("error getting MockModuleVersioning: %v", err)
		return prerelease{}
	}

	return prerelease{
		ModuleSetRelease: modSetRelease,
	}
}

// Positive-only test
func TestMockPrerelease(t *testing.T) {
	modSetMap := common.ModuleSetMap{
		"mod-set-1": common.ModuleSet{
			Version: "v1.2.3-RC1+meta",
			Modules: []common.ModulePath{
				"go.opentelemetry.io/test/test1",
				"go.opentelemetry.io/test/test2",
			},
		},
		"mod-set-2": common.ModuleSet{
			Version: "v0.1.0",
			Modules: []common.ModulePath{
				"go.opentelemetry.io/test3",
			},
		},
	}

	modPathMap := common.ModulePathMap{
		"go.opentelemetry.io/test/test1": "root/path/to/mod/test/test1/go.mod",
		"go.opentelemetry.io/test/test2": "root/path/to/mod/test/test2/go.mod",
		"go.opentelemetry.io/test3":      "root/test3/go.mod",
	}

	modSetName := "mod-set-1"

	expected := prerelease{
		ModuleSetRelease: common.ModuleSetRelease{
			ModuleVersioning: common.ModuleVersioning{
				ModSetMap:  modSetMap,
				ModPathMap: modPathMap,
				ModInfoMap: common.ModuleInfoMap{
					"go.opentelemetry.io/test/test1": common.ModuleInfo{
						ModuleSetName: "mod-set-1",
						Version:       "v1.2.3-RC1+meta",
					},
					"go.opentelemetry.io/test/test2": common.ModuleInfo{
						ModuleSetName: "mod-set-1",
						Version:       "v1.2.3-RC1+meta",
					},
					"go.opentelemetry.io/test3": common.ModuleInfo{
						ModuleSetName: "mod-set-2",
						Version:       "v0.1.0",
					},
				},
			},
			ModSetName: modSetName,
			ModSet:     modSetMap[modSetName],
			TagNames: []common.ModuleTagName{
				"path/to/mod/test/test1",
				"path/to/mod/test/test2",
			},
		},
	}

	actual := MockPrerelease(modSetMap, modPathMap, modSetName, "root")

	assert.IsType(t, prerelease{}, actual)
	assert.Equal(t, expected, actual)
}

//func TestVerifyGitTagsDoNotAlreadyExist(t *testing.T) {
//
//}
//
//func TestVerifyWorkingTreeClean(t *testing.T) {
//
//}
//
//func TestCreatePrereleaseBranch(t *testing.T) {
//
//}
//
//func TestUpdateVersionGo(t *testing.T) {
//
//}
//
//func TestRunMakeLint(t *testing.T) {
//
//}
//
//func TestCommitChanges(t *testing.T) {
//
//}
//
//func TestUpdateGoModVersions(t *testing.T) {
//
//}
//
//func TestUpdateAllGoModFiles(t *testing.T) {
//
//}

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
