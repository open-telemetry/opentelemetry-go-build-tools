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
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.opentelemetry.io/build-tools/multimod/internal/common/commontest"
)

var (
	testDataDir, _ = filepath.Abs("./test_data")
)

func TestUpdateAllVersionGo(t *testing.T) {
	testName := "update_all_version_go"
	versionsYamlDir := filepath.Join(testDataDir, testName)

	versioningFilename := filepath.Join(versionsYamlDir, "versions_valid.yaml")

	tmpRootDir, err := os.MkdirTemp(testDataDir, testName)
	if err != nil {
		t.Fatal("error creating temp dir:", err)
	}

	defer commontest.RemoveAll(t, tmpRootDir)

	modFiles := map[string][]byte{
		filepath.Join(tmpRootDir, "test", "test1", "go.mod"): []byte("module \"go.opentelemetry.io/test/test1\"\n\ngo 1.16\n\n" +
			"require (\n\t\"go.opentelemetry.io/testroot/v2\" v2.2.2\n)\n"),
		filepath.Join(tmpRootDir, "test", "go.mod"):          []byte("module go.opentelemetry.io/test2\n\ngo 1.16\n"),
		filepath.Join(tmpRootDir, "go.mod"):                  []byte("module go.opentelemetry.io/testroot/v2\n\ngo 1.16\n"),
		filepath.Join(tmpRootDir, "test", "test2", "go.mod"): []byte("module \"go.opentelemetry.io/test/testexcluded\"\n\ngo 1.16\n"),
	}

	versionGoFiles := map[string][]byte{
		filepath.Join(tmpRootDir, "test", "test1", "version.go"): []byte("package test1 // import \"go.opentelemetry.io/test/test1\"\n\n" +
			"// Version is the current release version of OpenTelemetry in use.\n" +
			"func Version() string {\n\t" +
			"return \"1.0.0-OLD\"\n" +
			"}\n"),
		filepath.Join(tmpRootDir, "test", "version.go"): []byte("package test2 // import \"go.opentelemetry.io/test/test2\"\n\n" +
			"// version is the current release version of OpenTelemetry in use.\n" +
			"func version() string {\n\t" +
			"return \"0.1.0-OLD\"\n" +
			"}\n"),
	}

	testCases := []struct {
		name                     string
		modSetName               string
		expectedVersionGoOutputs map[string][]byte
	}{
		{
			name:       "update_version_1",
			modSetName: "mod-set-1",
			expectedVersionGoOutputs: map[string][]byte{
				filepath.Join(tmpRootDir, "test", "test1", "version.go"): []byte("package test1 // import \"go.opentelemetry.io/test/test1\"\n\n" +
					"// Version is the current release version of OpenTelemetry in use.\n" +
					"func Version() string {\n\t" +
					"return \"1.2.3-RC1+meta\"\n" +
					"}\n"),
				filepath.Join(tmpRootDir, "test", "version.go"): []byte("package test2 // import \"go.opentelemetry.io/test/test2\"\n\n" +
					"// version is the current release version of OpenTelemetry in use.\n" +
					"func version() string {\n\t" +
					"return \"0.1.0-OLD\"\n" +
					"}\n"),
			},
		},
		{
			name:       "update_version_2",
			modSetName: "mod-set-2",
			expectedVersionGoOutputs: map[string][]byte{
				filepath.Join(tmpRootDir, "test", "test1", "version.go"): []byte("package test1 // import \"go.opentelemetry.io/test/test1\"\n\n" +
					"// Version is the current release version of OpenTelemetry in use.\n" +
					"func Version() string {\n\t" +
					"return \"1.0.0-OLD\"\n" +
					"}\n"),
				filepath.Join(tmpRootDir, "test", "version.go"): []byte("package test2 // import \"go.opentelemetry.io/test/test2\"\n\n" +
					"// version is the current release version of OpenTelemetry in use.\n" +
					"func version() string {\n\t" +
					"return \"0.1.0\"\n" +
					"}\n"),
			},
		},
		{
			name:       "no_version_go_in_set",
			modSetName: "mod-set-3",
			expectedVersionGoOutputs: map[string][]byte{
				filepath.Join(tmpRootDir, "test", "test1", "version.go"): []byte("package test1 // import \"go.opentelemetry.io/test/test1\"\n\n" +
					"// Version is the current release version of OpenTelemetry in use.\n" +
					"func Version() string {\n\t" +
					"return \"1.0.0-OLD\"\n" +
					"}\n"),
				filepath.Join(tmpRootDir, "test", "version.go"): []byte("package test2 // import \"go.opentelemetry.io/test/test2\"\n\n" +
					"// version is the current release version of OpenTelemetry in use.\n" +
					"func version() string {\n\t" +
					"return \"0.1.0-OLD\"\n" +
					"}\n"),
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if err := commontest.WriteTempFiles(modFiles); err != nil {
				t.Fatal("could not create go mod file tree", err)
			}

			if err := commontest.WriteTempFiles(versionGoFiles); err != nil {
				t.Fatal("could not create version.go file tree", err)
			}

			p, err := newPrerelease(versioningFilename, tc.modSetName, tmpRootDir)
			require.NoError(t, err)

			err = p.updateAllVersionGo()
			require.NoError(t, err)

			for versionGoFilePath, expectedByteOutput := range tc.expectedVersionGoOutputs {
				actual, err := ioutil.ReadFile(versionGoFilePath)
				require.NoError(t, err)

				assert.Equal(t, expectedByteOutput, actual)
			}
		})
	}
}
