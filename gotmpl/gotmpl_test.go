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

package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	bodyPath = "testdata/in.go.tmpl"
	jsonData = `{ "pkg": "myname" }`
)

func Test(t *testing.T) {
	want, err := os.ReadFile("testdata/want.go")
	require.NoError(t, err)
	wantText := string(want)

	destPath := filepath.Join(t.TempDir(), "out.go")

	err = gotmpl(bodyPath, jsonData, destPath)
	require.NoError(t, err)

	got, err := os.ReadFile(destPath) //nolint:gosec // It is a safe test temp filepath.
	require.NoError(t, err)
	gotText := string(got)

	assert.Equal(t, wantText, gotText)
}

func TestEmptyOut(t *testing.T) {
	err := gotmpl(bodyPath, jsonData, "")

	assert.ErrorContains(t, err, "gotmpl: output filepath must be set")
}

func TestMissingData(t *testing.T) {
	destPath := filepath.Join(t.TempDir(), "out.go")

	err := gotmpl(bodyPath, `{ "badkey": "val" }`, destPath)

	assert.ErrorContains(t, err, "gotmpl: execution failed")
}

func TestDataIsNotJSON(t *testing.T) {
	destPath := filepath.Join(t.TempDir(), "out.go")

	err := gotmpl(bodyPath, "{ bad[]", destPath)

	assert.ErrorContains(t, err, "gotmpl: data must be in JSON format")
}

func TestEmptyBodyPath(t *testing.T) {
	destPath := filepath.Join(t.TempDir(), "out.go")

	err := gotmpl("", jsonData, destPath)

	assert.ErrorContains(t, err, "gotmpl: template body filepath must be set")
}

func TestNotExistingBody(t *testing.T) {
	destPath := filepath.Join(t.TempDir(), "out.go")

	err := gotmpl("testdata/non-exiting.go", jsonData, destPath)

	assert.ErrorContains(t, err, "gotmpl: cannot parse template body file")
	assert.ErrorIs(t, err, os.ErrNotExist)
}

func TestComplex(t *testing.T) {
	wantText := `myapp v2.1.0`
	destPath := filepath.Join(t.TempDir(), "complex.txt")

	bodyPath := filepath.Join(t.TempDir(), "complex.txt.tmpl")
	body := `{{.app.name}} v{{.app.version.major}}.{{.app.version.minor}}.{{.app.version.patch}}`
	jsonData := `{"app":{"name": "myapp", "version": {"major": 2, "minor": 1, "patch": 0}}}`
	err := os.WriteFile(bodyPath, ([]byte)(body), 0o600)
	require.NoError(t, err)

	err = gotmpl(bodyPath, jsonData, destPath)
	require.NoError(t, err)

	got, err := os.ReadFile(destPath) //nolint:gosec // It is a safe test temp filepath.
	require.NoError(t, err)
	gotText := string(got)

	assert.Equal(t, wantText, gotText)
}
