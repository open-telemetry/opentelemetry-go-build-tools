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
	"path/filepath"
)

const (
	DefaultChangelogMD  = "CHANGELOG.md"
	DefaultChloggenDir  = ".chloggen"
	DefaultTemplateYAML = "TEMPLATE.yaml"
)

// Config enables tests by allowing them to work in an test directory
type Config struct {
	rootDir      string
	ChangelogMD  string
	ChloggenDir  string
	TemplateYAML string
}

type Option func(*Config)

func WithChloggenDir(chloggenDir string) Option {
	return func(ctx *Config) {
		ctx.ChloggenDir = filepath.Join(ctx.rootDir, chloggenDir)
		ctx.TemplateYAML = filepath.Join(ctx.rootDir, chloggenDir, DefaultTemplateYAML)
	}
}

func New(rootDir string, options ...Option) Config {
	ctx := Config{
		rootDir:      rootDir,
		ChangelogMD:  filepath.Join(rootDir, DefaultChangelogMD),
		ChloggenDir:  filepath.Join(rootDir, DefaultChloggenDir),
		TemplateYAML: filepath.Join(rootDir, DefaultChloggenDir, DefaultTemplateYAML),
	}
	for _, op := range options {
		op(&ctx)
	}
	return ctx
}
