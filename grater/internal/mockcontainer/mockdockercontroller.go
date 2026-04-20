// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

// Package mockcontainer provides a mock implementation of the DockerController interface.
package mockcontainer

import (
	"github.com/moby/moby/client"

	"go.opentelemetry.io/build-tools/grater/internal/container"
)

// MockDockerContainer is a mock implementation of the DockerContainer interface.
type MockDockerContainer struct {
	CreateVolumeMock   func(string) (func(), error)
	UseContainerMock   func(string, []string) (string, func(), error)
	ExecuteCommandMock func(string, []string) (string, client.ExecInspectResult, error)
}

var _ container.Container = (*MockDockerContainer)(nil)

// NewMockDockerContainer creates a new instance of MockDockerContainer.
func NewMockDockerContainer() *MockDockerContainer {
	return &MockDockerContainer{}
}

// CreateVolume creates a mock instance of Volume.
func (m *MockDockerContainer) CreateVolume(volumeName string) (func(), error) {
	if m.CreateVolumeMock != nil {
		return m.CreateVolumeMock(volumeName)
	}
	return func() {}, nil
}

// UseContainer creates a mock instance of Container.
func (m *MockDockerContainer) UseContainer(imageName string, volumeNames []string) (string, func(), error) {
	if m.UseContainerMock != nil {
		return m.UseContainerMock(imageName, volumeNames)
	}
	return "", func() {}, nil
}

// ExecuteCommand creates a mock instance of executing a command.
func (m *MockDockerContainer) ExecuteCommand(containerID string, cmd []string) (string, client.ExecInspectResult, error) {
	if m.ExecuteCommandMock != nil {
		return m.ExecuteCommandMock(containerID, cmd)
	}
	return "", client.ExecInspectResult{}, nil
}
