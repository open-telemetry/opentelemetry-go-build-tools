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
package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"

	cl "go.opentelemetry.io/build-tools/crosslink/internal"
)

func TestTransform(t *testing.T) {
	tests := []struct {
		testName   string
		inputSlice []string
	}{
		{
			testName: "with items",
			inputSlice: []string{
				"example.com/testA",
				"example.com/testB",
				"example.com/testC",
				"example.com/testD",
				"example.com/testE",
			},
		},
		{
			testName:   "with empty",
			inputSlice: []string{},
		},
	}

	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			actual := transformExclude(test.inputSlice)

			// len must match
			assert.Len(t, actual, len(test.inputSlice))

			// test for existence
			for _, val := range test.inputSlice {
				_, exists := actual[val]
				assert.True(t, exists)
			}
		})
	}
}

var configReset = func() {
	comCfg.runConfig = cl.DefaultRunConfig()
	comCfg.rootCommand.SetArgs([]string{})
}

// Validate run config is valid after pre run.
func TestPreRun(t *testing.T) {

	validRootPath, err := filepath.Abs("../../")
	if err != nil {
		t.Errorf("could not parse expected root path: %e", err)
	}
	validProdLogger, err := zap.NewProduction()
	assert.NoError(t, err, "failed to create prod logger")

	tests := []struct {
		testName       string
		args           []string
		mockConfig     cl.RunConfig
		expectedConfig cl.RunConfig
	}{
		{
			testName:   "Default Config",
			args:       []string{},
			mockConfig: cl.DefaultRunConfig(),
			expectedConfig: cl.RunConfig{
				Overwrite:     false,
				RootPath:      validRootPath,
				Logger:        validProdLogger,
				ExcludedPaths: make(map[string]struct{}),
			},
		},
		{
			testName: "with overwrite",
			mockConfig: cl.RunConfig{
				Overwrite: true,
			},
			expectedConfig: cl.RunConfig{
				Overwrite: true,
				Verbose:   true,
				RootPath:  validRootPath,
			},
			args: []string{"--overwrite"},
		},
		{
			testName: "with overwrite and verbose=false",
			mockConfig: cl.RunConfig{
				Overwrite: true,
				Verbose:   false,
			},
			expectedConfig: cl.RunConfig{
				Overwrite: true,
				Verbose:   false,
				RootPath:  validRootPath,
			},
			args: []string{"--overwrite", "--verbose=false"},
		},
		{
			testName: "with prune exclusive",
			mockConfig: cl.RunConfig{
				Prune: true,
			},
			expectedConfig: cl.RunConfig{
				Prune:    true,
				RootPath: validRootPath,
			},
			args: []string{"--prune"},
		},
		{
			testName: "with prune exclusive short",
			mockConfig: cl.RunConfig{
				Prune: true,
			},
			expectedConfig: cl.RunConfig{
				Prune:    true,
				RootPath: validRootPath,
			},
			args: []string{"-p"},
		},
		{
			testName: "with verbose exclusive",
			mockConfig: cl.RunConfig{
				Verbose: true,
			},
			expectedConfig: cl.RunConfig{
				Verbose:  true,
				RootPath: validRootPath,
			},
			args: []string{"--verbose"},
		},
		{
			testName: "with verbose exclusive short",
			mockConfig: cl.RunConfig{
				Verbose: true,
			},
			expectedConfig: cl.RunConfig{
				Verbose:  true,
				RootPath: validRootPath,
			},
			args: []string{"-v"},
		},
		{
			testName:   "with good root path",
			mockConfig: cl.DefaultRunConfig(),
			expectedConfig: cl.RunConfig{
				RootPath: validRootPath,
				Logger:   validProdLogger,
			},
			args: []string{fmt.Sprintf("--root=%s", validRootPath)},
		},
	}

	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			t.Cleanup(configReset)
			comCfg.runConfig = test.mockConfig

			err := comCfg.rootCommand.ParseFlags(test.args)
			if err != nil {
				t.Errorf("Failed to parse flags: %v", err)
			}

			testPreRun := comCfg.rootCommand.PersistentPreRunE
			err = testPreRun(&comCfg.rootCommand, nil)
			assert.NoError(t, err, "Pre Run returned error")

			if diff := cmp.Diff(test.expectedConfig, comCfg.runConfig, cmpopts.IgnoreFields(cl.RunConfig{}, "Logger", "ExcludedPaths", "SkippedPaths")); diff != "" {
				t.Errorf("TestCase: %s \n Config{} mismatch (-want +got):\n%s", test.testName, diff)
			}
		})
	}
}

// isolated test because the working directory needs to changed
// and it will keep the happy path test above clean
func TestBadRootPath(t *testing.T) {
	t.Cleanup(configReset)
	mockConfig := cl.DefaultRunConfig()
	args := []string{}

	// under the assumption that this is not a nested git repository
	err := os.Chdir("/../../..")
	assert.NoError(t, err, "failed to change working directory")
	comCfg.runConfig = mockConfig

	err = comCfg.rootCommand.ParseFlags(args)
	if err != nil {
		t.Errorf("Failed to parse flags: %v", err)
	}

	testPreRun := comCfg.rootCommand.PersistentPreRunE
	err = testPreRun(&comCfg.rootCommand, nil)
	assert.Error(t, err, "Pre Run did not return error")
}
