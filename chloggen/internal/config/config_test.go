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
	"gopkg.in/yaml.v3"
)

func TestNew(t *testing.T) {
	root := "/tmp"
	cfg := New(root)
	assert.Equal(t, filepath.Join(root, DefaultChloggenDir), cfg.ChlogsDir)
	assert.Equal(t, filepath.Join(root, DefaultChangelogMD), cfg.ChangelogMD)
	assert.Equal(t, filepath.Join(root, DefaultChloggenDir, DefaultTemplateYAML), cfg.TemplateYAML)
}

func TestNewFromFile(t *testing.T) {
	tempDir := t.TempDir()

	cfg := New(tempDir)
	cfg.ChlogsDir = ".test"
	cfg.ChangelogMD = "CHANGELOG-custom.md"
	cfg.TemplateYAML = "TEMPLATE-custom.yaml"

	cfgBytes, err := yaml.Marshal(cfg)
	require.NoError(t, err)

	cfgFile, err := os.CreateTemp(tempDir, "*.yaml")
	require.NoError(t, err)
	defer cfgFile.Close()

	_, err = cfgFile.Write(cfgBytes)
	require.NoError(t, err)

	actualCfg, err := NewFromFile(tempDir, filepath.Base(cfgFile.Name()))
	assert.NoError(t, err)
	assert.Equal(t, filepath.Join(tempDir, ".test"), actualCfg.ChlogsDir)
	assert.Equal(t, filepath.Join(tempDir, "CHANGELOG-custom.md"), actualCfg.ChangelogMD)
	assert.Equal(t, filepath.Join(tempDir, "TEMPLATE-custom.yaml"), actualCfg.TemplateYAML)
}

func TestNewFromFileErr(t *testing.T) {
	tempDir := t.TempDir()

	_, err := NewFromFile(tempDir, "nonexistent.yaml")
	assert.Error(t, err)

	// Write a file with invalid YAML and then read it back to get expected error
	cfgFile, err := os.CreateTemp(tempDir, "*.yaml")
	require.NoError(t, err)
	defer cfgFile.Close()

	_, err = cfgFile.WriteString("!@#$%^&*())")
	require.NoError(t, err)

	_, err = NewFromFile(tempDir, filepath.Base(cfgFile.Name()))
	assert.Error(t, err)
}
