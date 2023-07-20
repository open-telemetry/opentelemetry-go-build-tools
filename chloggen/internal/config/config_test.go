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
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	root := "/tmp"
	ctx := New(root)
	assert.Equal(t, root, ctx.rootDir)
	assert.Equal(t, filepath.Join(root, DefaultChloggenDir), ctx.ChloggenDir)
	assert.Equal(t, filepath.Join(root, DefaultChangelogMD), ctx.ChangelogMD)
	assert.Equal(t, filepath.Join(root, DefaultChloggenDir, DefaultTemplateYAML), ctx.TemplateYAML)
}

func TestWithChloggenDir(t *testing.T) {
	root := "/tmp"
	chloggenDir := ".test"
	ctx := New(root, WithChloggenDir(chloggenDir))
	assert.Equal(t, root, ctx.rootDir)
	assert.Equal(t, filepath.Join(root, chloggenDir), ctx.ChloggenDir)
	assert.Equal(t, filepath.Join(root, DefaultChangelogMD), ctx.ChangelogMD)
	assert.Equal(t, filepath.Join(root, chloggenDir, DefaultTemplateYAML), ctx.TemplateYAML)
}
