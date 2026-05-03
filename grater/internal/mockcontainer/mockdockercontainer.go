// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

// Package mockcontainer provides a mock implementation of the Container interface.
package mockcontainer

import (
	"context"

	"go.opentelemetry.io/build-tools/grater/internal/container"
)

// MockDockerContainer is a mock implementation of the Container interface.
type MockDockerContainer struct {
	CreateVolumeMock   func(context.Context, container.CreateVolumeConfig) (container.CreateVolumeResponse, error)
	UseContainerMock   func(context.Context, container.UseContainerConfig) (container.UseContainerResponse, error)
	ExecuteCommandMock func(context.Context, container.ExecuteCommandConfig) (container.ExecuteCommandResponse, error)
}

var _ container.Container = (*MockDockerContainer)(nil)

// NewMockDockerContainer creates a new instance of MockDockerContainer.
func NewMockDockerContainer() *MockDockerContainer {
	return &MockDockerContainer{}
}

// CreateVolume creates a mock instance of Volume.
func (m *MockDockerContainer) CreateVolume(ctx context.Context, cfg container.CreateVolumeConfig) (container.CreateVolumeResponse, error) {
	if m.CreateVolumeMock != nil {
		return m.CreateVolumeMock(ctx, cfg)
	}
	return container.CreateVolumeResponse{}, nil
}

// UseContainer creates a mock instance of Container.
func (m *MockDockerContainer) UseContainer(ctx context.Context, cfg container.UseContainerConfig) (container.UseContainerResponse, error) {
	if m.UseContainerMock != nil {
		return m.UseContainerMock(ctx, cfg)
	}
	return container.UseContainerResponse{}, nil
}

// ExecuteCommand creates a mock instance of executing a command.
func (m *MockDockerContainer) ExecuteCommand(ctx context.Context, cfg container.ExecuteCommandConfig) (container.ExecuteCommandResponse, error) {
	if m.ExecuteCommandMock != nil {
		return m.ExecuteCommandMock(ctx, cfg)
	}
	return container.ExecuteCommandResponse{}, nil
}
