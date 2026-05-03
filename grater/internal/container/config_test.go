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
				binds:     []string{},
				hostPaths: []string{},
			},
		},
		{
			[]UseContainerOption{
				WithImageName("test-image"),
			},
			UseContainerConfig{
				imageName: "test-image",
				binds:     []string{},
				hostPaths: []string{},
			},
		},
		{
			[]UseContainerOption{
				WithImageName("test-image"),
				WithImageName("test-image-2"),
			},
			UseContainerConfig{
				imageName: "test-image-2",
				binds:     []string{},
				hostPaths: []string{},
			},
		},
		{
			[]UseContainerOption{
				WithBinds([]string{"test-bind"}),
			},
			UseContainerConfig{
				binds:     []string{"test-bind"},
				hostPaths: []string{},
			},
		},
		{
			[]UseContainerOption{
				WithHostPaths([]string{"test-host-path"}),
			},
			UseContainerConfig{
				binds:     []string{},
				hostPaths: []string{"test-host-path"},
			},
		},
		{
			[]UseContainerOption{
				WithImageName("test-image"),
				WithBinds([]string{"test-bind"}),
				WithHostPaths([]string{"test-host-path"}),
			},
			UseContainerConfig{
				imageName: "test-image",
				binds:     []string{"test-bind"},
				hostPaths: []string{"test-host-path"},
			},
		},
	}
	for _, test := range tests {
		assert.Equal(t, test.expected, NewUseContainerConfig(test.options...))
	}
}