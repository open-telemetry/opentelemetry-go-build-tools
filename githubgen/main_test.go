// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"

	"go.opentelemetry.io/build-tools/githubgen/datatype"
	"go.opentelemetry.io/build-tools/githubgen/datatype/fake"
)

func Test_run(t *testing.T) {

	type args struct {
		folder            string
		allowlistFilePath string
		generators        fake.MockGenerator
		distributions     []datatype.DistributionData
		defaultCodeOwners string
		githubOrg         string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "codeowners",
			args: args{
				folder:            ".",
				allowlistFilePath: "cmd/githubgen/allowlist.txt",
				generators: fake.MockGenerator{
					GenerateFunc: func(_ datatype.GithubData) error {
						return nil
					},
				},
				distributions: []datatype.DistributionData{
					{
						Name:        "my-distro",
						URL:         "some-url",
						Maintainers: nil,
					},
				},
				defaultCodeOwners: "some-code-owners",
				githubOrg:         "some-org",
			},
			wantErr: false,
		},
		{
			name: "chloggen components",
			args: args{
				folder:            "testdata/chloggen_components",
				allowlistFilePath: "cmd/githubgen/allowlist.txt",
				generators: fake.MockGenerator{
					GenerateFunc: func(_ datatype.GithubData) error {
						return nil
					},
				},
				distributions: []datatype.DistributionData{
					{
						Name:        "my-distro",
						URL:         "some-url",
						Maintainers: nil,
					},
				},
				defaultCodeOwners: "some-code-owners",
				githubOrg:         "some-org",
			},
			wantErr: false,
		},
	}

	// nolint:govet
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := run(tt.args.folder, tt.args.allowlistFilePath, []datatype.Generator{&tt.args.generators}, tt.args.distributions, tt.args.defaultCodeOwners, tt.args.githubOrg); (err != nil) != tt.wantErr {
				t.Errorf("run() error = %v, wantErr %v", err, tt.wantErr)
			}
			require.Equal(t, len(tt.args.generators.GenerateCalls()), 1)
		})
	}
}

func Test_newIssueTemplatesGenerator(t *testing.T) {
	type args struct {
		trimSuffixes string
	}
	tests := []struct {
		name string
		args args
		want *issueTemplatesGenerator
	}{
		{
			name: "basic test",
			args: args{
				trimSuffixes: "one, two, three",
			},
			want: &issueTemplatesGenerator{
				trimSuffixes: []string{
					"one",
					"two",
					"three",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := newIssueTemplatesGenerator(tt.args.trimSuffixes); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("newIssueTemplatesGenerator() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_parseGenerators(t *testing.T) {
	skipCheck := false
	type args struct {
		args            []string
		trimSuffixes    string
		skipGithubCheck *bool
	}
	tests := []struct {
		name    string
		args    args
		wantLen int
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid single generator - issue-templates",
			args: args{
				args:            []string{"issue-templates"},
				trimSuffixes:    "receiver, exporter",
				skipGithubCheck: &skipCheck,
			},
			wantLen: 1,
			wantErr: false,
		},
		{
			name: "valid single generator - codeowners",
			args: args{
				args:            []string{"codeowners"},
				trimSuffixes:    "receiver, exporter",
				skipGithubCheck: &skipCheck,
			},
			wantLen: 1,
			wantErr: false,
		},
		{
			name: "valid single generator - distributions",
			args: args{
				args:            []string{"distributions"},
				trimSuffixes:    "receiver, exporter",
				skipGithubCheck: &skipCheck,
			},
			wantLen: 1,
			wantErr: false,
		},
		{
			name: "valid single generator - chloggen-components",
			args: args{
				args:            []string{"chloggen-components"},
				trimSuffixes:    "receiver, exporter",
				skipGithubCheck: &skipCheck,
			},
			wantLen: 1,
			wantErr: false,
		},
		{
			name: "valid multiple generators",
			args: args{
				args:            []string{"issue-templates", "codeowners", "distributions"},
				trimSuffixes:    "receiver, exporter",
				skipGithubCheck: &skipCheck,
			},
			wantLen: 3,
			wantErr: false,
		},
		{
			name: "all valid generators",
			args: args{
				args:            []string{"issue-templates", "codeowners", "distributions", "chloggen-components"},
				trimSuffixes:    "receiver, exporter",
				skipGithubCheck: &skipCheck,
			},
			wantLen: 4,
			wantErr: false,
		},
		{
			name: "no generators specified - should use defaults",
			args: args{
				args:            []string{},
				trimSuffixes:    "receiver, exporter",
				skipGithubCheck: &skipCheck,
			},
			wantLen: 2, // issue-templates and codeowners are defaults
			wantErr: false,
		},
		{
			name: "invalid generator",
			args: args{
				args:            []string{"invalid-generator"},
				trimSuffixes:    "receiver, exporter",
				skipGithubCheck: &skipCheck,
			},
			wantLen: 0,
			wantErr: true,
			errMsg:  "unknown generator",
		},
		{
			name: "mix of valid and invalid generators",
			args: args{
				args:            []string{"issue-templates", "invalid-generator"},
				trimSuffixes:    "receiver, exporter",
				skipGithubCheck: &skipCheck,
			},
			wantLen: 0,
			wantErr: true,
			errMsg:  "unknown generator",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseGenerators(tt.args.args, tt.args.trimSuffixes, tt.args.skipGithubCheck)

			if tt.wantErr {
				require.Error(t, err, "parseGenerators() should return an error")
				require.Contains(t, err.Error(), tt.errMsg, "error message should contain expected text")
				require.Nil(t, got, "generators should be nil on error")
			} else {
				require.NoError(t, err, "parseGenerators() should not return an error")
				require.Len(t, got, tt.wantLen, "unexpected number of generators")
				require.NotNil(t, got, "generators should not be nil")
			}
		})
	}
}
