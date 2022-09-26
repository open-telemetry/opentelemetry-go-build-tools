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
	"runtime"
)

const (
	changelogMD   = "CHANGELOG.md"
	unreleasedDir = "unreleased"
	templateYAML  = "TEMPLATE.yaml"
)

// Context enables tests by allowing them to work in an test directory
type Context struct {
	ChangelogMD   string
	UnreleasedDir string
	TemplateYAML  string
}

func New(rootDir string) Context {
	return Context{
		ChangelogMD:   filepath.Join(rootDir, changelogMD),
		UnreleasedDir: filepath.Join(rootDir, unreleasedDir),
		TemplateYAML:  filepath.Join(rootDir, unreleasedDir, templateYAML),
	}
}

var DefaultCtx = New(repoRoot())

func repoRoot() string {
	dir, err := os.Getwd()
	if err != nil {
		// This is not expected, but just in case
		fmt.Println("FAIL: Could not determine current working directory")
	}
	return dir
}

func moduleDir() string {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		// This is not expected, but just in case
		fmt.Println("FAIL: Could not determine module directory")
	}
	return filepath.Dir(filename)
}
