// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"reflect"
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

func Test_codeownersGenerator_longestNameSpaces(t *testing.T) {
	longName := "name-looooong"
	type args struct {
		data datatype.GithubData
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{
			name: "basic test",
			args: args{
				data: datatype.GithubData{
					Distributions: []datatype.DistributionData{
						{
							Name: "name-short",
						},
						{
							Name: longName,
						},
					},
				},
			},
			want: len(longName),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cg := &codeownersGenerator{}
			if got := cg.longestNameSpaces(tt.args.data); got != tt.want {
				t.Errorf("longestNameSpaces() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_formatGithubUser(t *testing.T) {
	type args struct {
		user string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "with @",
			args: args{
				user: "@some-user",
			},
			want: "@some-user",
		},
		{
			name: "without @",
			args: args{
				user: "some-user",
			},
			want: "@some-user",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := formatGithubUser(tt.args.user); got != tt.want {
				t.Errorf("formatGithubUser() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_injectContent(t *testing.T) {
	type args struct {
		startMagicString string
		endMagicString   string
		templateContents []byte
		replaceContent   []string
	}
	tests := []struct {
		name string
		args args
		want []byte
	}{
		{
			name: "successful replacement",
			args: args{
				startMagicString: "aaa",
				endMagicString:   "bbb",
				templateContents: []byte("aaa\n\n\nbbb"),
				replaceContent: []string{
					"ccc",
				},
			},
			want: []byte("aaa\n\nccc\n\nbbb"),
		},
		{
			name: "no replacement",
			args: args{
				startMagicString: "aaa",
				endMagicString:   "bbb",
				templateContents: []byte("aa\n\n\nbb"),
				replaceContent: []string{
					"ccc",
				},
			},
			want: []byte("aa\n\n\nbb"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := injectContent(tt.args.startMagicString, tt.args.endMagicString, tt.args.templateContents, tt.args.replaceContent); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("injectContent() = %s, want %s", got, tt.want)
			}
		})
	}
}

func Test_codeownersGenerator_Generate(t *testing.T) {
	type fields struct {
		skipGithub       bool
		getGitHubMembers func(skipGithub bool, githubOrg string) (map[string]struct{}, error)
		getFile          func(fileName string) ([]byte, error)
		setFile          func(fileName string, data []byte) error
	}
	type args struct {
		data datatype.GithubData
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "unmaintained/deprecated test",
			fields: fields{
				skipGithub:       true,
				getGitHubMembers: mockGithubMembers,
				getFile:          mockGetFile,
				setFile:          mockSetFile,
			},
			args: args{
				data: datatype.GithubData{
					RootFolder: "some-folder",
					Folders: []string{
						"folder1",
						"folder2",
					},
					Codeowners: []string{
						"user1",
					},
					AllowlistFilePath: "allowlist",
					MaxLength:         10,
					Components: map[string]datatype.Metadata{
						"folder1": {
							Status: &datatype.Status{
								Stability: map[string][]string{
									"deprecated": {""},
								},
								Distributions: nil,
								Class:         "",
								Codeowners:    nil,
							},
						},
						"folder2": {
							Status: &datatype.Status{
								Stability: map[string][]string{
									"unmaintained": {""},
								},
								Distributions: nil,
								Class:         "",
								Codeowners:    nil,
							},
						},
					},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cg := &codeownersGenerator{
				skipGithub:       tt.fields.skipGithub,
				getGitHubMembers: tt.fields.getGitHubMembers,
				getFile:          tt.fields.getFile,
				setFile:          tt.fields.setFile,
			}
			if err := cg.Generate(tt.args.data); (err != nil) != tt.wantErr {
				t.Errorf("Generate() error = %v, wantErr %v", err, tt.wantErr)
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

func mockGetFile(path string) ([]byte, error) {
	if path == "allowlist" {
		return []byte(""), nil
	} else if strings.Contains(path, ".github/CODEOWNERS") {
		return []byte("aaa\n\n\nbbb"), nil
	}
	return []byte(""), nil
}

func mockSetFile(string, []byte) error {
	return nil
}
