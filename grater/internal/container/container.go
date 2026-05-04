// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

// Package container provides an interface for managing Docker containers and volumes.
package container

import (
	"context"
)

// ExecuteCommandResponse represents the response from executing a command in a container.
type ExecuteCommandResponse struct {
	Output   string
	ExitCode int
}

// UseContainerResponse represents the response from using a container.
type UseContainerResponse struct {
	ContainerID string
	Cleanup     func()
}

// CreateVolumeResponse represents the response from creating a volume.
type CreateVolumeResponse struct {
	Cleanup func()
}

// NewExecuteCommandResponse creates a new ExecuteCommandResponse.
func NewExecuteCommandResponse(output string, exitCode int) ExecuteCommandResponse {
	return ExecuteCommandResponse{
		Output:   output,
		ExitCode: exitCode,
	}
}

// NewUseContainerResponse creates a new UseContainerResponse.
func NewUseContainerResponse(containerID string, cleanup func()) UseContainerResponse {
	return UseContainerResponse{
		ContainerID: containerID,
		Cleanup:     cleanup,
	}
}

// NewCreateVolumeResponse creates a new CreateVolumeResponse.
func NewCreateVolumeResponse(cleanup func()) CreateVolumeResponse {
	return CreateVolumeResponse{
		Cleanup: cleanup,
	}
}

// Container defines the interface for managing Docker containers and volumes.
type Container interface {
	CreateVolume(ctx context.Context, cfg CreateVolumeConfig) (CreateVolumeResponse, error)
	UseContainer(ctx context.Context, cfg UseContainerConfig) (UseContainerResponse, error)
	ExecuteCommand(ctx context.Context, cfg ExecuteCommandConfig) (ExecuteCommandResponse, error)
}
