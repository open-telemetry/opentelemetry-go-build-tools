package main

import (
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
				generators:        fake.MockGenerator{},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := run(tt.args.folder, tt.args.allowlistFilePath, []datatype.Generator{&tt.args.generators}); (err != nil) != tt.wantErr {
				t.Errorf("run() error = %v, wantErr %v", err, tt.wantErr)
			}
			require.Equal(t, tt.args.generators.GenerateCalls(), 1)
		})
	}
}
