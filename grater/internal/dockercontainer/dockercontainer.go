// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

// Package dockercontainer provides an implementation of the Container interface.
package dockercontainer

import (
	"bytes"
	"context"
	"io"
	"strings"

	"github.com/moby/go-archive"
	dockercontainer "github.com/moby/moby/api/types/container"
	"github.com/moby/moby/client"

	"go.opentelemetry.io/build-tools/grater/internal/container"
)

// DockerContainer is a controller for managing Docker containers and volumes.
type DockerContainer struct {
	cli *client.Client
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
	}, nil
}

// CreateVolume creates a volume with specified configuration and returns a create volume response.
func (dc *DockerContainer) CreateVolume(ctx context.Context, cfg container.CreateVolumeConfig) (container.CreateVolumeResponse, error) {
	_, err := dc.cli.VolumeCreate(ctx, client.VolumeCreateOptions{
		Name: cfg.VolumeName(),
	})
	if err != nil {
		return container.CreateVolumeResponse{}, err
	}

	cleanup := func() {
		_, _ = dc.cli.VolumeRemove(ctx, cfg.VolumeName(), client.VolumeRemoveOptions{Force: true})
	}

	return container.NewCreateVolumeResponse(cleanup), nil
}

// UseContainer creates a container with specified configuration and returns a use container response.
func (dc *DockerContainer) UseContainer(ctx context.Context, cfg container.UseContainerConfig) (container.UseContainerResponse, error) {
	if err := dc.pullImage(ctx, cfg.ImageName()); err != nil {
		return container.UseContainerResponse{}, err
	}

	binds := make([]string, 0, len(cfg.BindMounts()))
	for src, dst := range cfg.BindMounts() {
		binds = append(binds, src+":"+dst)
	}

	resp, err := dc.cli.ContainerCreate(ctx, client.ContainerCreateOptions{
		Config: &dockercontainer.Config{
			Image: cfg.ImageName(),
			Cmd:   []string{"sleep", "infinity"},
			Tty:   true,
		},
		HostConfig: &dockercontainer.HostConfig{
			Binds: binds,
		},
	})
	if err != nil {
		return container.UseContainerResponse{}, err
	}

	if _, err := dc.cli.ContainerStart(ctx, resp.ID, client.ContainerStartOptions{}); err != nil {
		return container.UseContainerResponse{}, err
	}

	for hostPath, containerPath := range cfg.HostToContainerPaths() {
		if _, err := dc.ExecuteCommand(ctx, container.NewExecuteCommandConfig(
			container.WithContainerID(resp.ID),
			container.WithCommand([]string{"mkdir", "-p", containerPath}),
		)); err != nil {
			return container.UseContainerResponse{}, err
		}
		tar, err := archive.TarWithOptions(hostPath, &archive.TarOptions{})
		if err != nil {
			return container.UseContainerResponse{}, err
		}
		if _, err := dc.cli.CopyToContainer(ctx, resp.ID, client.CopyToContainerOptions{
			DestinationPath: containerPath,
			Content:         tar,
		}); err != nil {
			return container.UseContainerResponse{}, err
		}
	}

	cleanup := func() {
		if _, err := dc.cli.ContainerStop(ctx, resp.ID, client.ContainerStopOptions{}); err != nil {
			return
		}
		_, _ = dc.cli.ContainerRemove(ctx, resp.ID, client.ContainerRemoveOptions{Force: true})
	}

	return container.NewUseContainerResponse(resp.ID, cleanup), nil
}

// ExecuteCommand executes a command in a container and returns the output.
func (dc *DockerContainer) ExecuteCommand(ctx context.Context, cfg container.ExecuteCommandConfig) (container.ExecuteCommandResponse, error) {
	execID, err := dc.cli.ExecCreate(ctx, cfg.ContainerID(), client.ExecCreateOptions{
		Cmd:          cfg.Cmd(),
		AttachStdout: true,
		AttachStderr: true,
		TTY:          true,
	})
	if err != nil {
		return container.ExecuteCommandResponse{}, err
	}

	resp, err := dc.cli.ExecAttach(ctx, execID.ID, client.ExecAttachOptions{TTY: true})
	if err != nil {
		return container.ExecuteCommandResponse{}, err
	}
	defer resp.Close()

	var buf bytes.Buffer
	if _, err = io.Copy(&buf, resp.Reader); err != nil {
		return container.ExecuteCommandResponse{}, err
	}

	inspect, err := dc.cli.ExecInspect(ctx, execID.ID, client.ExecInspectOptions{})
	if err != nil {
		return container.ExecuteCommandResponse{}, err
	}

	return container.NewExecuteCommandResponse(
		strings.TrimSpace(buf.String()),
		inspect.ExitCode,
	), nil
}

func (dc *DockerContainer) pullImage(ctx context.Context, imageName string) error {
	reader, err := dc.cli.ImagePull(ctx, imageName, client.ImagePullOptions{})
	if err != nil {
		return err
	}
	defer reader.Close()

	_, err = io.Copy(io.Discard, reader)
	return err
}
