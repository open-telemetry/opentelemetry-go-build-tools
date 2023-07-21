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
	"strings"

	"gopkg.in/yaml.v3"
)

const (
	DefaultChangelogMD  = "CHANGELOG.md"
	DefaultChloggenDir  = ".chloggen"
	DefaultTemplateYAML = "TEMPLATE.yaml"
)

type Config struct {
	ChangelogMD  string `yaml:"changelog_md"`
	ChlogsDir    string `yaml:"chlogs_dir"`
	TemplateYAML string `yaml:"template_yaml"`
	ConfigYAML   string
}

func New(rootDir string) Config {
	return Config{
		ChangelogMD:  filepath.Join(rootDir, DefaultChangelogMD),
		ChlogsDir:    filepath.Join(rootDir, DefaultChloggenDir),
		TemplateYAML: filepath.Join(rootDir, DefaultChloggenDir, DefaultTemplateYAML),
	}
}

func NewFromFile(rootDir string, filename string) (Config, error) {
	cfg := New(rootDir)
	cfg.ConfigYAML = filepath.Clean(filepath.Join(rootDir, filename))
	cfgBytes, err := os.ReadFile(cfg.ConfigYAML)
	if err != nil {
		return Config{}, err
	}
	if err = yaml.Unmarshal(cfgBytes, &cfg); err != nil {
		return Config{}, err
	}

	// If the user specified any of the following, interpret as a relative path from rootDir
	// (unless they specified an absolute path including rootDir)
	if !strings.HasPrefix(cfg.ChangelogMD, rootDir) {
		cfg.ChangelogMD = filepath.Join(rootDir, cfg.ChangelogMD)
	}
	if !strings.HasPrefix(cfg.ChlogsDir, rootDir) {
		cfg.ChlogsDir = filepath.Join(rootDir, cfg.ChlogsDir)
	}
	if !strings.HasPrefix(cfg.TemplateYAML, rootDir) {
		cfg.TemplateYAML = filepath.Join(rootDir, cfg.TemplateYAML)
	}
	return cfg, nil
}
