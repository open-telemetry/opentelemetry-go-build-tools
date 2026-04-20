// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

// Package dockercontroller provides a controller for managing Docker.
package dockercontroller

import (
	"bytes"
	"context"
	"io"
	"strings"

	dockercontainer "github.com/moby/moby/api/types/container"
	"github.com/moby/moby/client"

	"go.opentelemetry.io/build-tools/grater/internal/container"
)

// DockerContainer is a controller for managing Docker containers and volumes.
type DockerContainer struct {
	cli *client.Client
	ctx context.Context
}

var _ container.Container = (*DockerContainer)(nil)

// NewDockerContainer creates a new Docker container.
func NewDockerContainer() (*DockerContainer, error) {
	cli, err := client.New(client.FromEnv)
	if err != nil {
		return nil, err
	}

	return &DockerContainer{
		cli: cli,
		ctx: context.Background(),
	}, nil
}

// CreateVolume creates a volume with specified volume name and returns a cleanup function.
func (dc *DockerContainer) CreateVolume(volumeName string) (func(), error) {
	_, err := dc.cli.VolumeCreate(dc.ctx, client.VolumeCreateOptions{
		Name: volumeName,
	})
	if err != nil {
		return func() {}, err
	}

	cleanup := func() {
		_, _ = dc.cli.VolumeRemove(dc.ctx, volumeName, client.VolumeRemoveOptions{Force: true})
	}

	return cleanup, nil
}

// UseContainer creates a container with specified volumes and returns a cleanup function.
func (dc *DockerContainer) UseContainer(imageName string, volumeNames []string) (string, func(), error) {
	if err := dc.pullImage(imageName); err != nil {
		return "", nil, err
	}

	binds := make([]string, len(volumeNames))
	for i, v := range volumeNames {
		binds[i] = v + ":/data/" + v
	}

	resp, err := dc.cli.ContainerCreate(dc.ctx, client.ContainerCreateOptions{
		Config: &dockercontainer.Config{
			Image: imageName,
			Cmd:   []string{"sleep", "infinity"},
			Tty:   true,
		},
		HostConfig: &dockercontainer.HostConfig{
			Binds: binds,
		},
	})
	if err != nil {
		return "", nil, err
	}

	if _, err := dc.cli.ContainerStart(dc.ctx, resp.ID, client.ContainerStartOptions{}); err != nil {
		return "", nil, err
	}

	cleanup := func() {
		if _, err := dc.cli.ContainerStop(dc.ctx, resp.ID, client.ContainerStopOptions{}); err != nil {
			return
		}
		_, _ = dc.cli.ContainerRemove(dc.ctx, resp.ID, client.ContainerRemoveOptions{Force: true})
	}

	return resp.ID, cleanup, nil
}

// ExecuteCommand executes a command in a container and returns the output.
func (dc *DockerContainer) ExecuteCommand(containerID string, cmd []string) (string, client.ExecInspectResult, error) {
	execID, err := dc.cli.ExecCreate(dc.ctx, containerID, client.ExecCreateOptions{
		Cmd:          cmd,
		AttachStdout: true,
		AttachStderr: true,
		TTY:          true,
	})
	if err != nil {
		return "", client.ExecInspectResult{}, err
	}

	resp, err := dc.cli.ExecAttach(dc.ctx, execID.ID, client.ExecAttachOptions{TTY: true})
	if err != nil {
		return "", client.ExecInspectResult{}, err
	}
	defer resp.Close()

	var buf bytes.Buffer
	if _, err = io.Copy(&buf, resp.Reader); err != nil {
		return "", client.ExecInspectResult{}, err
	}

	inspect, err := dc.cli.ExecInspect(dc.ctx, execID.ID, client.ExecInspectOptions{})
	if err != nil {
		return "", client.ExecInspectResult{}, err
	}

	return strings.TrimSpace(buf.String()), inspect, nil
}

func (dc *DockerContainer) pullImage(imageName string) error {
	reader, err := dc.cli.ImagePull(dc.ctx, imageName, client.ImagePullOptions{})
	if err != nil {
		return err
	}
	defer reader.Close()

	_, err = io.Copy(io.Discard, reader)
	return err
}
