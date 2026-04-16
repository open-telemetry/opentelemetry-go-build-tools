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

	volNames := []string{"test-volume-grater", "test-volume-grater-2"}
	for _, volName := range volNames {
		cleanup, createErr := dc.CreateVolume(volName)
		require.NoError(t, createErr)
		t.Cleanup(cleanup)
	}

	volumes, err := dc.cli.VolumeList(context.Background(), volume.ListOptions{})
	require.NoError(t, err)

	volNameList := make([]string, len(volumes.Volumes))
	for i, v := range volumes.Volumes {
		volNameList[i] = v.Name
	}

	assert.Subset(t, volNameList, volNames)
}

func TestCreateVolumeCleanupRemovesVolume(t *testing.T) {
	dc, err := NewDockerController()
	require.NoError(t, err)

	volName := "test-volume-grater"
	cleanup, err := dc.CreateVolume(volName)
	require.NoError(t, err)

	cleanup()

	volumes, err := dc.cli.VolumeList(context.Background(), volume.ListOptions{})
	require.NoError(t, err)

	volNameList := make([]string, len(volumes.Volumes))
	for i, v := range volumes.Volumes {
		volNameList[i] = v.Name
	}

	assert.NotContains(t, volNameList, volName)
}

func TestUseContainer(t *testing.T) {
	dc, err := NewDockerController()
	require.NoError(t, err)

	imageName := "alpine:latest"
	resp, cleanup, err := dc.UseContainer(imageName, []string{})
	require.NoError(t, err)
	defer cleanup()

	containers, err := dc.cli.ContainerList(context.Background(), container.ListOptions{})
	require.NoError(t, err)

	containerIDs := make([]string, len(containers))
	for i, c := range containers {
		containerIDs[i] = c.ID
	}

	assert.Contains(t, containerIDs, resp)
}

func TestUseContainerBindsVolumes(t *testing.T) {
	dc, err := NewDockerController()
	require.NoError(t, err)

	volume1 := "test-volume-grater-1"
	volume2 := "test-volume-grater-2"

	cleanupVol1, err := dc.CreateVolume(volume1)
	require.NoError(t, err)
	defer cleanupVol1()

	cleanupVol2, err := dc.CreateVolume(volume2)
	require.NoError(t, err)
	defer cleanupVol2()

	imageName := "alpine:latest"
	containerID, cleanup, err := dc.UseContainer(imageName, []string{volume1, volume2})
	require.NoError(t, err)
	defer cleanup()

	expected := map[string]string{
		volume1: "/data/" + volume1,
		volume2: "/data/" + volume2,
	}

	inspect, err := dc.cli.ContainerInspect(context.Background(), containerID)
	require.NoError(t, err)

	found := make(map[string]bool)
	for _, m := range inspect.Mounts {
		if dest, ok := expected[m.Name]; ok && m.Destination == dest {
			found[m.Name] = true
		}
	}

	assert.Len(t, found, 2)
	assert.True(t, found[volume1])
	assert.True(t, found[volume2])
}

func TestUseContainerCleanupRemovesContainer(t *testing.T) {
    dc, err := NewDockerController()
    require.NoError(t, err)

    id, cleanup, err := dc.UseContainer("alpine:latest", []string{})
    require.NoError(t, err)
    cleanup()

    containers, err := dc.cli.ContainerList(context.Background(), container.ListOptions{All: true})
    require.NoError(t, err)

    ids := make([]string, len(containers))
    for i, c := range containers {
        ids[i] = c.ID
    }
    assert.NotContains(t, ids, id)
}

func TestExecuteCommand(t *testing.T) {
	dc, err := NewDockerController()
	require.NoError(t, err)

	imageName := "ubuntu:latest"
	resp, cleanup, err := dc.UseContainer(imageName, []string{})
	require.NoError(t, err)
	defer cleanup()

	output, err := dc.ExecuteCommand(resp, []string{"echo", "hello world"})
	require.NoError(t, err)

	assert.Equal(t, "hello world", output)
}

func TestPullImage(t *testing.T) {
	dc, err := NewDockerController()
	require.NoError(t, err)

	imageName := "ubuntu:latest"
	err = dc.pullImage(imageName)
	require.NoError(t, err)

	_, _, err = dc.cli.ImageInspectWithRaw(dc.ctx, imageName)
	require.NoError(t, err)
}

func TestPullImageFails(t *testing.T) {
	dc, err := NewDockerController()
	require.NoError(t, err)

	imageName := "invalid-image-name"
	err = dc.pullImage(imageName)
	require.Error(t, err)
}