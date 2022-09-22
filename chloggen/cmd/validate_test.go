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

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/build-tools/chloggen/internal/chlog"
)

func TestValidateE2E(t *testing.T) {
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
			ctx := setupTestDir(t, tc.entries)

			err := validate(ctx)
			if tc.wantErr != "" {
				require.Regexp(t, tc.wantErr, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
