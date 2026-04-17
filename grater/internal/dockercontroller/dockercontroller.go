// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

//go:build integration
// +build integration

// Package dockercontroller provides a controller for managing Docker.
package dockercontroller

import (
	"io"
	"bytes"
	"context"
	"strings"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/volume"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
	"go.opentelemetry.io/build-tools/grater/internal/controller"
)

// DockerController is a controller for managing Docker containers and volumes.
type DockerController struct {
	cli     *client.Client
	ctx     context.Context
	volumes []string
}

var _ controller.Controller = (*DockerController)(nil)

// NewDockerController creates a new Docker controller.
func NewDockerController() (*DockerController, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, err
	}

	return &DockerController{
		cli:     cli,
		ctx:     context.Background(),
		volumes: []string{},
	}, nil
}

// CreateVolume creates a volume with specified volume name and returns a cleanup function.
func (dc *DockerController) CreateVolume(volumeName string) (func(), error) {
	_, err := dc.cli.VolumeCreate(dc.ctx, volume.CreateOptions{
		Name: volumeName,
	})
	if err != nil {
		return func() {}, err
	}

	cleanup := func() {
		_ = dc.cli.VolumeRemove(dc.ctx, volumeName, true)
	}

	return cleanup, nil
}

// UseContainer creates a container with specified volumes and returns a cleanup function.
func (dc *DockerController) UseContainer(imageName string, volumeNames []string) (string, func(), error) {
	if err := dc.pullImage(imageName); err != nil {
		return "", nil, err
	}

	binds := make([]string, len(volumeNames))
	for i, v := range volumeNames {
		binds[i] = v + ":/data/" + v // Path inside container of format /data/<volume_name>
	}

	resp, err := dc.cli.ContainerCreate(
		dc.ctx,
		&container.Config{
			Image:   imageName,
			Cmd:     []string{"sleep", "infinity"},
			Tty:     true,
		},
		&container.HostConfig{
			Binds: binds,
		},
		nil, nil, "",
	)
	if err != nil {
		return "", nil, err
	}

	if err := dc.cli.ContainerStart(dc.ctx, resp.ID, container.StartOptions{}); err != nil {
		return "", nil, err
	}

	cleanup := func() {
		err := dc.cli.ContainerStop(dc.ctx, resp.ID, container.StopOptions{})
		if err != nil {
			return
		}
		err = dc.cli.ContainerRemove(dc.ctx, resp.ID, container.RemoveOptions{Force: true})
		if err != nil {
			return
		}
	}

	return resp.ID, cleanup, nil
}

// ExecuteCommand executes a command in a container and returns the output.
func (dc *DockerController) ExecuteCommand(containerID string, cmd []string) (string, container.ExecInspect, error) {
	execConfig := container.ExecOptions{
		Cmd:          cmd,
		AttachStdout: true,
		AttachStderr: true,
	}

	execID, err := dc.cli.ContainerExecCreate(dc.ctx, containerID, execConfig)
	if err != nil {
		return "", container.ExecInspect{}, err
	}

	resp, err := dc.cli.ContainerExecAttach(dc.ctx, execID.ID, container.ExecStartOptions{})
	if err != nil {
		return "", container.ExecInspect{}, err
	}
	defer resp.Close()

	var buf bytes.Buffer
	_, err = stdcopy.StdCopy(&buf, &buf, resp.Reader)
	if err != nil {
		return "", container.ExecInspect{}, err
	}

	inspect, err := dc.cli.ContainerExecInspect(dc.ctx, execID.ID)
	if err != nil {
		return "", container.ExecInspect{}, err
	}

	return strings.TrimSpace(buf.String()), inspect, nil
}

func (dc *DockerController) pullImage(imageName string) error {
	reader, err := dc.cli.ImagePull(dc.ctx, imageName, image.PullOptions{})
	if err != nil {
		return err
	}
	defer reader.Close()

	_, err = io.Copy(io.Discard, reader)
	if err != nil {
		return err
	}

	return nil
}