// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

//go:build integration
// +build integration

// Package container provides an interface for managing Docker containers and volumes.
package container

import (
    "github.com/docker/docker/api/types/container"
)

// Container is an interface for managing Docker containers and volumes.
type Container interface {
	CreateVolume(volumeName string) (func(), error)
    UseContainer(imageName string, volumeNames []string) (string, func(), error)
    ExecuteCommand(containerID string, cmd []string) (string, container.ExecInspect, error)
}