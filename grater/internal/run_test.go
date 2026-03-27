// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package internal

import (
	"testing"

	"github.com/docker/docker/client"

	"github.com/docker/docker/api/types/container"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateContainer(t *testing.T) {
	cli, err := createDockerTestEnvironment()
	require.NoError(t, err)
	defer cli.Close()

	volName := "test-volume"
	ctx, resp, err := createContainer(volName, "golang:1.24-alpine", cli)
	require.NoError(t, err)
	assert.NotEmpty(t, resp.ID)

	inspected, err := cli.ContainerInspect(ctx, resp.ID)
	assert.NoError(t, err)
	assert.Equal(t, resp.ID, inspected.ID)
	assert.Equal(t, "golang:1.24-alpine", inspected.Config.Image)
}

func TestCreateContainerFails(t *testing.T) {
	cli, err := createDockerTestEnvironment()
	require.NoError(t, err)
	defer cli.Close()

	volName := "test-volume"
	_, resp, err := createContainer(volName, "invalid-image", cli)
	require.Error(t, err)
	assert.Nil(t, resp)
}

func TestRunCommand(t *testing.T) {
	cli, err := createDockerTestEnvironment()
	require.NoError(t, err)
	defer cli.Close()

	volName := "test-volume"
	ctx, resp, err := createContainer(volName, "golang:1.24-alpine", cli)
	require.NoError(t, err)
	assert.NotEmpty(t, resp.ID)

	cli.ContainerStart(ctx, resp.ID, container.StartOptions{})
	require.NoError(t, err)
	defer cli.ContainerStop(ctx, resp.ID, container.StopOptions{})

	err = runCommand(ctx, cli, resp, []string{"go", "version"})
	assert.NoError(t, err)
	assert.NotEmpty(t, resp.ID)
}

func testRunCommandFails(t *testing.T) {
	cli, err := createDockerTestEnvironment()
	require.NoError(t, err)
	defer cli.Close()

	volName := "test-volume"
	ctx, resp, err := createContainer(volName, "alpine/git", cli)
	require.NoError(t, err)
	assert.NotEmpty(t, resp.ID)

	cli.ContainerStart(ctx, resp.ID, container.StartOptions{})
	require.NoError(t, err)
	defer cli.ContainerStop(ctx, resp.ID, container.StopOptions{})

	err = runCommand(ctx, cli, resp, []string{"go", "version"})
	assert.Error(t, err)
	assert.NotEmpty(t, resp.ID)
	assert.Contains(t, err.Error(), "not found")
}

func createDockerTestEnvironment() (*client.Client, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, err
	}

	return cli, nil
}
