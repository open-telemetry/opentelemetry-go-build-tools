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
			config:   func(config *RunConfig) {},
			expSched: []string{".", "testC", "testB", "testA"},
		},
		{ // A <=> B -> C without allowlist should error
			name:   "testTidyNotAllowlisted",
			mock:   "testTidyCyclic",
			config: func(config *RunConfig) {},
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
			} else {
				require.NoError(t, err, "unexpected error in Tidy")
			}

			outputBytes, err := os.ReadFile(outputPath)
			require.NoError(t, err, "error reading output file")
			schedule := strings.Split(string(outputBytes), "\n")
			require.Equal(t, test.expSched, schedule, "generated schedule differs from expectation")
		})
	}
}
