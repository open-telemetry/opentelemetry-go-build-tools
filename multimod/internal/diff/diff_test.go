// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package diff

import (
	"errors"
	"path/filepath"
	"testing"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/stretchr/testify/require"

	"go.opentelemetry.io/build-tools/internal/repo"
	"go.opentelemetry.io/build-tools/multimod/internal/common"
)

func TestNormalizeVersion(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		output string
	}{
		{
			name:   "version not prefixed",
			input:  "0.1.2",
			output: "v0.1.2",
		},
		{
			name:   "version prefixed",
			input:  "v0.2.2",
			output: "v0.2.2",
		},
	}
	for _, tt := range tests {
		require.Equal(t, tt.output, normalizeVersion(tt.input))
	}
}

func TestNormalizeTag(t *testing.T) {
	tests := []struct {
		name   string
		input  common.ModuleTagName
		output string
	}{
		{
			name:   "repo root tag",
			input:  common.RepoRootTag,
			output: "v1.2.3",
		},
		{
			name:   "version prefixed",
			input:  "modset",
			output: "modset/v1.2.3",
		},
	}
	for _, tt := range tests {
		require.Equal(t, tt.output, normalizeTag(tt.input, "v1.2.3"))
	}
}

func TestHasChanged(t *testing.T) {
	tests := []struct {
		name         string
		tag          string
		modset       string
		versionsFile string
		repoRoot     string
		err          error
	}{
		{
			name:     "invalid repoRoot",
			err:      errors.New("repository does not exist"),
			tag:      "v0.8.0",
			modset:   "tools",
			repoRoot: "invalid-repo-root",
		},
		{
			name:   "invalid tag",
			err:    errors.New("tag not found"),
			tag:    "1.2.3",
			modset: "tools",
		},
		{
			name:   "invalid modset",
			err:    errors.New("could not find module set"),
			tag:    "v0.8.0",
			modset: "invalid",
		},
		{
			name:         "invalid versions file",
			err:          errors.New("no such file or directory"),
			tag:          "v0.8.0",
			versionsFile: "invalid.yaml",
			modset:       "tools",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var repoRoot, versionFile string
			var err error

			if len(tt.repoRoot) > 0 {
				repoRoot = tt.repoRoot
			} else {
				repoRoot, err = repo.FindRoot()
				require.NoError(t, err)
			}

			if len(tt.versionsFile) > 0 {
				versionFile = filepath.Join(repoRoot, tt.versionsFile)
			} else {
				versionFile = filepath.Join(repoRoot, "versions.yaml")
			}
			changedFiles, err := HasChanged(repoRoot, versionFile, tt.tag, tt.modset)
			if tt.err != nil {
				require.Error(t, err)
				require.ErrorContains(t, err, tt.err.Error())
			} else {
				require.NoError(t, err)
			}
			require.False(t, len(changedFiles) > 0)
		})
	}
}

type MockClient struct {
	files         []string
	headCommitErr error
	tagCommitErr  error
}

func (c MockClient) HeadCommit(_ *git.Repository) (*object.Commit, error) {
	return nil, c.headCommitErr
}
func (c MockClient) TagCommit(_ *git.Repository, _ string) (*object.Commit, error) {
	return nil, c.tagCommitErr
}
func (c MockClient) FilesChanged(_ *object.Commit, _ *object.Commit, _ string, _ string) ([]string, error) {
	return c.files, nil
}

func TestFilesChanged(t *testing.T) {
	tests := []struct {
		name     string
		version  string
		modset   string
		tagNames []common.ModuleTagName
		cli      MockClient
		expected bool
		err      error
	}{
		{
			name:    "error with head commit",
			version: "v0.8.0",
			modset:  "tools",
			cli: MockClient{
				headCommitErr: object.ErrEntryNotFound,
			},
			err: object.ErrEntryNotFound,
		},
		{
			name: "tag missing",
			cli: MockClient{
				tagCommitErr: git.ErrTagNotFound,
			},
			tagNames: []common.ModuleTagName{"tools"},
			version:  "v0.8.0",
			modset:   "tools",
			err:      git.ErrTagNotFound,
		},
		{
			name:    "changes found in to, tag exists",
			version: "v0.8.0",
			modset:  "tools",
			cli: MockClient{
				files: []string{"file1.go"},
			},
			tagNames: []common.ModuleTagName{"tools"},
			expected: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			changedFiles, err := filesChanged(nil, tt.modset, tt.version, tt.tagNames, tt.cli)
			if tt.err != nil {
				require.Error(t, err)
				require.ErrorContains(t, err, tt.err.Error())
			} else {
				require.NoError(t, err)
			}
			require.Equal(t, len(changedFiles) > 0, tt.expected)
		})
	}
}
