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
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"go.opentelemetry.io/build-tools/chloggen/internal/chlog"
)

const validateUsage = `Usage:
  chloggen validate [flags]

Flags:
  -h, --help   help for validate

Global Flags:
      --chloggen-directory string   directory containing unreleased change log entries (default: .chloggen)`

func TestValidateErr(t *testing.T) {
	var out, err string

	out, err = runCobra(t, "validate", "--help")
	assert.Contains(t, out, validateUsage)
	assert.Empty(t, err)

	out, err = runCobra(t, "validate")
	assert.Contains(t, out, validateUsage)
	assert.Contains(t, err, "no such file or directory")
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		entries []*chlog.Entry
		wantErr string
	}{
		{
			name:    "all_valid",
			entries: getSampleEntries(),
		},
		{
			name: "invalid_change_type",
			entries: func() []*chlog.Entry {
				return append(getSampleEntries(), &chlog.Entry{
					ChangeType: "fake",
					Component:  "receiver/foo",
					Note:       "Add some bar",
					Issues:     []int{12345},
				})
			}(),
			wantErr: "'fake' is not a valid 'change_type'",
		},
		{
			name: "missing_component",
			entries: func() []*chlog.Entry {
				return append(getSampleEntries(), &chlog.Entry{
					ChangeType: chlog.BugFix,
					Component:  "",
					Note:       "Add some bar",
					Issues:     []int{12345},
				})
			}(),
			wantErr: "specify a 'component'",
		},
		{
			name: "empty_component",
			entries: func() []*chlog.Entry {
				return append(getSampleEntries(), &chlog.Entry{
					ChangeType: chlog.BugFix,
					Component:  " ",
					Note:       "Add some bar",
					Issues:     []int{12345},
				})
			}(),
			wantErr: "specify a 'component'",
		},
		{
			name: "missing_note",
			entries: func() []*chlog.Entry {
				return append(getSampleEntries(), &chlog.Entry{
					ChangeType: chlog.BugFix,
					Component:  "receiver/foo",
					Note:       "",
					Issues:     []int{12345},
				})
			}(),
			wantErr: "specify a 'note'",
		},
		{
			name: "empty_note",
			entries: func() []*chlog.Entry {
				return append(getSampleEntries(), &chlog.Entry{
					ChangeType: chlog.BugFix,
					Component:  "receiver/foo",
					Note:       " ",
					Issues:     []int{12345},
				})
			}(),
			wantErr: "specify a 'note'",
		},
		{
			name: "missing_issue",
			entries: func() []*chlog.Entry {
				return append(getSampleEntries(), &chlog.Entry{
					ChangeType: chlog.BugFix,
					Component:  "receiver/foo",
					Note:       "Add some bar",
					Issues:     []int{},
				})
			}(),
			wantErr: "specify one or more issues #'s",
		},
		{
			name: "all_invalid",
			entries: func() []*chlog.Entry {
				sampleEntries := getSampleEntries()
				for _, e := range sampleEntries {
					e.ChangeType = "fake"
				}
				return sampleEntries
			}(),
			wantErr: "'fake' is not a valid 'change_type'",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			globalCfg = setupTestDir(t, tc.entries)

			out, err := runCobra(t, "validate")

			if tc.wantErr != "" {
				assert.Regexp(t, tc.wantErr, err)
			} else {
				assert.Empty(t, err)
				assert.Contains(t, out, fmt.Sprintf("PASS: all files in %s/ are valid", globalCfg.ChloggenDir))
			}
		})
	}
}
