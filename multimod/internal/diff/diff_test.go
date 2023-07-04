// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package diff

import (
	"errors"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"go.opentelemetry.io/build-tools/internal/repo"
)

func TestNormalizeVersion(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		output string
	}{
		{
			name:   "version not prefixed",
			input:  "0.1.2",
			output: "v0.1.2",
		},
		{
			name:   "version prefixed",
			input:  "v0.2.2",
			output: "v0.2.2",
		},
	}
	for _, tt := range tests {
		require.Equal(t, tt.output, normalizeVersion(tt.input))
	}
}

func TestHasChanged(t *testing.T) {
	tests := []struct {
		name         string
		tag          string
		modset       string
		versionsFile string
		expected     bool
		err          error
	}{
		{
			name:     "changes found, tag exists",
			expected: true,
			tag:      "v0.8.0",
			modset:   "tools",
		},
		{
			name:     "invalid tag",
			expected: false,
			err:      errors.New("invalid tag"),
			tag:      "1.2.3",
			modset:   "tools",
		},
		{
			name:     "invalid modset",
			expected: false,
			err:      errors.New("invalid modset"),
			tag:      "v0.8.0",
			modset:   "invalid",
		},
		{
			name:         "invalid versions file",
			expected:     false,
			err:          errors.New("invalid versions file"),
			tag:          "v0.8.0",
			versionsFile: "invalid.yaml",
			modset:       "tools",
		},
	}
	repoRoot, err := repo.FindRoot()

	require.NoError(t, err)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			versionFile := filepath.Join(repoRoot, "versions.yaml")
			if len(tt.versionsFile) > 0 {
				versionFile = filepath.Join(repoRoot, tt.versionsFile)
			}
			actual, changedFiles, err := HasChanged(repoRoot, versionFile, tt.tag, tt.modset)
			if tt.err != nil {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
			require.Equal(t, tt.expected, actual)
			if actual {
				require.True(t, len(changedFiles) > 0)
			} else {
				require.False(t, len(changedFiles) > 0)
			}
		})
	}
}
