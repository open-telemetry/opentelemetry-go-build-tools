// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package main

import "testing"

func Test_folderToSlug(t *testing.T) {
	type args struct {
		folder string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "special case",
			args: args{
				folder: "internal/coreinternal",
			},
			want: "internal/core",
		},
		{
			name: "1",
			args: args{
				folder: "processor/resourcedetectionprocessor/internal/aws/ec2",
			},
			want: "processor/resourcedetection/internal/aws/ec2",
		},
		{
			name: "2",
			args: args{
				folder: "receiver/hostmetricsreceiver/internal/scraper/loadscraper",
			},
			want: "receiver/hostmetrics/internal/scraper/loadscraper",
		},
		{
			name: "3",
			args: args{
				folder: "receiver/apachesparkreceiver",
			},
			want: "receiver/apachespark",
		},
		{
			name: "4",
			args: args{
				folder: "testbed",
			},
			want: "testbed",
		},
		{
			name: "5",
			args: args{
				folder: "processor/resourcedetectionprocessor/internal/aws/ec2",
			},
			want: "processor/resourcedetection/internal/aws/ec2",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			itg := &issueTemplatesGenerator{
				trimSuffixes: []string{"receiver", "exporter", "extension", "processor", "connector", "internal"},
			}
			if got := itg.folderToSlug(tt.args.folder); got != tt.want {
				t.Errorf("folderToSlug() = %v, want %v", got, tt.want)
			}
		})
	}
}
