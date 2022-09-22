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
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/build-tools/chloggen/internal/chlog"
)

func TestUpdateE2E(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Windows line breaks cause comparison failures w/ golden files.")
	}
	tests := []struct {
		name    string
		entries []*chlog.Entry
		version string
		dry     bool
	}{
		{
			name:    "all_change_types",
			entries: getSampleEntries(),
			version: "v0.45.0",
		},
		{
			name:    "all_change_types_multiple",
			entries: append(getSampleEntries(), getSampleEntries()...),
			version: "v0.45.0",
		},
		{
			name:    "dry_run",
			entries: getSampleEntries(),
			version: "v0.45.0",
			dry:     true,
		},
		{
			name:    "deprecation_only",
			entries: []*chlog.Entry{deprecationEntry()},
			version: "v0.45.0",
		},
		{
			name:    "new_component_only",
			entries: []*chlog.Entry{newComponentEntry()},
			version: "v0.45.0",
		},
		{
			name:    "bug_fix_only",
			entries: []*chlog.Entry{bugFixEntry()},
			version: "v0.45.0",
		},
		{
			name:    "enhancement_only",
			entries: []*chlog.Entry{enhancementEntry()},
			version: "v0.45.0",
		},
		{
			name:    "breaking_only",
			entries: []*chlog.Entry{breakingEntry()},
			version: "v0.45.0",
		},
		{
			name:    "subtext",
			entries: []*chlog.Entry{entryWithSubtext()},
			version: "v0.45.0",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ctx := setupTestDir(t, tc.entries)

			require.NoError(t, update(ctx, tc.version, tc.dry))

			actualBytes, err := os.ReadFile(ctx.ChangelogMD)
			require.NoError(t, err)

			expectedChangelogMD := filepath.Join("testdata", tc.name+".md")
			expectedBytes, err := os.ReadFile(filepath.Clean(expectedChangelogMD))
			require.NoError(t, err)

			require.Equal(t, string(expectedBytes), string(actualBytes))

			remainingYAMLs, err := filepath.Glob(filepath.Join(ctx.UnreleasedDir, "*.yaml"))
			require.NoError(t, err)
			if tc.dry {
				require.Equal(t, 1+len(tc.entries), len(remainingYAMLs))
			} else {
				require.Equal(t, 1, len(remainingYAMLs))
				require.Equal(t, ctx.TemplateYAML, remainingYAMLs[0])
			}
		})
	}
}
