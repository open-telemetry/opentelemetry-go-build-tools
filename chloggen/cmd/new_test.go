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
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"go.opentelemetry.io/build-tools/chloggen/internal/chlog"
	"go.opentelemetry.io/build-tools/chloggen/internal/config"
)

const newUsage = `Usage:
  chloggen new [flags]

Flags:
  -f, --filename string   name of the file to add
  -h, --help              help for new

Global Flags:
      --config string   (optional) chloggen config file`

func TestNewErr(t *testing.T) {
	var out, err string

	out, err = runCobra(t, "new", "--help")
	assert.Contains(t, out, newUsage)
	assert.Empty(t, err)

	out, err = runCobra(t, "new")
	assert.Contains(t, out, newUsage)
	assert.Contains(t, err, `required flag(s) "filename" not set`)

	out, err = runCobra(t, "new", "--filename", "my-change")
	assert.Contains(t, out, newUsage)
	switch {
	case strings.Contains(err, "cannot find the path specified"):
		// Windows returns a different error message
		assert.Contains(t, err, "cannot find the path specified")
	default:
		assert.Contains(t, err, "no such file or directory")
	}
}

func TestNew(t *testing.T) {
	globalCfg = config.New(t.TempDir())
	setupTestDir(t, []*chlog.Entry{})

	var out, err string

	out, err = runCobra(t, "new", "--filename", "my-change")
	assert.Contains(t, out, fmt.Sprintf("Changelog entry template copied to: %s", filepath.Join(globalCfg.EntriesDir, "my-change.yaml")))
	assert.Empty(t, err)

	out, err = runCobra(t, "new", "--filename", "some-change.yaml")
	assert.Contains(t, out, fmt.Sprintf("Changelog entry template copied to: %s", filepath.Join(globalCfg.EntriesDir, "some-change.yaml")))
	assert.Empty(t, err)

	out, err = runCobra(t, "new", "--filename", "some-change.yml")
	assert.Contains(t, out, fmt.Sprintf("Changelog entry template copied to: %s", filepath.Join(globalCfg.EntriesDir, "some-change.yaml")))
	assert.Empty(t, err)

	out, err = runCobra(t, "new", "--filename", "replace/forward/slash")
	assert.Contains(t, out, fmt.Sprintf("Changelog entry template copied to: %s", filepath.Join(globalCfg.EntriesDir, "replace_forward_slash.yaml")))
	assert.Empty(t, err)

	out, err = runCobra(t, "new", "--filename", "not.an.extension")
	assert.Contains(t, out, fmt.Sprintf("Changelog entry template copied to: %s", filepath.Join(globalCfg.EntriesDir, "not.an.extension.yaml")))
	assert.Empty(t, err)

	out, err = runCobra(t, "new", "--filename", "my-change.txt")
	assert.Contains(t, out, fmt.Sprintf("Changelog entry template copied to: %s", filepath.Join(globalCfg.EntriesDir, "my-change.txt.yaml")))
	assert.Empty(t, err)
}

func TestCleanFilename(t *testing.T) {
	assert.Equal(t, "fix_some_bug", cleanFileName("fix/some_bug"))
	assert.Equal(t, "fix_some_bug", cleanFileName("fix\\some_bug"))
}
