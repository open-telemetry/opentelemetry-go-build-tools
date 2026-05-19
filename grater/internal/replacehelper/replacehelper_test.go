// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

// Package replacehelper provides utilities for working with replacements in the .grater directory.
package replacehelper

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.opentelemetry.io/build-tools/grater/internal/module"
	"go.opentelemetry.io/build-tools/grater/internal/workspace"
)

const fileReadWrite = 0644

func TestReplace(t *testing.T) {
	testDir := t.TempDir()
	t.Chdir(testDir)

	ws, err := workspace.NewWorkspace()
	require.NoError(t, err)

	err = Replace(ws, []string{
		"foo/bar@v1.0.0 baz/qux@v2.0.0",
		"abc/def xyz/pqr@v3.0.0",
	})
	require.NoError(t, err)

	replacements := ws.GetReplacements()

	assert.ElementsMatch(t, replacements, [][]module.Module{
		{
			*module.NewModule("foo/bar", "v1.0.0"),
			*module.NewModule("baz/qux", "v2.0.0"),
		},
		{
			*module.NewModule("abc/def", ""),
			*module.NewModule("xyz/pqr", "v3.0.0"),
		},
	})

	content, err := os.ReadFile(".grater/replacements.json")
	require.NoError(t, err)

	assert.JSONEq(t,
		`[
			[
				{
					"module_name":"bar",
					"module_path":"foo/bar",
					"module_version":"v1.0.0"
				},
				{
					"module_name":"qux",
					"module_path":"baz/qux",
					"module_version":"v2.0.0"
				}
			],
			[
				{
					"module_name":"def",
					"module_path":"abc/def",
					"module_version":""
				},
				{
					"module_name":"pqr",
					"module_path":"xyz/pqr",
					"module_version":"v3.0.0"
				}
			]
		]`,
		string(content),
	)
}

func TestReplaceFromFile(t *testing.T) {
	testDir := t.TempDir()
	t.Chdir(testDir)

	ws, err := workspace.NewWorkspace()
	require.NoError(t, err)

	err = os.WriteFile(
		"replacements.txt",
		[]byte("foo/bar baz/qux@v1.0.0\nabc/def@v2.0.0 xyz/pqr\n"),
		fileReadWrite,
	)
	require.NoError(t, err)

	err = AddFromFile(ws, "replacements.txt")
	require.NoError(t, err)

	replacements := ws.GetReplacements()

	assert.ElementsMatch(t, replacements, [][]module.Module{
		{
			*module.NewModule("foo/bar", ""),
			*module.NewModule("baz/qux", "v1.0.0"),
		},
		{
			*module.NewModule("abc/def", "v2.0.0"),
			*module.NewModule("xyz/pqr", ""),
		},
	})

	content, err := os.ReadFile(".grater/replacements.json")
	require.NoError(t, err)

	assert.JSONEq(t,
		`[
			[
				{
					"module_name":"bar",
					"module_path":"foo/bar",
					"module_version":""
				},
				{
					"module_name":"qux",
					"module_path":"baz/qux",
					"module_version":"v1.0.0"
				}
			],
			[
				{
					"module_name":"def",
					"module_path":"abc/def",
					"module_version":"v2.0.0"
				},
				{
					"module_name":"pqr",
					"module_path":"xyz/pqr",
					"module_version":""
				}
			]
		]`,
		string(content),
	)
}

func TestReplaceFromFileFails(t *testing.T) {
	testDir := t.TempDir()
	t.Chdir(testDir)

	ws, err := workspace.NewWorkspace()
	require.NoError(t, err)

	err = AddFromFile(ws, "non_existent.txt")
	require.Error(t, err)

	assert.Contains(t, err.Error(), "failed to read replacements from file")
}

func TestReplaceFailsInvalidFormat(t *testing.T) {
	testDir := t.TempDir()
	t.Chdir(testDir)

	ws, err := workspace.NewWorkspace()
	require.NoError(t, err)

	err = Replace(ws, []string{
		"foo/bar@v1.0.0",
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid replacement format")
}