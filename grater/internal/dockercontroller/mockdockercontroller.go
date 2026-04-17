// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package dockercontroller

import (
    "github.com/stretchr/testify/mock"
	"github.com/docker/docker/api/types/container"
	"go.opentelemetry.io/build-tools/grater/internal/controller"
)

type MockDockerController struct {
    mock.Mock
}

var _ controller.Controller = (*MockDockerController)(nil)

func NewMockDockerController() *MockDockerController {
    return &MockDockerController{}
}

func (m *MockDockerController) CreateVolume(volumeName string) (func(), error) {
    args := m.Called(volumeName)
    return args.Get(0).(func()), args.Error(1)
}

func (m *MockDockerController) UseContainer(imageName string, volumeNames []string) (string, func(), error) {
    args := m.Called(imageName, volumeNames)
    return args.String(0), args.Get(1).(func()), args.Error(2)
}

func (m *MockDockerController) ExecuteCommand(containerID string, cmd []string) (string, container.ExecInspect, error) {
    args := m.Called(containerID, cmd)
    return args.String(0), args.Get(1).(container.ExecInspect), args.Error(2)
}
