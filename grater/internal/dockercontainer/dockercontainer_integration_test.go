// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

//go:build integration
// +build integration

package dockercontainer

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/moby/moby/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/build-tools/grater/internal/container"
)

func TestCreateVolume(t *testing.T) {
	dc, err := NewDockerContainer()
	require.NoError(t, err)

	volumeNames := []string{"test-volume-grater-1", "test-volume-grater-2"}
	for _, volumeName := range volumeNames {
		resp, createErr := dc.CreateVolume(container.NewCreateVolumeConfig(
			container.WithVolumeName(volumeName),
		))
		require.NoError(t, createErr)
		t.Cleanup(resp.Cleanup)
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
	resp, err := dc.CreateVolume(container.NewCreateVolumeConfig(
		container.WithVolumeName(volumeName),
	))
	require.NoError(t, err)

	resp.Cleanup()

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

	resp, err := dc.UseContainer(container.NewUseContainerConfig(
		container.WithImageName("alpine:latest"),
	))
	require.NoError(t, err)
	defer resp.Cleanup()

	containers, err := dc.cli.ContainerList(context.Background(), client.ContainerListOptions{})
	require.NoError(t, err)

	containerIDs := make([]string, len(containers.Items))
	for i, c := range containers.Items {
		containerIDs[i] = c.ID
	}

	assert.Contains(t, containerIDs, resp.ContainerID)
}

func TestUseContainerBindsVolumes(t *testing.T) {
	dc, err := NewDockerContainer()
	require.NoError(t, err)

	volumeNames := []string{"test-volume-grater-1", "test-volume-grater-2"}
	for _, volumeName := range volumeNames {
		volResp, createErr := dc.CreateVolume(container.NewCreateVolumeConfig(
			container.WithVolumeName(volumeName),
		))
		require.NoError(t, createErr)
		t.Cleanup(volResp.Cleanup)
	}

	resp, err := dc.UseContainer(container.NewUseContainerConfig(
		container.WithImageName("alpine:latest"),
		container.WithBinds(volumeNames),
	))
	require.NoError(t, err)
	defer resp.Cleanup()

	expectedBinds := map[string]string{
		volumeNames[0]: "/data/" + volumeNames[0],
		volumeNames[1]: "/data/" + volumeNames[1],
	}

	inspect, err := dc.cli.ContainerInspect(context.Background(), resp.ContainerID, client.ContainerInspectOptions{})
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
	volResp, err := dc.CreateVolume(container.NewCreateVolumeConfig(
		container.WithVolumeName(volumeName),
	))
	require.NoError(t, err)
	defer volResp.Cleanup()

	resp, err := dc.UseContainer(container.NewUseContainerConfig(
		container.WithImageName("alpine:latest"),
		container.WithBinds([]string{volumeName}),
	))
	require.NoError(t, err)
	defer resp.Cleanup()

	_, err = dc.ExecuteCommand(
		container.NewExecuteCommandConfig(
			container.WithContainerID(resp.ContainerID),
			container.WithCommand([]string{"sh", "-c", "echo 'Hello World' > /data/" + volumeName + "/test_file.txt"}),
		),
	)
	require.NoError(t, err)

	resp2, err := dc.UseContainer(container.NewUseContainerConfig(
		container.WithImageName("alpine:latest"),
		container.WithBinds([]string{volumeName}),
	))
	require.NoError(t, err)
	defer resp2.Cleanup()

	cmdResp2, err := dc.ExecuteCommand(container.NewExecuteCommandConfig(
		container.WithContainerID(resp2.ContainerID),
		container.WithCommand([]string{"cat", "/data/" + volumeName + "/test_file.txt"}),
	))
	require.NoError(t, err)

	assert.Equal(t, "Hello World", cmdResp2.Output)
	assert.Equal(t, 0, cmdResp2.ExitCode)
}

func TestUseContainerBindsLocalPaths(t *testing.T) {
	dc, err := NewDockerContainer()
	require.NoError(t, err)

	testDir := t.TempDir()
	t.Chdir(testDir)

	localPath := "./testdata"
	require.NoError(t, os.MkdirAll(localPath, 0755))
	f, err := os.Create(filepath.Join(localPath, "hello.txt"))
	require.NoError(t, err)
	f.Close()

	resp, err := dc.UseContainer(container.NewUseContainerConfig(
		container.WithImageName("alpine:latest"),
		container.WithHostPaths([]string{localPath}),
	))
	require.NoError(t, err)
	defer resp.Cleanup()

	cmdResp, err := dc.ExecuteCommand(container.NewExecuteCommandConfig(
		container.WithContainerID(resp.ContainerID),
		container.WithCommand([]string{"ls", "/data/" + filepath.Base(localPath)}),
	))
	require.NoError(t, err)

	assert.Equal(t, 0, cmdResp.ExitCode)
	assert.Contains(t, cmdResp.Output, "hello.txt")
}

func TestUseContainerCleanupRemovesContainer(t *testing.T) {
	dc, err := NewDockerContainer()
	require.NoError(t, err)

	resp, err := dc.UseContainer(container.NewUseContainerConfig(
		container.WithImageName("alpine:latest"),
	))
	require.NoError(t, err)
	resp.Cleanup()

	containers, err := dc.cli.ContainerList(context.Background(), client.ContainerListOptions{All: true})
	require.NoError(t, err)

	containerIDs := make([]string, len(containers.Items))
	for i, c := range containers.Items {
		containerIDs[i] = c.ID
	}
	assert.NotContains(t, containerIDs, resp.ContainerID)
}

func TestExecuteCommand(t *testing.T) {
	dc, err := NewDockerContainer()
	require.NoError(t, err)

	resp, err := dc.UseContainer(container.NewUseContainerConfig(
		container.WithImageName("ubuntu:latest"),
	))
	require.NoError(t, err)
	defer resp.Cleanup()

	cmdResp, err := dc.ExecuteCommand(container.NewExecuteCommandConfig(
		container.WithContainerID(resp.ContainerID),
		container.WithCommand([]string{"echo", "hello world"}),
	))
	require.NoError(t, err)

	assert.Equal(t, "hello world", cmdResp.Output)
	assert.Equal(t, 0, cmdResp.ExitCode)
}

func TestExecuteCommandExitCode1(t *testing.T) {
	dc, err := NewDockerContainer()
	require.NoError(t, err)

	resp, err := dc.UseContainer(container.NewUseContainerConfig(
		container.WithImageName("ubuntu:latest"),
	))
	require.NoError(t, err)
	defer resp.Cleanup()

	cmdResp, err := dc.ExecuteCommand(container.NewExecuteCommandConfig(
		container.WithContainerID(resp.ContainerID),
		container.WithCommand([]string{"false"}),
	))
	require.NoError(t, err)

	assert.Equal(t, "", cmdResp.Output)
	assert.Equal(t, 1, cmdResp.ExitCode)
}

func TestExecuteCommandInvalidContainerFails(t *testing.T) {
	dc, err := NewDockerContainer()
	require.NoError(t, err)

	cmdResp, err := dc.ExecuteCommand(container.NewExecuteCommandConfig(
		container.WithContainerID("invalid-container"),
		container.WithCommand([]string{"echo", "test"}),
	))
	require.Error(t, err)
	assert.Empty(t, cmdResp.Output)
	assert.Equal(t, 0, cmdResp.ExitCode)
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
