// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package mockcontainer

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/build-tools/grater/internal/container"
)

func TestMockCreateVolume(t *testing.T) {
	m := NewMockDockerContainer()

	called := false

	m.CreateVolumeMock = func(cfg container.CreateVolumeConfig) (container.CreateVolumeResponse, error) {
		called = true
		return container.CreateVolumeResponse{}, nil
	}

	_, err := m.CreateVolume(container.NewCreateVolumeConfig(
		container.WithVolumeName("test-volume"),
	))
	assert.NoError(t, err)
	assert.True(t, called)
}

func TestMockUseContainer(t *testing.T) {
	m := NewMockDockerContainer()

	called := false

	m.UseContainerMock = func(cfg container.UseContainerConfig) (container.UseContainerResponse, error) {
		called = true
		return container.UseContainerResponse{}, nil
	}

	_, err := m.UseContainer(container.NewUseContainerConfig(
		container.WithImageName("test-image"),
		container.WithBinds([]string{"test-volume"}),
		container.WithHostPaths([]string{"./testdata"}),
	))
	assert.NoError(t, err)
	assert.True(t, called)
}

func TestMockExecuteCommand(t *testing.T) {
	m := NewMockDockerContainer()

	called := false

	m.ExecuteCommandMock = func(cfg container.ExecuteCommandConfig) (container.ExecuteCommandResponse, error) {
		called = true
		return container.ExecuteCommandResponse{}, nil
	}

	_, err := m.ExecuteCommand(container.NewExecuteCommandConfig(
		container.WithContainerID("container-id"),
		container.WithCommand([]string{"ls", "-la"}),
	))
	assert.NoError(t, err)
	assert.True(t, called)
}
