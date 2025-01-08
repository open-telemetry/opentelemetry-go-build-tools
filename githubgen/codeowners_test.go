package main

import (
	"testing"

	"go.opentelemetry.io/build-tools/githubgen/datatype"
)

func Test_codeownersGenerator_verifyCodeOwnerOrgMembership(t *testing.T) {
	type args struct {
		allowlistData []byte
		data          datatype.GithubData
	}
	tests := []struct {
		name       string
		skipGithub bool
		args       args
		wantErr    bool
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
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cg := &codeownersGenerator{
				skipGithub: tt.skipGithub,
			}
			if err := cg.verifyCodeOwnerOrgMembership(tt.args.allowlistData, tt.args.data); (err != nil) != tt.wantErr {
				t.Errorf("verifyCodeOwnerOrgMembership() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
