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

func createVolumeTest(t *testing.T) {
	err, dc := NewDockerController()
	require.NoError(t, err)

	volNames := []string{"test-volume", "test-volume-2"}
	for _, volName := range volNames {
		err := dc.CreateVolume(volName)
		require.NoError(t, err)
	}

	volumes, err := dc.cli.VolumeList(context.Background(), volume.ListOptions{})
	require.NoError(t, err)

	volNameList := make([]string, len(volumes.Volumes))
	for i, v := range volumes.Volumes {
		volNameList[i] = v.Name
	}

	assert.ElementsMatch(t, volNames, volNameList)
}

func createContainerTest(t *testing.T) {
	err, dc := NewDockerController()
	require.NoError(t, err)

	imageName := "alpine:latest"
	resp, cleanup, err := dc.UseContainer(imageName, []string{})
	require.NoError(t, err)
	defer cleanup()

	containers, err := dc.cli.ContainerList(context.Background(), container.ListOptions{})
	require.NoError(t, err)

	found := false
	for _, c := range containers {
		if c.ID == resp {
			found = true
			break
		}
	}

	assert.True(t, found, "container should exist in container list")
}

func executeCommandTest(t *testing.T) {
	err, dc := NewDockerController()
	require.NoError(t, err)

	imageName := "alpine:latest"
	resp, cleanup, err := dc.UseContainer(imageName, []string{})
	require.NoError(t, err)
	defer cleanup()

	cmd := []string{"echo", "hello world"}

	output, err := dc.ExecuteCommand(resp, cmd)
	require.NoError(t, err)

	assert.Equal(t, "hello world", output)
}