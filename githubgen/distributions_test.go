package main

import (
	"testing"

	"go.opentelemetry.io/build-tools/githubgen/datatype"
)

func Test_distributionsGenerator_Generate(t *testing.T) {
	tests := []struct {
		name    string
		data    datatype.GithubData
		wantErr bool
	}{
		{
			name: "basic test",
			data: datatype.GithubData{
				RootFolder: "some-folder",
				Components: map[string]datatype.Metadata{
					"some-component": {
						Type:   "otlp",
						Parent: "",
						Status: &datatype.Status{
							Distributions: []string{"some-distro"},
							Class:         "exporter",
						},
					},
				},
				Distributions: []datatype.DistributionData{
					{
						Name: "some-distro",
					},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dg := &distributionsGenerator{
				writeDistribution: mockWriteDistribution,
			}
			if err := dg.Generate(tt.data); (err != nil) != tt.wantErr {
				t.Errorf("Generate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func mockWriteDistribution(string, string, distOutput) error {
	return nil
}
