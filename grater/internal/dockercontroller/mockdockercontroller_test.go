// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

//go:build integration
// +build integration

package dockercontroller

import (
	"testing"

	dockercontainer "github.com/docker/docker/api/types/container"
	"github.com/stretchr/testify/assert"
)

func TestMockCreateVolume(t *testing.T) {
	m := NewMockDockerController()

	called := false

	m.CreateVolumeMock = func(name string) (func(), error) {
		called = true
		assert.Equal(t, "test-volume", name)
		return func() {}, nil
	}

	cleanup, err := m.CreateVolume("test-volume")
	assert.NoError(t, err)
	assert.True(t, called)

	defer cleanup()
}

func TestMockUseContainer(t *testing.T) {
	m := NewMockDockerController()

	m.UseContainerMock = func(image string, vols []string) (string, func(), error) {
		assert.Equal(t, "test-image", image)
		assert.Equal(t, []string{"test-volume"}, vols)
		return "container-id", func() {}, nil
	}

	containerID, cleanup, err := m.UseContainer("test-image", []string{"test-volume"})
	assert.NoError(t, err)

	assert.Equal(t, "container-id", containerID)
	defer cleanup()
}

func TestMockExecuteCommand(t *testing.T) {
	m := NewMockDockerController()

	m.ExecuteCommandMock = func(id string, cmd []string) (string, dockercontainer.ExecInspect, error) {
		assert.Equal(t, "container-id", id)
		assert.Equal(t, []string{"ls", "-la"}, cmd)
		return "output", dockercontainer.ExecInspect{ExitCode: 0}, nil
	}

	output, execInspect, err := m.ExecuteCommand("container-id", []string{"ls", "-la"})
	assert.NoError(t, err)

	assert.Equal(t, "output", output)
	assert.Equal(t, 0, execInspect.ExitCode)
}
