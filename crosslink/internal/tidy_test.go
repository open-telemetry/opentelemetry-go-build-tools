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

package crosslink

import (
	"os"
	"path"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestTidy(t *testing.T) {
	defaultConfig := DefaultRunConfig()
	defaultConfig.Logger, _ = zap.NewDevelopment()
	defaultConfig.Verbose = true
	defaultConfig.Validate = true

	tests := []struct {
		name     string
		mock     string
		config   func(*RunConfig)
		expErr   string
		expSched []string
	}{
		{ // A -> B -> C should give CBA
			name:     "testTidyAcyclic",
			mock:     "testTidyAcyclic",
			config:   func(*RunConfig) {},
			expSched: []string{".", "testC", "testB", "testA"},
		},
		{ // A <=> B -> C without allowlist should error
			name:   "testTidyNotAllowlisted",
			mock:   "testTidyCyclic",
			config: func(*RunConfig) {},
			expErr: "list of circular dependencies does not match allowlist",
		},
		{ // A <=> B -> C with an over-permissive allowlist should error
			name: "testTidyOverpermissive",
			mock: "testTidyCyclic",
			config: func(config *RunConfig) {
				config.AllowCircular = path.Join(config.RootPath, "allow-circular-overpermissive.txt")
			},
			expErr: "list of circular dependencies does not match allowlist",
		},
		{ // A <=> B -> C should give CBAB
			name: "testTidyCyclic",
			mock: "testTidyCyclic",
			config: func(config *RunConfig) {
				config.AllowCircular = path.Join(config.RootPath, "allow-circular.txt")
			},
			expSched: []string{".", "testC", "testB", "testA", "testB"},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			testDir, err := createTempTestDir(test.mock)
			require.NoError(t, err, "error creating temporary directory")
			t.Cleanup(func() { os.RemoveAll(testDir) })
			err = renameGoMod(testDir)
			require.NoError(t, err, "error renaming gomod files")
			outputPath := path.Join(testDir, "schedule.txt")

			config := defaultConfig
			config.RootPath = testDir
			test.config(&config)

			err = Tidy(config, outputPath)

			if test.expErr != "" {
				require.ErrorContains(t, err, test.expErr, "expected error in Tidy")
				return
			}
			require.NoError(t, err, "unexpected error in Tidy")

			outputBytes, err := os.ReadFile(outputPath) // #nosec G304 -- Path comes from os.MkdirTemp
			require.NoError(t, err, "error reading output file")
			schedule := strings.Split(string(outputBytes), "\n")
			require.Equal(t, test.expSched, schedule, "generated schedule differs from expectation")
		})
	}
}
