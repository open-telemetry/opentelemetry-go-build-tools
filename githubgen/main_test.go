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
