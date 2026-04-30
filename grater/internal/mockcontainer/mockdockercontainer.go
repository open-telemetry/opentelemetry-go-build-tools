// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

// Package mockcontainer provides a mock implementation of the Container interface.
package mockcontainer

import "go.opentelemetry.io/build-tools/grater/internal/container"

// MockDockerContainer is a mock implementation of the Container interface.
type MockDockerContainer struct {
	CreateVolumeMock   func(string) (func(), error)
	UseContainerMock   func(string, []string, []string) (string, func(), error)
	ExecuteCommandMock func(string, []string) (string, int, error)
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
func (m *MockDockerContainer) UseContainer(imageName string, volumeNames, localPaths []string) (string, func(), error) {
	if m.UseContainerMock != nil {
		return m.UseContainerMock(imageName, volumeNames, localPaths)
	}
	return "", func() {}, nil
}

// ExecuteCommand creates a mock instance of executing a command.
func (m *MockDockerContainer) ExecuteCommand(containerID string, cmd []string) (string, int, error) {
	if m.ExecuteCommandMock != nil {
		return m.ExecuteCommandMock(containerID, cmd)
	}
	return "", 0, nil
}
