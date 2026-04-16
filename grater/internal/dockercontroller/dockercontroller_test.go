// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

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
	err, dc := NewDockerController()
	require.NoError(t, err)

	volNames := []string{"test-volume-grater", "test-volume-grater-2"}
	for _, volName := range volNames {
		cleanup, err := dc.CreateVolume(volName)
		require.NoError(t, err)
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

func TestRemoveVolumeCleanupRemovesVolume(t *testing.T) {
	err, dc := NewDockerController()
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
	err, dc := NewDockerController()
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

func TestUseContainerCleanupRemovesContainer(t *testing.T) {
    err, dc := NewDockerController()
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
	err, dc := NewDockerController()
	require.NoError(t, err)

	imageName := "ubuntu:latest"
	resp, cleanup, err := dc.UseContainer(imageName, []string{})
	require.NoError(t, err)
	defer cleanup()

	output, err := dc.ExecuteCommand(resp, []string{"echo", "hello world"})
	require.NoError(t, err)

	assert.Equal(t, "hello world", output)
}
