// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"strings"
	"testing"

	"go.opentelemetry.io/build-tools/githubgen/datatype"
)

func Test_codeownersGenerator_verifyCodeOwnerOrgMembership(t *testing.T) {
	type args struct {
		allowlistData []byte
		data          datatype.GithubData
	}
	tests := []struct {
		name        string
		skipGithub  bool
		args        args
		wantErr     bool
		errContains string
	}{
		{
			name:       "happy path",
			skipGithub: true,
			args: args{
				allowlistData: []byte(""),
				data: datatype.GithubData{
					Codeowners: []string{
						"user1", "user2", "user3",
					},
				},
			},
			wantErr: false,
		},
		{
			name:       "codeowner is member but also on allowlist",
			skipGithub: true,
			args: args{
				allowlistData: []byte("user1"),
				data: datatype.GithubData{
					Codeowners: []string{
						"user1", "user2", "user3",
					},
				},
			},
			wantErr:     true,
			errContains: "codeowners members duplicate in allowlist",
		},
		{
			name:       "codeowner is not a member but exists in allowlist",
			skipGithub: true,
			args: args{
				allowlistData: []byte("user4"),
				data: datatype.GithubData{
					Codeowners: []string{
						"user4",
					},
				},
			},
			wantErr: false,
		},
		{
			name:       "codeowner is not a member and does not exist in allowlist",
			skipGithub: false,
			args: args{
				allowlistData: []byte(""),
				data: datatype.GithubData{
					Codeowners: []string{
						"user4",
					},
				},
			},
			wantErr:     true,
			errContains: "codeowners are not members",
		},
		{
			name:       "user in allowlist but is not a codeowner",
			skipGithub: true,
			args: args{
				allowlistData: []byte("user4\nuser5"),
				data: datatype.GithubData{
					Codeowners: []string{
						"user4",
					},
				},
			},
			wantErr:     true,
			errContains: "unused members in allowlist",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cg := &codeownersGenerator{
				skipGithub:       tt.skipGithub,
				getGitHubMembers: mockGithubMembers,
			}
			err := cg.verifyCodeOwnerOrgMembership(tt.args.allowlistData, tt.args.data)
			if (err != nil) != tt.wantErr && strings.Contains(err.Error(), tt.errContains) {
				t.Errorf("verifyCodeOwnerOrgMembership() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func mockGithubMembers(bool, string) (map[string]struct{}, error) {
	return map[string]struct{}{
		"user1": {},
		"user2": {},
		"user3": {},
	}, nil
}
