package check

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIsComponentImport(t *testing.T) {
	type args struct {
		importStr             string
		importPrefixesToCheck []string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "Match",
			args: args{
				importStr: "matching/prefix",
				importPrefixesToCheck: []string{
					"some/prefix",
					"matching/prefix",
				},
			},
			want: true,
		},
		{
			name: "No match",
			args: args{
				importStr: "some/prefix",
				importPrefixesToCheck: []string{
					"expecting/prefix",
				},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsComponentImport(tt.args.importStr, tt.args.importPrefixesToCheck); got != tt.want {
				t.Errorf("isComponentImport() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetImportPrefixesToCheck(t *testing.T) {
	tests := []struct {
		name   string
		module string
		want   []string
	}{
		{
			name:   "Get import prefixes - 1",
			module: "test",
			want: []string{
				"test/extension",
				"test/receiver",
				"test/processor",
				"test/exporter",
			},
		},
		{
			name:   "Get import prefixes - 2",
			module: "test/",
			want: []string{
				"test/extension",
				"test/receiver",
				"test/processor",
				"test/exporter",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetImportPrefixesToCheck(tt.module); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getImportPrefixesToCheck() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCheckFile(t *testing.T) {
	type args struct {
		projectPath                   string
		relativeDefaultComponentsPath string
		projectGoModule               string
		filename                      string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Invalid project path",
			args: args{
				projectPath:                   "invalid/project",
				relativeDefaultComponentsPath: "invalid/file",
				projectGoModule:               "go.opentelemetry.io/collector",
			},
			wantErr: true,
		},
		{
			name: "Invalid files",
			args: args{
				projectPath:                   getProjectPath(t),
				relativeDefaultComponentsPath: "service/defaultcomponents/invalid.go",
				projectGoModule:               "go.opentelemetry.io/collector",
			},
			wantErr: true,
		},
		{
			name: "Invalid imports",
			args: args{
				projectPath:                   getProjectPath(t),
				relativeDefaultComponentsPath: "component/componenttest/testdata/invalid_go.txt",
				projectGoModule:               "go.opentelemetry.io/collector",
			},
			wantErr: true,
		},
		{
			name: "README does not exist",
			args: args{
				projectPath:                   getProjectPath(t),
				relativeDefaultComponentsPath: "component/componenttest/testdata/valid_go.txt",
				projectGoModule:               "go.opentelemetry.io/collector",
				filename:                      "README.md",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := CheckFile(tt.args.projectPath, tt.args.relativeDefaultComponentsPath, tt.args.projectGoModule, tt.args.filename); (err != nil) != tt.wantErr {
				t.Errorf("checkDocs() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func getProjectPath(t *testing.T) string {
	wd, err := os.Getwd()
	require.NoError(t, err, "failed to get working directory: %v")

	// Absolute path to the project root directory
	projectPath := filepath.Join(wd, "../../")

	return projectPath
}
