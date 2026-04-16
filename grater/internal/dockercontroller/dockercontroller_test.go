// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

//go:build integration
// +build integration

package dockercontroller

import (
	"context"
	"testing"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/volume"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateVolume(t *testing.T) {
	dc, err := NewDockerController()
	require.NoError(t, err)

	volumeNames := []string{"test-volume-grater-1", "test-volume-grater-2"}
	for _, volumeName := range volumeNames {
		cleanup, createErr := dc.CreateVolume(volumeName)
		require.NoError(t, createErr)
		t.Cleanup(cleanup)
	}

	volumes, err := dc.cli.VolumeList(context.Background(), volume.ListOptions{})
	require.NoError(t, err)

	volumeNameList := make([]string, len(volumes.Volumes))
	for i, volume := range volumes.Volumes {
		volumeNameList[i] = volume.Name
	}

	assert.Subset(t, volumeNameList, volumeNames)
}

func TestCreateVolumeCleanupRemovesVolume(t *testing.T) {
	dc, err := NewDockerController()
	require.NoError(t, err)

	volumeName := "test-volume-grater"
	cleanup, err := dc.CreateVolume(volumeName)
	require.NoError(t, err)

	cleanup() // Call cleanup immediately to remove volume. 

	volumes, err := dc.cli.VolumeList(context.Background(), volume.ListOptions{})
	require.NoError(t, err)

	volumeNameList := make([]string, len(volumes.Volumes))
	for i, volume := range volumes.Volumes {
		volumeNameList[i] = volume.Name
	}

	assert.NotContains(t, volumeNameList, volumeName)
}

func TestUseContainer(t *testing.T) {
	dc, err := NewDockerController()
	require.NoError(t, err)

	imageName := "alpine:latest"
	containerID, cleanup, err := dc.UseContainer(imageName, []string{})
	require.NoError(t, err)
	defer cleanup()

	containers, err := dc.cli.ContainerList(context.Background(), container.ListOptions{})
	require.NoError(t, err)

	containerIDs := make([]string, len(containers))
	for i, c := range containers {
		containerIDs[i] = c.ID
	}

	assert.Contains(t, containerIDs, containerID)
}

func TestUseContainerBindsVolumes(t *testing.T) {
	dc, err := NewDockerController()
	require.NoError(t, err)

	volumeNames := []string{"test-volume-grater-1", "test-volume-grater-2"}
	for _, volumeName := range volumeNames {
		cleanup, createErr := dc.CreateVolume(volumeName)
		require.NoError(t, createErr)
		t.Cleanup(cleanup)
	}

	imageName := "alpine:latest"
	containerID, cleanup, err := dc.UseContainer(imageName, volumeNames)
	require.NoError(t, err)
	defer cleanup()

	expectedBinds := map[string]string{
		volumeNames[0]: "/data/" + volumeNames[0],
		volumeNames[1]: "/data/" + volumeNames[1],
	}

	inspect, err := dc.cli.ContainerInspect(context.Background(), containerID)
	require.NoError(t, err)

	binds := make(map[string]bool)
	for _, mount := range inspect.Mounts {
		if path, ok := expectedBinds[mount.Name]; ok && mount.Destination == path {
			binds[mount.Name] = true
		}
	}

	assert.Len(t, binds, 2)
	assert.True(t, binds[volumeNames[0]])
	assert.True(t, binds[volumeNames[1]])
}

func TestUseContainerReadsAndWritesToVolume(t *testing.T) {
	dc, err := NewDockerController()
	require.NoError(t, err)

	volumeName := "test-volume-grater"
	cleanupVolume, err := dc.CreateVolume(volumeName)
	require.NoError(t, err)
	defer cleanupVolume()

	imageName := "alpine:latest"
	containerID, cleanup, err := dc.UseContainer(imageName, []string{volumeName})
	require.NoError(t, err)
	defer cleanup()

	out, inspect, err := dc.ExecuteCommand(
		containerID,
		[]string{"sh", "-c", "echo 'Hello World' > /data/" + volumeName + "/test_file.txt"},
	)
	require.NoError(t, err)

	containerID2, cleanup2, err := dc.UseContainer(imageName, []string{volumeName})
	require.NoError(t, err)
	defer cleanup2()

	out, inspect, err = dc.ExecuteCommand(
		containerID2,
		[]string{"cat", "/data/" + volumeName + "/test_file.txt"},
	)
	require.NoError(t, err)

	assert.Equal(t, "Hello World", out)
	assert.Equal(t, 0, inspect.ExitCode)
}

func TestUseContainerCleanupRemovesContainer(t *testing.T) {
    dc, err := NewDockerController()
    require.NoError(t, err)

    containerID, cleanup, err := dc.UseContainer("alpine:latest", []string{})
    require.NoError(t, err)
    cleanup() // Call cleanup immediately to remove container

    containers, err := dc.cli.ContainerList(context.Background(), container.ListOptions{All: true})
    require.NoError(t, err)

    containerIDs := make([]string, len(containers))
    for i, container := range containers {
        containerIDs[i] = container.ID
    }
    assert.NotContains(t, containerIDs, containerID)
}

func TestExecuteCommand(t *testing.T) {
	dc, err := NewDockerController()
	require.NoError(t, err)

	containerID, cleanup, err := dc.UseContainer("ubuntu:latest", []string{})
	require.NoError(t, err)
	defer cleanup()

	out, inspect, err := dc.ExecuteCommand(containerID, []string{"echo", "hello world"})
	require.NoError(t, err)

	assert.Equal(t, "hello world", out)
	assert.Equal(t, 0, inspect.ExitCode)
}

func TestExecuteCommandExitCode1(t *testing.T) {
	dc, err := NewDockerController()
	require.NoError(t, err)

	containerID, cleanup, err := dc.UseContainer("ubuntu:latest", []string{})
	require.NoError(t, err)
	defer cleanup()

	out, inspect, err := dc.ExecuteCommand(containerID, []string{"false"})
	require.NoError(t, err)

	assert.Equal(t, "", out)
	assert.Equal(t, 1, inspect.ExitCode)
}

func TestExecuteCommandInvalidContainerFails(t *testing.T) {
	dc, err := NewDockerController()
	require.NoError(t, err)

	out, inspect, err := dc.ExecuteCommand("invalid-container", []string{"echo", "test"})
	require.Error(t, err)
	assert.Empty(t, out)
	assert.Equal(t, 0, inspect.ExitCode)
}

func TestPullImage(t *testing.T) {
	dc, err := NewDockerController()
	require.NoError(t, err)

	err = dc.pullImage("ubuntu:latest")
	require.NoError(t, err)

	_, _, err = dc.cli.ImageInspectWithRaw(dc.ctx, "ubuntu:latest")
	require.NoError(t, err)
}

func TestPullImageFails(t *testing.T) {
	dc, err := NewDockerController()
	require.NoError(t, err)

	err = dc.pullImage("invalid-image-name")
	require.Error(t, err)
}