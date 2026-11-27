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

package cmd

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSubcommandsRegistered(t *testing.T) {
	want := map[string]*cobra.Command{
		"diff":       diffCmd,
		"prerelease": prereleaseCmd,
		"sync":       syncCmd,
		"tag":        tagCmd,
		"verify":     verifyCmd,
	}
	for name, c := range want {
		t.Run(name, func(t *testing.T) {
			found, _, err := rootCmd.Find([]string{name})
			require.NoError(t, err)
			assert.Same(t, c, found)
			assert.NotEmpty(t, c.Short)
		})
	}
}

func TestRootVersioningFileFlag(t *testing.T) {
	f := rootCmd.PersistentFlags().Lookup("versioning-file")
	require.NotNil(t, f)
	assert.Equal(t, "v", f.Shorthand)
	assert.Contains(t, f.DefValue, "versions.yaml")
}

func TestCommandFlags(t *testing.T) {
	tests := []struct {
		cmd       *cobra.Command
		flag      string
		shorthand string
		required  bool
	}{
		{diffCmd, "module-set-name", "m", true},
		{diffCmd, "previous-version", "p", true},
		{prereleaseCmd, "all-module-sets", "a", false},
		{prereleaseCmd, "module-set-names", "m", false},
		{prereleaseCmd, "commit-to-different-branch", "b", false},
		{syncCmd, "other-repo-root", "o", true},
		{syncCmd, "module-set-names", "m", true},
		{syncCmd, "other-versioning-file", "", false},
		{syncCmd, "commit-hash", "", false},
		{syncCmd, "compute-versions-locally", "", false},
		{syncCmd, "all-module-sets", "a", false},
		{syncCmd, "skip-go-mod-tidy", "s", false},
		{tagCmd, "commit-hash", "c", true},
		{tagCmd, "module-set-name", "m", true},
		{tagCmd, "delete-module-set-tags", "d", false},
		{tagCmd, "print-tags", "p", false},
	}
	for _, tc := range tests {
		t.Run(tc.cmd.Use+"/"+tc.flag, func(t *testing.T) {
			f := tc.cmd.Flags().Lookup(tc.flag)
			require.NotNil(t, f)
			assert.Equal(t, tc.shorthand, f.Shorthand)
			_, isRequired := f.Annotations[cobra.BashCompOneRequiredFlag]
			if tc.required {
				assert.Equal(t, []string{"true"}, f.Annotations[cobra.BashCompOneRequiredFlag])
			} else if isRequired {
				assert.Equal(t, []string{"false"}, f.Annotations[cobra.BashCompOneRequiredFlag])
			}
		})
	}
}

func TestSyncPreRunAllModuleSets(t *testing.T) {
	old := allModuleSetsSync
	t.Cleanup(func() {
		allModuleSetsSync = old
		syncCmd.Flags().Lookup("module-set-names").Annotations[cobra.BashCompOneRequiredFlag] = []string{"true"}
	})

	allModuleSetsSync = true
	syncCmd.PreRun(syncCmd, nil)
	f := syncCmd.Flags().Lookup("module-set-names")
	assert.Equal(t, []string{"false"}, f.Annotations[cobra.BashCompOneRequiredFlag])
}

func TestPrereleasePreRunAllModuleSets(t *testing.T) {
	require.NotNil(t, prereleaseCmd.PreRun)
	old := allModuleSets
	t.Cleanup(func() { allModuleSets = old })
	allModuleSets = true
	assert.NotPanics(t, func() { prereleaseCmd.PreRun(prereleaseCmd, nil) })
}
