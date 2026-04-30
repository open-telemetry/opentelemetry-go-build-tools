// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

//go:build integration
// +build integration

package dockercontainer

import (
	"context"
	"testing"
	"path/filepath"

	"github.com/moby/moby/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateVolume(t *testing.T) {
	dc, err := NewDockerContainer()
	require.NoError(t, err)

	volumeNames := []string{"test-volume-grater-1", "test-volume-grater-2"}
	for _, volumeName := range volumeNames {
		cleanup, createErr := dc.CreateVolume(volumeName)
		require.NoError(t, createErr)
		t.Cleanup(cleanup)
	}

	volumes, err := dc.cli.VolumeList(context.Background(), client.VolumeListOptions{})
	require.NoError(t, err)

	volumeNameList := make([]string, len(volumes.Items))
	for i, volume := range volumes.Items {
		volumeNameList[i] = volume.Name
	}

	assert.Subset(t, volumeNameList, volumeNames)
}

func TestCreateVolumeCleanupRemovesVolume(t *testing.T) {
	dc, err := NewDockerContainer()
	require.NoError(t, err)

	volumeName := "test-volume-grater"
	cleanup, err := dc.CreateVolume(volumeName)
	require.NoError(t, err)

	cleanup()

	volumes, err := dc.cli.VolumeList(context.Background(), client.VolumeListOptions{})
	require.NoError(t, err)

	volumeNameList := make([]string, len(volumes.Items))
	for i, volume := range volumes.Items {
		volumeNameList[i] = volume.Name
	}

	assert.NotContains(t, volumeNameList, volumeName)
}

func TestUseContainer(t *testing.T) {
	dc, err := NewDockerContainer()
	require.NoError(t, err)

	imageName := "alpine:latest"
	containerID, cleanup, err := dc.UseContainer(imageName, []string{}, []string{})
	require.NoError(t, err)
	defer cleanup()

	containers, err := dc.cli.ContainerList(context.Background(), client.ContainerListOptions{})
	require.NoError(t, err)

	containerIDs := make([]string, len(containers.Items))
	for i, c := range containers.Items {
		containerIDs[i] = c.ID
	}

	assert.Contains(t, containerIDs, containerID)
}

func TestUseContainerBindsVolumes(t *testing.T) {
	dc, err := NewDockerContainer()
	require.NoError(t, err)

	volumeNames := []string{"test-volume-grater-1", "test-volume-grater-2"}
	for _, volumeName := range volumeNames {
		cleanup, createErr := dc.CreateVolume(volumeName)
		require.NoError(t, createErr)
		t.Cleanup(cleanup)
	}

	imageName := "alpine:latest"
	containerID, cleanup, err := dc.UseContainer(imageName, volumeNames, []string{})
	require.NoError(t, err)
	defer cleanup()

	expectedBinds := map[string]string{
		volumeNames[0]: "/data/" + volumeNames[0],
		volumeNames[1]: "/data/" + volumeNames[1],
	}

	inspect, err := dc.cli.ContainerInspect(context.Background(), containerID, client.ContainerInspectOptions{})
	require.NoError(t, err)

	binds := make(map[string]bool)
	for _, mount := range inspect.Container.Mounts {
		if path, ok := expectedBinds[mount.Name]; ok && mount.Destination == path {
			binds[mount.Name] = true
		}
	}

	assert.Len(t, binds, 2)
	assert.True(t, binds[volumeNames[0]])
	assert.True(t, binds[volumeNames[1]])
}

func TestUseContainerReadsAndWritesToVolume(t *testing.T) {
	dc, err := NewDockerContainer()
	require.NoError(t, err)

	volumeName := "test-volume-grater"
	cleanupVolume, err := dc.CreateVolume(volumeName)
	require.NoError(t, err)
	defer cleanupVolume()

	imageName := "alpine:latest"
	containerID, cleanup, err := dc.UseContainer(imageName, []string{volumeName}, []string{})
	require.NoError(t, err)
	defer cleanup()

	out, exitCode, err := dc.ExecuteCommand(
		containerID,
		[]string{"sh", "-c", "echo 'Hello World' > /data/" + volumeName + "/test_file.txt"},
	)
	require.NoError(t, err)

	containerID2, cleanup2, err := dc.UseContainer(imageName, []string{volumeName}, []string{})
	require.NoError(t, err)
	defer cleanup2()

	out, exitCode, err = dc.ExecuteCommand(
		containerID2,
		[]string{"cat", "/data/" + volumeName + "/test_file.txt"},
	)
	require.NoError(t, err)

	assert.Equal(t, "Hello World", out)
	assert.Equal(t, 0, exitCode)
}

func TestUseContainerBindsLocalPaths(t *testing.T) {
	dc, err := NewDockerContainer()
	require.NoError(t, err)

	testDir := t.TempDir()
	t.Chdir(testDir)

	localPath := "./testdata"
	containerID, cleanup, err := dc.UseContainer("alpine:latest", []string{}, []string{localPath})
	require.NoError(t, err)
	defer cleanup()

	out, exitCode, err := dc.ExecuteCommand(containerID, []string{"ls", "/data/" + filepath.Base(localPath)})
	require.NoError(t, err)

	assert.Equal(t, 0, exitCode)
	assert.NotEmpty(t, out)
}

func TestUseContainerCleanupRemovesContainer(t *testing.T) {
	dc, err := NewDockerContainer()
	require.NoError(t, err)

	containerID, cleanup, err := dc.UseContainer("alpine:latest", []string{}, []string{})
	require.NoError(t, err)
	cleanup()

	containers, err := dc.cli.ContainerList(context.Background(), client.ContainerListOptions{All: true})
	require.NoError(t, err)

	containerIDs := make([]string, len(containers.Items))
	for i, c := range containers.Items {
		containerIDs[i] = c.ID
	}
	assert.NotContains(t, containerIDs, containerID)
}

func TestExecuteCommand(t *testing.T) {
	dc, err := NewDockerContainer()
	require.NoError(t, err)

	containerID, cleanup, err := dc.UseContainer("ubuntu:latest", []string{}, []string{})
	require.NoError(t, err)
	defer cleanup()

	out, exitCode, err := dc.ExecuteCommand(containerID, []string{"echo", "hello world"})
	require.NoError(t, err)

	assert.Equal(t, "hello world", out)
	assert.Equal(t, 0, exitCode)
}

func TestExecuteCommandExitCode1(t *testing.T) {
	dc, err := NewDockerContainer()
	require.NoError(t, err)

	containerID, cleanup, err := dc.UseContainer("ubuntu:latest", []string{}, []string{})
	require.NoError(t, err)
	defer cleanup()

	out, exitCode, err := dc.ExecuteCommand(containerID, []string{"false"})
	require.NoError(t, err)

	assert.Equal(t, "", out)
	assert.Equal(t, 1, exitCode)
}

func TestExecuteCommandInvalidContainerFails(t *testing.T) {
	dc, err := NewDockerContainer()
	require.NoError(t, err)

	out, exitCode, err := dc.ExecuteCommand("invalid-container", []string{"echo", "test"})
	require.Error(t, err)
	assert.Empty(t, out)
	assert.Equal(t, 0, exitCode)
}

func TestPullImage(t *testing.T) {
	dc, err := NewDockerContainer()
	require.NoError(t, err)

	err = dc.pullImage("ubuntu:latest")
	require.NoError(t, err)

	_, err = dc.cli.ImageInspect(dc.ctx, "ubuntu:latest")
	require.NoError(t, err)
}

func TestPullImageFails(t *testing.T) {
	dc, err := NewDockerContainer()
	require.NoError(t, err)

	err = dc.pullImage("invalid-image-name")
	require.Error(t, err)
}
