// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

//go:build integration
// +build integration

package commands

import (
	"testing"
	"context"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.opentelemetry.io/build-tools/grater/internal/container"
	"go.opentelemetry.io/build-tools/grater/internal/dockercontainer"
	"go.opentelemetry.io/build-tools/grater/internal/module"
)

func TestShallowClone(t *testing.T) {
	ctx := context.Background()

	var c container.Container

	c, err := dockercontainer.NewDockerContainer()
	require.NoError(t, err)

	useContainerResp, err := c.UseContainer(ctx,
		container.NewUseContainerConfig(
			container.WithImageName("golang:1.22"),
		),
	)

	module := *module.NewModule("github.com/open-telemetry/opentelemetry-go-build-tools", "")
	modulePath := "/mainModule/" + module.ModuleName
	branch := "main"

	err = ShallowClone(ctx, c, useContainerResp, module, branch, modulePath)
	require.NoError(t, err)

	resp, err := c.ExecuteCommand(ctx,
		container.NewExecuteCommandConfig(
			container.WithContainerID(useContainerResp.ContainerID),
			container.WithCommand([]string{"ls", "/mainModule/"}),
		),
	)
	require.NoError(t, err)
	assert.Equal(t, "opentelemetry-go-build-tools", resp.Output)

	resp, err = c.ExecuteCommand(ctx,
    container.NewExecuteCommandConfig(
        container.WithContainerID(useContainerResp.ContainerID),
        container.WithCommand([]string{"git", "-C", modulePath, "rev-parse", "--is-shallow-repository"}),
		),
	)
	require.NoError(t, err)
	assert.Equal(t, "true", resp.Output)
}

func TestCheckoutBranch(t *testing.T) {
	// TODO: Add remote testing functionality for checkout branch.
}

func TestSetReplaceDirective(t *testing.T) {
	ctx := context.Background()

	var c container.Container

	c, err := dockercontainer.NewDockerContainer()
	require.NoError(t, err)

	binds := map[string]string{
		"../testdata/dependent":"/dependent/",
	}
	useContainerResp, err := c.UseContainer(ctx,
		container.NewUseContainerConfig(
			container.WithImageName("golang:1.22"),
			container.WithHostToContainerPaths(binds),
		),
	)
	require.NoError(t, err)

	err = SetReplaceDirective(ctx, c, useContainerResp, "go.opentelemetry.io/build-tools/grater/module", "../moduleFail", "/dependent/")
	require.NoError(t, err)

	resp, err := c.ExecuteCommand(ctx,
	container.NewExecuteCommandConfig(
			container.WithContainerID(useContainerResp.ContainerID),
			container.WithCommand([]string{"cat", "/dependent/go.mod"}),
		),
	)

	assert.Contains(t, resp.Output, `replace go.opentelemetry.io/build-tools/grater/module => ../moduleFail`)
}

func TestRunModuleTest(t *testing.T) {
	ctx := context.Background()

	var c container.Container

	c, err := dockercontainer.NewDockerContainer()
	require.NoError(t, err)

	binds := map[string]string{
		"../testdata/dependent":"/dependent/",
		"../testdata/modulePass":"/modulePass/",
	}
	useContainerResp, err := c.UseContainer(ctx,
		container.NewUseContainerConfig(
			container.WithImageName("golang:1.25.0"),
			container.WithHostToContainerPaths(binds),
		),
	)
	require.NoError(t, err)

	resp, err := RunModuleTest(ctx, c, useContainerResp, "/dependent/")
	require.NoError(t, err)
	assert.Equal(t, 0, resp.ExitCode)
}