// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

// Package container provides an interface for managing Docker containers and volumes.
package container

type ExecuteCommandResponse struct {
	Output    string
	ExitCode  int
	Error     error
}

type UseContainerResponse struct {
	ContainerID string
	cleanup     func()
	Error       error
}

type CreateVolumeResponse struct {
	cleanup func()
	Error   error
}

// Container defines the interface for managing Docker containers and volumes.
type Container interface {
	CreateVolume(volumeName string) (func(), error)
	UseContainer(imageName string, volumeNames, localPaths []string) (string, func(), error)
	ExecuteCommand(containerID string, cmd []string) (string, int, error)
}
