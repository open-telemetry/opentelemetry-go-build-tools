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
	"fmt"
	"os"
	"path/filepath"
)

const (
	DefaultChangelogMD  = "CHANGELOG.md"
	DefaultChloggenDir  = ".chloggen"
	DefaultTemplateYAML = "TEMPLATE.yaml"
)

// Context enables tests by allowing them to work in an test directory
type Context struct {
	rootDir      string
	ChangelogMD  string
	ChloggenDir  string
	TemplateYAML string
}

type Option func(*Context)

func WithChloggenDir(chloggenDir string) Option {
	return func(ctx *Context) {
		ctx.ChloggenDir = filepath.Join(ctx.rootDir, chloggenDir)
		ctx.TemplateYAML = filepath.Join(ctx.rootDir, chloggenDir, DefaultTemplateYAML)
	}
}

func New(rootDir string, options ...Option) Context {
	ctx := Context{
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

func RepoRoot() string {
	dir, err := os.Getwd()
	if err != nil {
		// This is not expected, but just in case
		fmt.Println("FAIL: Could not determine current working directory")
	}
	return dir
}
