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

package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewFromFile_Deep(t *testing.T) {
	tempDir := t.TempDir()
	rootDir := filepath.Join(tempDir, "root")
	err := os.Mkdir(rootDir, 0750)
	require.NoError(t, err)

	// Case 1: All Relative Paths (Regression Check)
	t.Run("RelativePaths", func(t *testing.T) {
		configFile := filepath.Join(rootDir, "rel_config.yaml")
		content := `
entries_dir: my_entries
template_yaml: my_template.yaml
change_logs:
  default: my_changelog.md
`
		err := os.WriteFile(configFile, []byte(content), 0600)
		require.NoError(t, err)

		cfg, err := NewFromFile(rootDir, "rel_config.yaml")
		require.NoError(t, err)

		assert.Equal(t, filepath.Join(rootDir, "my_entries"), cfg.EntriesDir)
		assert.Equal(t, filepath.Join(rootDir, "my_template.yaml"), cfg.TemplateYAML)
		assert.Equal(t, filepath.Join(rootDir, "my_changelog.md"), cfg.ChangeLogs["default"])
	})

	// Case 2: All Absolute Paths (The Fix)
	t.Run("AbsolutePaths", func(t *testing.T) {
		// Create paths outside the root
		absEntries := filepath.Join(tempDir, "abs_entries")
		absTemplate := filepath.Join(tempDir, "abs_template.yaml")
		absChangelog := filepath.Join(tempDir, "abs_changelog.md")

		configFile := filepath.Join(tempDir, "abs_config.yaml")
		content := "entries_dir: " + absEntries + "\n" +
			"template_yaml: " + absTemplate + "\n" +
			"change_logs:\n  default: " + absChangelog + "\n"

		err := os.WriteFile(configFile, []byte(content), 0600)
		require.NoError(t, err)

		// Note: We pass tempDir as rootDir here just to verify it doesn't try to join them
		// But strictly speaking, the rootDir doesn't matter for absolute paths now.
		cfg, err := NewFromFile(rootDir, configFile)
		require.NoError(t, err)

		assert.Equal(t, absEntries, cfg.EntriesDir)
		assert.Equal(t, absTemplate, cfg.TemplateYAML)
		assert.Equal(t, absChangelog, cfg.ChangeLogs["default"])
	})
}
