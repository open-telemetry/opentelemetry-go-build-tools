// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package container

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewExecuteCommandConfigAllOptions(t *testing.T) {
	cfg := NewExecuteCommandConfig(
		WithContainerID("test-container"),
		WithCommand("echo hello"),
	)
	assert.Equal(t, "test-container", cfg.containerID)
	assert.Equal(t, "echo hello", cfg.cmd)
}

func TestNewExecuteCommandConfigSomeOptions(t *testing.T) {
	cfg := NewExecuteCommandConfig(
		WithContainerID("test-container"),
	)
	assert.Equal(t, "test-container", cfg.containerID)
	assert.Empty(t, cfg.cmd)

	cfg = NewExecuteCommandConfig(
		WithCommand("echo hello"),
	)
	assert.Empty(t, cfg.containerID)
	assert.Equal(t, "echo hello", cfg.cmd)
}

func TestNewExecuteCommandConfigNoOptions(t *testing.T) {
	cfg := NewExecuteCommandConfig()
	assert.Empty(t, cfg.containerID)
	assert.Empty(t, cfg.cmd)
}

func TestNewCreateVolumeConfigAllOptions(t *testing.T) {
	cfg := NewCreateVolumeConfig(
		WithVolumeName("test-volume"),
	)
	assert.Equal(t, "test-volume", cfg.volumeName)
}

func TestNewCreateVolumeConfigNoOptions(t *testing.T) {
	cfg := NewCreateVolumeConfig()
	assert.Empty(t, cfg.volumeName)
}

func TestNewUseContainerConfigAllOptions(t *testing.T) {
	cfg := NewUseContainerConfig(
		WithImageName("test-image"),
		WithBinds([]string{"test-bind"}),
		WithHostPaths([]string{"test-host-path"}),
	)
	assert.Equal(t, "test-image", cfg.imageName)
	assert.Equal(t, []string{"test-bind"}, cfg.binds)
	assert.Equal(t, []string{"test-host-path"}, cfg.hostPaths)
}

func TestNewUseContainerConfigSomeOptions(t *testing.T) {
	cfg := NewUseContainerConfig(
		WithImageName("test-image"),
		WithBinds([]string{"test-bind"}),
	)
	assert.Equal(t, "test-image", cfg.imageName)
	assert.Equal(t, []string{"test-bind"}, cfg.binds)
	assert.Empty(t, cfg.hostPaths)
}

func TestNewUseContainerConfigNoOptions(t *testing.T) {
	cfg := NewUseContainerConfig()
	assert.Empty(t, cfg.imageName)
	assert.Empty(t, cfg.binds)
	assert.Empty(t, cfg.hostPaths)
}