// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.opentelemetry.io/build-tools/githubgen/datatype"
)

func Test_chloggencomponents_verifyComponentsList(t *testing.T) {
	tests := []struct {
		name     string
		data     datatype.GithubData
		expected datatype.ChloggenConfig
	}{
		{
			name: "happy path",
			data: datatype.GithubData{
				Components: map[string]datatype.Metadata{
					"otlp": {
						Type:   "otlp",
						Status: &datatype.Status{Class: "receiver"},
					},
					"foo": {
						Type:   "foo",
						Status: &datatype.Status{Class: "processor"},
					},
					"exporterfoo": {
						Type:   "foo",
						Status: &datatype.Status{Class: "exporter"},
					},
					"barext": {
						Type:   "bar",
						Status: &datatype.Status{Class: "extension"},
					},
					"conn": {
						Type:   "count",
						Status: &datatype.Status{Class: "connector"},
					},
					"pkg": {
						Type:   "mypkg",
						Status: &datatype.Status{Class: "pkg"},
					},
					"no status": {
						Type: "no_status",
					},
				},
				Chloggen: datatype.ChloggenConfig{
					ChangeLogs:        nil,
					DefaultChangeLogs: nil,
					EntriesDir:        "",
					TemplateYAML:      "",
					SummaryTemplate:   "",
					Components:        []string{},
				},
			},
			expected: datatype.ChloggenConfig{Components: []string{"all", "connector/count", "exporter/foo", "extension/bar", "no_status", "pkg/mypkg", "processor/foo", "receiver/otlp"}},
		},
		{
			name: "respect existing config",
			data: datatype.GithubData{
				Components: map[string]datatype.Metadata{
					"otlp": {
						Type:   "otlp",
						Status: &datatype.Status{Class: "receiver"},
					},
				},
				Chloggen: datatype.ChloggenConfig{
					ChangeLogs:        nil,
					DefaultChangeLogs: []string{"api"},
					EntriesDir:        "",
					TemplateYAML:      "",
					SummaryTemplate:   "",
					Components:        []string{},
				},
			},
			expected: datatype.ChloggenConfig{
				DefaultChangeLogs: []string{"api"},
				Components:        []string{"all", "receiver/otlp"}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cg := &chloggenComponentsGenerator{
				writeComponents: func(_ string, cfg datatype.ChloggenConfig) error {
					assert.Equal(t, tt.expected, cfg)
					return nil
				},
			}
			err := cg.Generate(tt.data)
			require.NoError(t, err)
		})
	}
}
