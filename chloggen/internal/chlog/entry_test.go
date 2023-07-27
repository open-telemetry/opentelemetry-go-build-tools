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
		name             string
		entry            Entry
		requireChangeLog bool
		validChangeLogs  []string
		expectErr        string
		toString         string
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
			name: "missing_required_changelog",
			entry: Entry{
				ChangeType: "bug_fix",
				Component:  "bar",
				Note:       "fix bar",
				Issues:     []int{123},
				SubText:    "",
			},
			requireChangeLog: true,
			validChangeLogs:  []string{"foo"},
			expectErr:        "specify one or more 'change_logs'",
		},
		{
			name: "invalid_changelog",
			entry: Entry{
				ChangeLogs: []string{"bar"},
				ChangeType: "bug_fix",
				Component:  "bar",
				Note:       "fix bar",
				Issues:     []int{123},
				SubText:    "",
			},
			validChangeLogs: []string{"foo"},
			expectErr:       "'bar' is not a valid 'change_log'. Specify one of [foo]",
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
		{
			name: "required_changelog",
			entry: Entry{
				ChangeLogs: []string{"foo"},
				ChangeType: "breaking",
				Component:  "foo",
				Note:       "broke foo",
				Issues:     []int{123},
				SubText:    "more details",
			},
			requireChangeLog: true,
			validChangeLogs:  []string{"foo"},
			toString:         "- `foo`: broke foo (#123)\n  more details",
		},
		{
			name: "default_changelog",
			entry: Entry{
				ChangeLogs: []string{"foo"},
				ChangeType: "breaking",
				Component:  "foo",
				Note:       "broke foo",
				Issues:     []int{123},
				SubText:    "more details",
			},
			requireChangeLog: false,
			validChangeLogs:  []string{"foo"},
			toString:         "- `foo`: broke foo (#123)\n  more details",
		},
		{
			name: "subset_of_changelogs",
			entry: Entry{
				ChangeLogs: []string{"foo", "bar"},
				ChangeType: "breaking",
				Component:  "foo",
				Note:       "broke foo",
				Issues:     []int{123},
				SubText:    "more details",
			},
			validChangeLogs: []string{"foo", "bar", "baz"},
			toString:        "- `foo`: broke foo (#123)\n  more details",
		},
		{
			name: "all_changelogs",
			entry: Entry{
				ChangeLogs: []string{"foo", "bar"},
				ChangeType: "breaking",
				Component:  "foo",
				Note:       "broke foo",
				Issues:     []int{123},
				SubText:    "more details",
			},
			validChangeLogs: []string{"foo", "bar"},
			toString:        "- `foo`: broke foo (#123)\n  more details",
		},
		{
			name: "all_changelogs",
			entry: Entry{
				ChangeLogs: []string{"foo", "bar"},
				ChangeType: "breaking",
				Component:  "foo",
				Note:       "broke foo",
				Issues:     []int{123},
				SubText:    "more details",
			},
			validChangeLogs: []string{"foo", "bar"},
			toString:        "- `foo`: broke foo (#123)\n  more details",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.entry.Validate(tc.requireChangeLog, tc.validChangeLogs...)
			if tc.expectErr != "" {
				assert.Error(t, err)
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
	entriesDir := filepath.Join(tempDir, config.DefaultChlogsDir)
	require.NoError(t, os.Mkdir(entriesDir, os.ModePerm))

	entryA := Entry{
		ChangeLogs: []string{"foo"},
		ChangeType: "breaking",
		Component:  "foo",
		Note:       "broke foo",
		Issues:     []int{123},
	}
	writeEntry(t, entriesDir, &entryA)

	entryB := Entry{
		ChangeLogs: []string{"bar"},
		ChangeType: "bug_fix",
		Component:  "bar",
		Note:       "fix bar",
		Issues:     []int{345, 678},
		SubText:    "more details",
	}
	writeEntry(t, entriesDir, &entryB)

	entryC := Entry{
		ChangeLogs: []string{},
		ChangeType: "enhancement",
		Component:  "other",
		Note:       "enhance!",
		Issues:     []int{555},
	}
	writeEntry(t, entriesDir, &entryC)

	entryD := Entry{
		ChangeLogs: []string{"foo", "bar"},
		ChangeType: "deprecation",
		Component:  "foobar",
		Note:       "deprecate something",
		Issues:     []int{999},
	}
	writeEntry(t, entriesDir, &entryD)

	// Put config and template files in chlogs_dir to ensure they are ignored when reading/deleting entries
	configYAML, err := os.Create(filepath.Join(entriesDir, "config.yaml")) //nolint:gosec
	require.NoError(t, err)
	defer configYAML.Close()

	templateYAML, err := os.Create(filepath.Join(entriesDir, "TEMPLATE.yaml")) //nolint:gosec
	require.NoError(t, err)
	defer templateYAML.Close()

	cfg := &config.Config{
		ConfigYAML:   configYAML.Name(),
		TemplateYAML: templateYAML.Name(),
		ChangeLogs: map[string]string{
			"foo": filepath.Join(entriesDir, "CHANGELOG.foo.md"),
			"bar": filepath.Join(entriesDir, "CHANGELOG.bar.md"),
		},
		DefaultChangeLogs: []string{"foo"},
		ChlogsDir:         entriesDir,
	}

	changeLogEntries, err := ReadEntries(cfg)
	assert.NoError(t, err)

	assert.Equal(t, 2, len(changeLogEntries))

	assert.Contains(t, changeLogEntries, "foo")
	assert.Contains(t, changeLogEntries, "bar")

	assert.ElementsMatch(t, []*Entry{&entryA, &entryC, &entryD}, changeLogEntries["foo"])
	assert.ElementsMatch(t, []*Entry{&entryB, &entryD}, changeLogEntries["bar"])

	assert.NoError(t, DeleteEntries(cfg))
	changeLogEntries, err = ReadEntries(cfg)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(changeLogEntries))
	assert.Empty(t, changeLogEntries["foo"])
	assert.Empty(t, changeLogEntries["bar"])

	// Ensure these weren't deleted
	assert.FileExists(t, cfg.ConfigYAML)
	assert.FileExists(t, cfg.TemplateYAML)
}

func writeEntry(t *testing.T, dir string, entry *Entry) {
	entryBytes, err := yaml.Marshal(entry)
	require.NoError(t, err)

	entryFile, err := os.CreateTemp(dir, "*.yaml")
	require.NoError(t, err)
	defer entryFile.Close()

	_, err = entryFile.Write(entryBytes)
	require.NoError(t, err)
}
