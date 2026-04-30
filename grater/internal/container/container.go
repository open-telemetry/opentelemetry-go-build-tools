// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

// Package container provides an interface for managing Docker containers and volumes.
package container

// Container defines the interface for managing Docker containers and volumes.
type Container interface {
	CreateVolume(volumeName string) (func(), error)
	UseContainer(imageName string, volumeNames, localPaths []string) (string, func(), error)
	ExecuteCommand(containerID string, cmd []string) (string, int, error)
}
