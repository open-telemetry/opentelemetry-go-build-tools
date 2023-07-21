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

package chlog

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"

	"go.opentelemetry.io/build-tools/chloggen/internal/config"
)

func TestEntry(t *testing.T) {
	testCases := []struct {
		name      string
		entry     Entry
		expectErr string
		toString  string
	}{
		{
			name:      "empty",
			entry:     Entry{},
			expectErr: "'' is not a valid 'change_type'. Specify one of [breaking deprecation new_component enhancement bug_fix]",
		},
		{
			name: "missing_component",
			entry: Entry{
				ChangeType: "enhancement",
				Note:       "enhance!",
				Issues:     []int{123},
				SubText:    "",
			},
			expectErr: "specify a 'component'",
		},
		{
			name: "missing_note",
			entry: Entry{
				ChangeType: "bug_fix",
				Component:  "bar",
				Issues:     []int{123},
				SubText:    "",
			},
			expectErr: "specify a 'note'",
		},
		{
			name: "missing_issue",
			entry: Entry{
				ChangeType: "bug_fix",
				Component:  "bar",
				Note:       "fix bar",
				SubText:    "",
			},
			expectErr: "specify one or more issues #'s",
		},
		{
			name: "valid",
			entry: Entry{
				ChangeType: "breaking",
				Component:  "foo",
				Note:       "broke foo",
				Issues:     []int{123},
				SubText:    "",
			},
			toString: "- `foo`: broke foo (#123)",
		},
		{
			name: "multiple_issues",
			entry: Entry{
				ChangeType: "breaking",
				Component:  "foo",
				Note:       "broke foo",
				Issues:     []int{123, 345},
				SubText:    "",
			},
			toString: "- `foo`: broke foo (#123, #345)",
		},
		{
			name: "subtext",
			entry: Entry{
				ChangeType: "breaking",
				Component:  "foo",
				Note:       "broke foo",
				Issues:     []int{123},
				SubText:    "more details",
			},
			toString: "- `foo`: broke foo (#123)\n  more details",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.entry.Validate()
			if tc.expectErr != "" {
				assert.Equal(t, tc.expectErr, err.Error())
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tc.toString, tc.entry.String())
		})
	}

}

func TestReadDeleteEntries(t *testing.T) {
	tempDir := t.TempDir()
	entriesDir := filepath.Join(tempDir, config.DefaultChloggenDir)
	require.NoError(t, os.Mkdir(entriesDir, os.ModePerm))

	entryA := Entry{
		ChangeType: "breaking",
		Component:  "foo",
		Note:       "broke foo",
		Issues:     []int{123},
	}

	bytesA, err := yaml.Marshal(entryA)
	require.NoError(t, err)

	fileA, err := os.CreateTemp(entriesDir, "*.yaml")
	require.NoError(t, err)
	defer fileA.Close()

	_, err = fileA.Write(bytesA)
	require.NoError(t, err)

	entryB := Entry{
		ChangeType: "bug_fix",
		Component:  "bar",
		Note:       "fix bar",
		Issues:     []int{345, 678},
		SubText:    "more details",
	}

	bytesB, err := yaml.Marshal(entryB)
	require.NoError(t, err)

	fileB, err := os.CreateTemp(entriesDir, "*.yaml")
	require.NoError(t, err)
	defer fileB.Close()

	_, err = fileB.Write(bytesB)
	require.NoError(t, err)

	// Put config and template files in chlogs_dir to ensure they are ignored when reading/deleting entries
	configYAML, err := os.CreateTemp(entriesDir, "config.yaml")
	require.NoError(t, err)
	defer configYAML.Close()

	templateYAML, err := os.CreateTemp(entriesDir, "TEMPLATE.yaml")
	require.NoError(t, err)
	defer templateYAML.Close()

	cfg := config.New(tempDir)
	cfg.ConfigYAML = configYAML.Name()
	cfg.TemplateYAML = templateYAML.Name()

	entries, err := ReadEntries(cfg)
	assert.NoError(t, err)

	assert.ElementsMatch(t, []*Entry{&entryA, &entryB}, entries)

	assert.NoError(t, DeleteEntries(cfg))
	entries, err = ReadEntries(cfg)
	assert.NoError(t, err)
	assert.Empty(t, entries)

	// Ensure these weren't deleted
	assert.FileExists(t, cfg.ConfigYAML)
	assert.FileExists(t, cfg.TemplateYAML)
}
