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
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

const (
	DefaultEntriesDir        = ".chloggen"
	DefaultTemplateYAML      = "TEMPLATE.yaml"
	DefaultChangeLogKey      = "default"
	DefaultChangeLogFilename = "CHANGELOG.md"
)

type Config struct {
	ChangeLogs        map[string]string `yaml:"change_logs"`
	DefaultChangeLogs []string          `yaml:"default_change_logs"`
	EntriesDir        string            `yaml:"entries_dir"`
	TemplateYAML      string            `yaml:"template_yaml"`
	Components        []string          `yaml:"components"`
	IssueLink         string            `yaml:"issue_link"`
	ConfigYAML        string
}

func New(rootDir string) *Config {
	return &Config{
		ChangeLogs:        map[string]string{DefaultChangeLogKey: filepath.Join(rootDir, DefaultChangeLogFilename)},
		DefaultChangeLogs: []string{DefaultChangeLogKey},
		EntriesDir:        filepath.Join(rootDir, DefaultEntriesDir),
		TemplateYAML:      filepath.Join(rootDir, DefaultEntriesDir, DefaultTemplateYAML),
	}
}

func NewFromFile(rootDir string, cfgFilename string) (*Config, error) {
	cfgYAML := filepath.Clean(filepath.Join(rootDir, cfgFilename))
	cfgBytes, err := os.ReadFile(cfgYAML)
	if err != nil {
		return nil, err
	}
	cfg := &Config{}
	if err = yaml.Unmarshal(cfgBytes, &cfg); err != nil {
		return nil, err
	}

	cfg.ConfigYAML = cfgYAML
	if cfg.EntriesDir == "" {
		cfg.EntriesDir = filepath.Join(rootDir, DefaultEntriesDir)
	} else if !strings.HasPrefix(cfg.EntriesDir, rootDir) {
		cfg.EntriesDir = filepath.Join(rootDir, cfg.EntriesDir)
	}

	if cfg.TemplateYAML == "" {
		cfg.TemplateYAML = filepath.Join(rootDir, DefaultEntriesDir, DefaultTemplateYAML)
	} else if !strings.HasPrefix(cfg.TemplateYAML, rootDir) {
		cfg.TemplateYAML = filepath.Join(rootDir, cfg.TemplateYAML)
	}

	if len(cfg.ChangeLogs) == 0 && len(cfg.DefaultChangeLogs) > 0 {
		return nil, errors.New("cannot specify 'default_changelogs' without 'changelogs'")
	}

	if len(cfg.ChangeLogs) == 0 {
		cfg.ChangeLogs = map[string]string{DefaultChangeLogKey: filepath.Join(rootDir, DefaultChangeLogFilename)}
		cfg.DefaultChangeLogs = []string{DefaultChangeLogKey}
		return cfg, nil
	}

	// The user specified at least one changelog. Interpret filename as a relative path from rootDir
	// (unless they specified an absolute path including rootDir)
	for key, filename := range cfg.ChangeLogs {
		if !strings.HasPrefix(filename, rootDir) {
			cfg.ChangeLogs[key] = filepath.Join(rootDir, filename)
		}
	}

	for _, key := range cfg.DefaultChangeLogs {
		if _, ok := cfg.ChangeLogs[key]; !ok {
			return nil, fmt.Errorf("'default_changelogs' contains key %q which is not defined in 'changelogs'", key)
		}
	}

	return cfg, nil
}
