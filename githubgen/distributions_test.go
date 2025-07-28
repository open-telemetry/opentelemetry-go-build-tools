// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"testing"

	"github.com/stretchr/testify/require"

	"go.opentelemetry.io/build-tools/githubgen/datatype"
)

func Test_distributionsGenerator_Generate(t *testing.T) {
	tests := []struct {
		name              string
		data              datatype.GithubData
		wantErr           bool
		writeDistribution func(t *testing.T, rootFolder string, distName string, data distOutput) error
	}{
		{
			name: "basic test",
			data: datatype.GithubData{
				RootFolder: "some-folder",
				Components: map[string]datatype.Metadata{
					"some-component": {
						Type: "otlp",
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
			writeDistribution: func(_ *testing.T, _ string, _ string, _ distOutput) error {
				return nil
			},
		},
		{
			name: "none",
			data: datatype.GithubData{
				RootFolder: "some-folder",
				Components: map[string]datatype.Metadata{
					"some-component": {
						Type: "otlp",
						Status: &datatype.Status{
							Distributions: []string{"some-distro"},
							Class:         "exporter",
						},
					},
					"unused-component": {
						Type: "foo",
						Status: &datatype.Status{
							Class: "exporter",
						},
					},
				},
				Distributions: []datatype.DistributionData{
					{
						Name: "some-distro",
					},
					{
						Name: "unused",
						None: true,
					},
				},
			},
			wantErr: false,
			writeDistribution: func(t *testing.T, _ string, distName string, data distOutput) error {
				if distName == "unused" {
					require.Len(t, data.Components, 1)
					require.Equal(t, "foo", data.Components["exporter"][0])
				}
				return nil
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dg := &distributionsGenerator{
				writeDistribution: func(rootFolder string, distName string, distData distOutput) error {
					return tt.writeDistribution(t, rootFolder, distName, distData)
				},
			}
			if err := dg.Generate(tt.data); (err != nil) != tt.wantErr {
				t.Errorf("Generate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
