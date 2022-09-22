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

func TestNew(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		wantErr  string
	}{
		{
			name:     "no_extension",
			filename: "my-change",
		},
		{
			name:     "yaml_extension",
			filename: "some-change.yaml",
		},
		{
			name:     "yml_extension",
			filename: "some-change.yml",
		},
		{
			name:     "replace_forward_slash",
			filename: "replace/forward/slash",
		},
		{
			name:     "bad_extension",
			filename: "my-change.txt",
			wantErr:  "non-yaml extension",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ctx := setupTestDir(t, []*chlog.Entry{})
			err := initialize(ctx, tc.filename)
			if tc.wantErr != "" {
				require.Regexp(t, tc.wantErr, err)
				return
			}
			require.NoError(t, err)

			require.Error(t, validate(ctx), "The new entry should not be valid without user input")
		})
	}
}

func TestCleanFilename(t *testing.T) {
	require.Equal(t, "fix_some_bug", cleanFileName("fix/some_bug"))
	require.Equal(t, "fix_some_bug", cleanFileName("fix\\some_bug"))
}
