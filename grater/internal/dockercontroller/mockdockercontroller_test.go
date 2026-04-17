// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package dockercontroller

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/docker/docker/api/types/container"
)

func TestCreateVolume(t *testing.T) {
    m := NewMockDockerController()
    defer m.AssertExpectations(t)

    m.On("CreateVolume", "test-volume").Return(func() {}, nil)

    cleanup, err := m.CreateVolume("test-volume")
    assert.NoError(t, err)
    defer cleanup()
}

func TestUseContainer(t *testing.T) {
	m := NewMockDockerController()
	defer m.AssertExpectations(t)

	m.On("UseContainer", "test-image", []string{"test-volume"}).Return("container-id", func() {}, nil)

	containerID, cleanup, err := m.UseContainer("test-image", []string{"test-volume"})
	assert.NoError(t, err)
	defer cleanup()

	assert.Equal(t, "container-id", containerID)
}

func TestExecuteCommand(t *testing.T) {
	m := NewMockDockerController()
	defer m.AssertExpectations(t)

	m.On("ExecuteCommand", "container-id", []string{"ls", "-la"}).Return("output", container.ExecInspect{}, nil)

	output, execInspect, err := m.ExecuteCommand("container-id", []string{"ls", "-la"})
	assert.NoError(t, err)

	assert.Equal(t, "output", output)
	assert.Empty(t, execInspect)
}
