package main

import (
	"testing"

	"go.opentelemetry.io/build-tools/githubgen/datatype"
)

func Test_codeownersGenerator_verifyCodeOwnerOrgMembership(t *testing.T) {
	type fields struct {
		skipGithub bool
	}
	type args struct {
		allowlistData []byte
		data          datatype.GithubData
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cg := &codeownersGenerator{
				skipGithub: tt.fields.skipGithub,
			}
			if err := cg.verifyCodeOwnerOrgMembership(tt.args.allowlistData, tt.args.data); (err != nil) != tt.wantErr {
				t.Errorf("verifyCodeOwnerOrgMembership() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
