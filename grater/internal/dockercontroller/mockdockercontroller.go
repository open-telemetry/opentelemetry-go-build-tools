// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

//go:build integration
// +build integration

// Package dockercontroller provides a mock implementation of the DockerController interface.
package dockercontroller

import (
	dockercontainer "github.com/docker/docker/api/types/container"
	"go.opentelemetry.io/build-tools/grater/internal/container"
)

// MockDockerController is a mock implementation of the DockerController interface.
type MockDockerController struct {
	CreateVolumeMock  func(string) (func(), error)
	UseContainerMock  func(string, []string) (string, func(), error)
	ExecuteCommandMock func(string, []string) (string, dockercontainer.ExecInspect, error)
}

var _ container.Container = (*MockDockerController)(nil)

// NewMockDockerController creates a new instance of MockDockerController.
func NewMockDockerController() *MockDockerController {
	return &MockDockerController{}
}

// CreateVolume creates a mock instance of Volume.
func (m *MockDockerController) CreateVolume(volumeName string) (func(), error) {
	if m.CreateVolumeMock != nil {
		return m.CreateVolumeMock(volumeName)
	}
	return func() {}, nil
}

// UseContainer creates a mock instance of Container.
func (m *MockDockerController) UseContainer(imageName string, volumeNames []string) (string, func(), error) {
	if m.UseContainerMock != nil {
		return m.UseContainerMock(imageName, volumeNames)
	}
	return "", func() {}, nil
}

// ExecuteCommand creates a mock instance of executing a command.
func (m *MockDockerController) ExecuteCommand(containerID string, cmd []string) (string, dockercontainer.ExecInspect, error) {
	if m.ExecuteCommandMock != nil {
		return m.ExecuteCommandMock(containerID, cmd)
	}
	return "", dockercontainer.ExecInspect{}, nil
}