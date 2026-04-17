// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

// Package controller provides an interface for managing Docker containers and volumes.
package controller

import (
    "github.com/docker/docker/api/types/container"
)

// Controller is an interface for managing Docker containers and volumes.
type Controller interface {
	CreateVolume(volumeName string) (func(), error)
    UseContainer(imageName string, volumeNames []string) (string, func(), error)
    ExecuteCommand(containerID string, cmd []string) (string, container.ExecInspect, error)
}