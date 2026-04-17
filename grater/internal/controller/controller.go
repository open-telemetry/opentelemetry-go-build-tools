// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package controller

import (
    "github.com/docker/docker/api/types/container"
)

type Controller interface {
	CreateVolume(volumeName string) (func(), error)
    UseContainer(imageName string, volumeNames []string) (string, func(), error)
    ExecuteCommand(containerID string, cmd []string) (string, container.ExecInspect, error)
}