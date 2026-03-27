// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package internal

import (
	"context"
	"fmt"
	"os"
	"io"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/volume"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
)

func RunTests(repo, base, head string) error {
	cli, volName, err := setUpDockerEnvironment(repo, base, head)
	if err != nil { return err }
	defer cli.Close()

	dependents := []string{"example-dep"}

	for _, d := range dependents {
		if err := runTest(volName, repo, base, head, d, cli); err != nil {
			return err
		}
	}
	return nil
}

func setUpDockerEnvironment(repo, base, head string) (*client.Client, string, error) {
	ctx := context.Background()
	volName := "library-clone"
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil { return nil, "", err }

	_, err = cli.VolumeCreate(ctx, volume.CreateOptions{Name: volName})
	if err != nil { return nil, "", err }

	ctxContainer, respContainer, err := createContainer(volName, "alpine/git", cli)
	if err != nil { return nil, "", err }

	if err := cli.ContainerStart(ctxContainer, respContainer.ID, container.StartOptions{}); err != nil {
		return nil, "", err
	}

	if err := runCommand(ctxContainer, cli, respContainer, []string{"git", "clone", repo, "."}); err != nil {
		return nil, "", err
	}
	if err := runCommand(ctxContainer, cli, respContainer, []string{"git", "checkout", head}); err != nil {
		return nil, "", err
	}
	if err := runCommand(ctxContainer, cli, respContainer, []string{"git", "checkout", base}); err != nil {
		return nil, "", err
	}

	cli.ContainerStop(ctxContainer, respContainer.ID, container.StopOptions{})

	return cli, volName, nil
}

func runTest(volName, repo, base, head, dependent string, cli *client.Client) error {
	ctx, resp, err := createContainer(volName, "golang:1.24-alpine", cli)
	if err != nil { 
		return err 
	}
	defer func() {
		cli.ContainerStop(ctx, resp.ID, container.StopOptions{})
	}()

	if err := cli.ContainerStart(ctx, resp.ID, container.StartOptions{}); err != nil {
		return err
	}

	if err := runCommand(ctx, cli, resp, []string{"git", "clone", dependent, "--depth", "1", "dependent"}); err != nil {
		return err
	}

	if err := runCommand(ctx, cli, resp, []string{"sh", "-c", "cd dependent && go inject /src"}); err != nil {
		return err
	}

	if err := runBuildCommandsInDir(ctx, cli, resp, "dependent"); err != nil {
		return err
	}
	if err := runTestCommandsInDir(ctx, cli, resp, "dependent"); err != nil {
		return err
	}

	if err := runCommand(ctx, cli, resp, []string{"git", "-C", "/src", "checkout", head}); err != nil {
		return err
	}

	if err := runCommand(ctx, cli, resp, []string{"sh", "-c", "cd dependent && go inject /src"}); err != nil {
		return err
	}

	if err := runBuildCommandsInDir(ctx, cli, resp, "dependent"); err != nil {
		return err
	}
	if err := runTestCommandsInDir(ctx, cli, resp, "dependent"); err != nil {
		return err
	}

	return nil
}

func runBuildCommandsInDir(ctx context.Context, cli *client.Client, resp *container.CreateResponse, dir string) error {
	err1 := runCommand(ctx, cli, resp, []string{"sh", "-c", fmt.Sprintf("cd %s && go build ./...", dir)})
	err2 := runCommand(ctx, cli, resp, []string{"sh", "-c", fmt.Sprintf("cd %s && make", dir)})
	
	if err1 != nil && err2 != nil {
		return fmt.Errorf("both go build and make failed in %s: go build error: %v, make error: %v", dir, err1, err2)
	}

	return nil
}

func runTestCommandsInDir(ctx context.Context, cli *client.Client, resp *container.CreateResponse, dir string) error {
	err1 := runCommand(ctx, cli, resp, []string{"sh", "-c", fmt.Sprintf("cd %s && go test ./...", dir)})
	err2 := runCommand(ctx, cli, resp, []string{"sh", "-c", fmt.Sprintf("cd %s && make test", dir)})
	
	if err1 != nil && err2 != nil {
		return fmt.Errorf("both go test and make test failed in %s: go test error: %v, make test error: %v", dir, err1, err2)
	}

	return nil
}

func createContainer(volName, imageName string, cli *client.Client) (context.Context, *container.CreateResponse, error) {
	ctx := context.Background()
	reader, err := cli.ImagePull(ctx, imageName, image.PullOptions{})
	if err != nil {
		return nil, nil, err
	}
	defer reader.Close()

	io.Copy(io.Discard, reader)
	resp, err := cli.ContainerCreate(ctx,
		&container.Config{
			Image:      imageName,
			Cmd:        []string{"sleep", "infinity"},
			WorkingDir: "/src",
		}, &container.HostConfig{
			Binds: []string{volName + ":/src"},
		}, nil, nil, "")
	if err != nil { return nil, nil, err }

	return ctx, &resp, nil
}

func runCommand(ctx context.Context, cli *client.Client, resp *container.CreateResponse, cmd []string) error {
	execConfig := container.ExecOptions{
		AttachStdout: true,
		AttachStderr: true,
		Cmd:          cmd,
		WorkingDir:   "/src",
	}

	execID, err := cli.ContainerExecCreate(ctx, resp.ID, execConfig)
	if err != nil { return err }

	attach, err := cli.ContainerExecAttach(ctx, execID.ID, container.ExecStartOptions{})
	if err != nil { return err }
	defer attach.Close()
	
	stdcopy.StdCopy(os.Stdout, os.Stderr, attach.Reader)

	inspect, err := cli.ContainerExecInspect(ctx, execID.ID)
	if err != nil { return err }

	if inspect.ExitCode != 0 {
		return fmt.Errorf("command failed with exit code %d: %v", inspect.ExitCode, cmd)
	}

	return nil
}
