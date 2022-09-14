// Copyright The OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package crosslink

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBuildDependencyGraph(t *testing.T) {

	tests := []struct {
		testName string
		mockDir  string
		config   RunConfig
		expected map[string][]string
	}{
		{
			testName: "testSimple",
			mockDir:  "testSimple",
			config:   DefaultRunConfig(),
			expected: map[string][]string{
				"go.opentelemetry.io/build-tools/crosslink/testroot": {
					"go.opentelemetry.io/build-tools/crosslink/testroot/testA",
					"go.opentelemetry.io/build-tools/crosslink/testroot/testB"},
				"go.opentelemetry.io/build-tools/crosslink/testroot/testA": {"go.opentelemetry.io/build-tools/crosslink/testroot/testB"},
				"go.opentelemetry.io/build-tools/crosslink/testroot/testB": {},
			},
		},
		{
			testName: "testCyclic",
			mockDir:  "testCyclic",
			config:   DefaultRunConfig(),
			expected: map[string][]string{
				"go.opentelemetry.io/build-tools/crosslink/testroot": {
					"go.opentelemetry.io/build-tools/crosslink/testroot/testA",
					"go.opentelemetry.io/build-tools/crosslink/testroot/testB"},
				"go.opentelemetry.io/build-tools/crosslink/testroot/testA": {
					"go.opentelemetry.io/build-tools/crosslink/testroot/testB",
					"go.opentelemetry.io/build-tools/crosslink/testroot"},
				// b has req on root but not necessary to write out with current comparison logic
				"go.opentelemetry.io/build-tools/crosslink/testroot/testB": {
					"go.opentelemetry.io/build-tools/crosslink/testroot/testA",
					"go.opentelemetry.io/build-tools/crosslink/testroot"},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			tmpRootDir, err := createTempTestDir(test.mockDir)
			if err != nil {
				t.Fatal("creating temp dir:", err)
			}

			err = renameGoMod(tmpRootDir)
			if err != nil {
				t.Errorf("error renaming gomod files: %v", err)
			}
			t.Cleanup(func() { os.RemoveAll(tmpRootDir) })

			test.config.RootPath = tmpRootDir

			rootModulePath, err := identifyRootModule(test.config.RootPath)
			if err != nil {
				t.Fatalf("error identifying root module: %v", err)
			}

			receivedMap, err := buildDepedencyGraph(test.config, rootModulePath)

			if assert.NoError(t, err, "error message on graph build %s") {
				assert.Equal(t, len(test.expected), len(receivedMap), "Module count does not match")
				for modName, moduleInfoActual := range receivedMap {
					requiredReplaceStatementsActual := moduleInfoActual.requiredReplaceStatements
					expectedReplaceStatements := test.expected[modName]
					// verify that the amount of replace statements in module match the amount that are in module.
					assert.Equal(t, len(expectedReplaceStatements), len(requiredReplaceStatementsActual), fmt.Sprintf("ModFilePath: %v \n Expected: %v \n Actual : %v",
						modName, expectedReplaceStatements, requiredReplaceStatementsActual))
					// ensure that they contain the same values
					for _, expectedReplaceStatement := range expectedReplaceStatements {

						if _, contains := requiredReplaceStatementsActual[expectedReplaceStatement]; !contains {
							t.Errorf("Expected replace statement : %s not in map %v", expectedReplaceStatement, requiredReplaceStatementsActual)
						}
					}

				}
			}

		})
	}
}
