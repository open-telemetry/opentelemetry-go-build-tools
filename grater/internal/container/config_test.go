// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package container

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewExecuteCommandConfig(t *testing.T) {
	tests := []struct {
		options  []ExecuteCommandOption
		expected ExecuteCommandConfig
	}{
		{
			[]ExecuteCommandOption{},
			ExecuteCommandConfig{
				cmd: []string{},
			},
		},
		{
			[]ExecuteCommandOption{
				WithContainerID("test-container"),
			},
			ExecuteCommandConfig{
				containerID: "test-container",
				cmd:         []string{},
			},
		},
		{
			[]ExecuteCommandOption{
				WithContainerID("test-container"),
				WithContainerID("test-container-2"),
			},
			ExecuteCommandConfig{
				containerID: "test-container-2",
				cmd:         []string{},
			},
		},
		{
			[]ExecuteCommandOption{
				WithCommand([]string{"echo", "hello"}),
			},
			ExecuteCommandConfig{
				cmd: []string{"echo", "hello"},
			},
		},
		{
			[]ExecuteCommandOption{
				WithCommand([]string{"echo", "hello"}),
				WithCommand([]string{"echo", "world"}),
			},
			ExecuteCommandConfig{
				cmd: []string{"echo", "world"},
			},
		},
		{
			[]ExecuteCommandOption{
				WithContainerID("test-container"),
				WithCommand([]string{"echo", "hello"}),
			},
			ExecuteCommandConfig{
				containerID: "test-container",
				cmd:         []string{"echo", "hello"},
			},
		},
	}
	for _, test := range tests {
		assert.Equal(t, test.expected, NewExecuteCommandConfig(test.options...))
	}
}

func TestNewCreateVolumeConfig(t *testing.T) {
	tests := []struct {
		options  []CreateVolumeOption
		expected CreateVolumeConfig
	}{
		{
			[]CreateVolumeOption{},
			CreateVolumeConfig{},
		},
		{
			[]CreateVolumeOption{
				WithVolumeName("test-volume"),
			},
			CreateVolumeConfig{
				volumeName: "test-volume",
			},
		},
		{
			[]CreateVolumeOption{
				WithVolumeName("test-volume"),
				WithVolumeName("test-volume-2"),
			},
			CreateVolumeConfig{
				volumeName: "test-volume-2",
			},
		},
	}
	for _, test := range tests {
		assert.Equal(t, test.expected, NewCreateVolumeConfig(test.options...))
	}
}

func TestNewUseContainerConfig(t *testing.T) {
	tests := []struct {
		options  []UseContainerOption
		expected UseContainerConfig
	}{
		{
			[]UseContainerOption{},
			UseContainerConfig{
				bindMounts:           map[string]string{},
				hostToContainerPaths: map[string]string{},
			},
		},
		{
			[]UseContainerOption{
				WithImageName("test-image"),
			},
			UseContainerConfig{
				imageName:            "test-image",
				bindMounts:           map[string]string{},
				hostToContainerPaths: map[string]string{},
			},
		},
		{
			[]UseContainerOption{
				WithImageName("test-image"),
				WithImageName("test-image-2"),
			},
			UseContainerConfig{
				imageName:            "test-image-2",
				bindMounts:           map[string]string{},
				hostToContainerPaths: map[string]string{},
			},
		},
		{
			[]UseContainerOption{
				WithBindMounts(map[string]string{"test-src": "/data/test-src"}),
			},
			UseContainerConfig{
				bindMounts:           map[string]string{"test-src": "/data/test-src"},
				hostToContainerPaths: map[string]string{},
			},
		},
		{
			[]UseContainerOption{
				WithHostToContainerPaths(map[string]string{"./testdata": "/data/testdata"}),
			},
			UseContainerConfig{
				bindMounts:           map[string]string{},
				hostToContainerPaths: map[string]string{"./testdata": "/data/testdata"},
			},
		},
		{
			[]UseContainerOption{
				WithImageName("test-image"),
				WithBindMounts(map[string]string{"test-src": "/data/test-src"}),
				WithHostToContainerPaths(map[string]string{"./testdata": "/data/testdata"}),
			},
			UseContainerConfig{
				imageName:            "test-image",
				bindMounts:           map[string]string{"test-src": "/data/test-src"},
				hostToContainerPaths: map[string]string{"./testdata": "/data/testdata"},
			},
		},
	}
	for _, test := range tests {
		assert.Equal(t, test.expected, NewUseContainerConfig(test.options...))
	}
}
