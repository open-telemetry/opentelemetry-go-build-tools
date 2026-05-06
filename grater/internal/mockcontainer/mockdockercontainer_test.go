// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package mockcontainer

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"go.opentelemetry.io/build-tools/grater/internal/container"
)

func TestMockCreateVolume(t *testing.T) {
	m := NewMockDockerContainer()

	called := false

	m.CreateVolumeMock = func(_ context.Context, _ container.CreateVolumeConfig) (container.CreateVolumeResponse, error) {
		called = true
		return container.CreateVolumeResponse{}, nil
	}

	_, err := m.CreateVolume(context.Background(), container.NewCreateVolumeConfig(
		container.WithVolumeName("test-volume"),
	))
	assert.NoError(t, err)
	assert.True(t, called)
}

func TestMockUseContainer(t *testing.T) {
	m := NewMockDockerContainer()

	called := false

	m.UseContainerMock = func(_ context.Context, _ container.UseContainerConfig) (container.UseContainerResponse, error) {
		called = true
		return container.UseContainerResponse{}, nil
	}

	_, err := m.UseContainer(context.Background(), container.NewUseContainerConfig(
		container.WithImageName("test-image"),
		container.WithBindMounts(map[string]string{"test-volume": "/data/test-volume"}),
		container.WithHostToContainerPaths(map[string]string{"./testdata": "/data/testdata"}),
	))
	assert.NoError(t, err)
	assert.True(t, called)
}

func TestMockExecuteCommand(t *testing.T) {
	m := NewMockDockerContainer()

	called := false

	m.ExecuteCommandMock = func(_ context.Context, _ container.ExecuteCommandConfig) (container.ExecuteCommandResponse, error) {
		called = true
		return container.ExecuteCommandResponse{}, nil
	}

	_, err := m.ExecuteCommand(context.Background(), container.NewExecuteCommandConfig(
		container.WithContainerID("container-id"),
		container.WithCommand([]string{"ls", "-la"}),
	))
	assert.NoError(t, err)
	assert.True(t, called)
}
