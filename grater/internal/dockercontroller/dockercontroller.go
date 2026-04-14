// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

// Package dockercontroller provides a controller for managing Docker.
package dockercontroller

import (
	"bytes"
	"context"
	"strings"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/volume"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
)

type dockercontroller struct {
	cli     *client.Client
	ctx     context.Context
	volumes []string
}

// NewDockerController creates a new Docker controller.
func NewDockerController() (error, *dockercontroller) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return err, nil
	}

	return nil, &dockercontroller{
		cli:     cli,
		ctx:     context.Background(),
		volumes: []string{},
	}
}

// CreateVolume creates a volume.
func (dc *dockercontroller) CreateVolume(volName string) error {
	_, err := dc.cli.VolumeCreate(dc.ctx, volume.CreateOptions{
		Name: volName,
	})
	return err
}

// UseContainer creates a container with specified volumes and returns a cleanup function.
func (dc *dockercontroller) UseContainer(imageName string, volumes []string) (string, func(), error) {
	reader, err := dc.cli.ImagePull(dc.ctx, imageName, image.PullOptions{})
	if err != nil {
		return "", nil, err
	}
	defer reader.Close()

	volumeMap := make(map[string]struct{})
	for _, v := range volumes {
		volumeMap[v] = struct{}{}
	}

	resp, err := dc.cli.ContainerCreate(
		dc.ctx,
		&container.Config{
			Image:   imageName,
			Volumes: volumeMap,
		},
		&container.HostConfig{
			Binds: volumes,
		},
		nil,
		nil,
		"",
	)
	if err != nil {
		return "", nil, err
	}

	if err := dc.cli.ContainerStart(dc.ctx, resp.ID, container.StartOptions{}); err != nil {
		return "", nil, err
	}

	cleanup := func() {
		timeout := 10
		dc.cli.ContainerStop(dc.ctx, resp.ID, container.StopOptions{Timeout: &timeout})
		dc.cli.ContainerRemove(dc.ctx, resp.ID, container.RemoveOptions{Force: true})
	}

	return resp.ID, cleanup, nil
}

// ExecuteCommand executes a command in a container and returns the output.
func (dc *dockercontroller) ExecuteCommand(containerID string, cmd []string) (string, error) {
	execConfig := container.ExecOptions{
		Cmd:          cmd,
		AttachStdout: true,
		AttachStderr: true,
	}

	execID, err := dc.cli.ContainerExecCreate(dc.ctx, containerID, execConfig)
	if err != nil {
		return "", err
	}

	resp, err := dc.cli.ContainerExecAttach(dc.ctx, execID.ID, container.ExecStartOptions{})
	if err != nil {
		return "", err
	}
	defer resp.Close()

	var buf bytes.Buffer
	_, err = stdcopy.StdCopy(&buf, &buf, resp.Reader)
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(buf.String()), nil
}