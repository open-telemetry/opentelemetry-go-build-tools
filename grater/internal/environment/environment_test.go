// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

//go:build integration
// +build integration

package environment

import (
	"testing"
	"context"
	//"path/filepath"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/build-tools/grater/internal/dockercontainer"
	"go.opentelemetry.io/build-tools/grater/internal/module"
	"go.opentelemetry.io/build-tools/grater/internal/container"
)

func TestGetReplacementBindsHostAndRemoteModules(t *testing.T) {
	ctx := context.Background()
	
	dc, err := dockercontainer.NewDockerContainer()
	require.NoError(t, err)

	env := NewEnvironment(dc)

	replacements := [][]module.Module{
		{
			
			*module.NewModule("../testdata/modulePass", ""),
			*module.NewModule("go.opentelemetry.io/otel", "v1.24.0"),
		},
	}

	Cleanup, binds, hostBinds, err := env.getReplacementBinds(ctx, replacements)
	require.NoError(t, err)
	defer Cleanup()

	respUseContainer, err := env.c.UseContainer(ctx,
		container.NewUseContainerConfig(
			container.WithImageName("golang:1.25"),
			container.WithBindMounts(binds),
			container.WithHostToContainerPaths(hostBinds),
		),
	)
	require.NoError(t, err)
	defer respUseContainer.Cleanup()

	respExecuteCommand, err := env.c.ExecuteCommand(ctx,
		container.NewExecuteCommandConfig(
			container.WithContainerID(respUseContainer.ContainerID),
			container.WithCommand([]string{"cat", "/hostReplacements/modulePass/go.mod"}),
		),
	)
	assert.Contains(t, respExecuteCommand.Output, "go.opentelemetry.io/build-tools/grater/internal/testdata/module")
	
	respExecuteCommand, err = env.c.ExecuteCommand(ctx,
		container.NewExecuteCommandConfig(
			container.WithContainerID(respUseContainer.ContainerID),
			container.WithCommand([]string{"cat", "/remoteReplacements/otel/go.mod"}),
		),
	)
	assert.Contains(t, respExecuteCommand.Output, "go.opentelemetry.io/otel")
}

func TestGetReplacementBindsHostModules(t *testing.T) {
	ctx := context.Background()
	
	dc, err := dockercontainer.NewDockerContainer()
	require.NoError(t, err)

	env := NewEnvironment(dc)

	replacements := [][]module.Module{
		{
			
			*module.NewModule("../testdata/modulePass", ""),
			*module.NewModule("../testdata/moduleFail", ""),
		},
	}

	Cleanup, binds, hostBinds, err := env.getReplacementBinds(ctx, replacements)
	require.NoError(t, err)
	defer Cleanup()

	respUseContainer, err := env.c.UseContainer(ctx,
		container.NewUseContainerConfig(
			container.WithImageName("golang:1.25"),
			container.WithBindMounts(binds),
			container.WithHostToContainerPaths(hostBinds),
		),
	)
	require.NoError(t, err)
	defer respUseContainer.Cleanup()

	respExecuteCommand, err := env.c.ExecuteCommand(ctx,
		container.NewExecuteCommandConfig(
			container.WithContainerID(respUseContainer.ContainerID),
			container.WithCommand([]string{"cat", "/hostReplacements/modulePass/go.mod"}),
		),
	)
	assert.Contains(t, respExecuteCommand.Output, "go.opentelemetry.io/build-tools/grater/internal/testdata/module")

	respExecuteCommand, err = env.c.ExecuteCommand(ctx,
		container.NewExecuteCommandConfig(
			container.WithContainerID(respUseContainer.ContainerID),
			container.WithCommand([]string{"cat", "/hostReplacements/moduleFail/go.mod"}),
		),
	)
	assert.Contains(t, respExecuteCommand.Output, "go.opentelemetry.io/build-tools/grater/internal/testdata/module")
}

func TestGetReplacementBindsRemoteModules(t *testing.T) {
	ctx := context.Background()
	
	dc, err := dockercontainer.NewDockerContainer()
	require.NoError(t, err)

	env := NewEnvironment(dc)

	replacements := [][]module.Module{
		{
			
			*module.NewModule("go.opentelemetry.io/otel", "v1.24.0"),
			*module.NewModule("go.opentelemetry.io/otel", "v1.24.0"),
		},
	}

	Cleanup, binds, hostBinds, err := env.getReplacementBinds(ctx, replacements)
	require.NoError(t, err)
	defer Cleanup()

	respUseContainer, err := env.c.UseContainer(ctx,
		container.NewUseContainerConfig(
			container.WithImageName("golang:1.22"),
			container.WithBindMounts(binds),
			container.WithHostToContainerPaths(hostBinds),
		),
	)
	require.NoError(t, err)
	defer respUseContainer.Cleanup()

	respExecuteCommand, err := env.c.ExecuteCommand(ctx,
		container.NewExecuteCommandConfig(
			container.WithContainerID(respUseContainer.ContainerID),
			container.WithCommand([]string{"cat", "/remoteReplacements/otel/go.mod"}),
		),
	)
	assert.Contains(t, respExecuteCommand.Output, "go.opentelemetry.io/otel")

	respExecuteCommand, err = env.c.ExecuteCommand(ctx,
		container.NewExecuteCommandConfig(
			container.WithContainerID(respUseContainer.ContainerID),
			container.WithCommand([]string{"cat", "/remoteReplacements/otel/go.mod"}),
		),
	)
	assert.Contains(t, respExecuteCommand.Output, "go.opentelemetry.io/otel")
}

func TestGetMainModuleBindsHostAndRemoteModules(t *testing.T) {
	ctx := context.Background()

	dc, err := dockercontainer.NewDockerContainer()
	require.NoError(t, err)

	env := NewEnvironment(dc)

	moduleBase := *module.NewModule("../testdata/dependent", "")
	moduleHead := *module.NewModule("go.opentelemetry.io/otel", "v1.24.0")

	Cleanup, binds, hostBinds, err := env.getMainModuleBinds(ctx, moduleBase, moduleHead)
	require.NoError(t, err)
	defer Cleanup()

	respUseContainer, err := env.c.UseContainer(ctx,
        container.NewUseContainerConfig(
            container.WithImageName("golang:1.25"),
            container.WithHostToContainerPaths(hostBinds),
			container.WithBindMounts(binds),
        ),
    )
	require.NoError(t, err)
	defer respUseContainer.Cleanup()

	respExecuteCommand, err := env.c.ExecuteCommand(ctx,
		container.NewExecuteCommandConfig(
			container.WithContainerID(respUseContainer.ContainerID),
			container.WithCommand([]string{"cat", "/hostMainModule/base/" + moduleBase.ModuleName + "/go.mod"}),
		),
	)
	assert.Contains(t, respExecuteCommand.Output, "go.opentelemetry.io/build-tools/grater/internal/testdata/dependent")

	respExecuteCommand, err = env.c.ExecuteCommand(ctx,
		container.NewExecuteCommandConfig(
			container.WithContainerID(respUseContainer.ContainerID),
			container.WithCommand([]string{"cat", "/remoteMainModule/head/" + moduleHead.ModuleName + "/go.mod"}),
		),
	)
	assert.Contains(t, respExecuteCommand.Output, "go.opentelemetry.io/otel")
}

func TestGetMainModuleBindsHostModules(t *testing.T) {
	ctx := context.Background()

	dc, err := dockercontainer.NewDockerContainer()
	require.NoError(t, err)

	env := NewEnvironment(dc)

	moduleBase := *module.NewModule("../testdata/dependent", "")
	moduleHead := *module.NewModule("../testdata/modulePass", "")

	Cleanup, binds, hostBinds, err := env.getMainModuleBinds(ctx, moduleBase, moduleHead)
	require.NoError(t, err)
	defer Cleanup()

	respUseContainer, err := env.c.UseContainer(ctx,
        container.NewUseContainerConfig(
            container.WithImageName("golang:1.25"),
            container.WithHostToContainerPaths(hostBinds),
			container.WithBindMounts(binds),
        ),
    )
	require.NoError(t, err)
	defer respUseContainer.Cleanup()

	respExecuteCommand, err := env.c.ExecuteCommand(ctx,
		container.NewExecuteCommandConfig(
			container.WithContainerID(respUseContainer.ContainerID),
			container.WithCommand([]string{"cat", "/hostMainModule/base/" + moduleBase.ModuleName + "/go.mod"}),
		),
	)
	assert.Contains(t, respExecuteCommand.Output, "go.opentelemetry.io/build-tools/grater/internal/testdata/dependent")

	respExecuteCommand, err = env.c.ExecuteCommand(ctx,
		container.NewExecuteCommandConfig(
			container.WithContainerID(respUseContainer.ContainerID),
			container.WithCommand([]string{"cat", "/hostMainModule/head/" + moduleHead.ModuleName + "/go.mod"}),
		),
	)
	assert.Contains(t, respExecuteCommand.Output, "go.opentelemetry.io/build-tools/grater/internal/testdata/module")
}

func TestGetMainModuleBindsRemoteModules(t *testing.T) {
	ctx := context.Background()

	dc, err := dockercontainer.NewDockerContainer()
	require.NoError(t, err)

	env := NewEnvironment(dc)

	moduleBase := *module.NewModule("go.opentelemetry.io/otel", "v1.24.0")
	moduleHead := *module.NewModule("go.opentelemetry.io/otel", "v1.24.0")

	Cleanup, binds, hostBinds, err := env.getMainModuleBinds(ctx, moduleBase, moduleHead)
	require.NoError(t, err)
	defer Cleanup()

	respUseContainer, err := env.c.UseContainer(ctx,
        container.NewUseContainerConfig(
            container.WithImageName("golang:1.22"),
            container.WithBindMounts(binds),
			container.WithHostToContainerPaths(hostBinds),
        ),
    )
	require.NoError(t, err)
	defer respUseContainer.Cleanup()

	respExecuteCommand, err := env.c.ExecuteCommand(ctx,
		container.NewExecuteCommandConfig(
			container.WithContainerID(respUseContainer.ContainerID),
			container.WithCommand([]string{"cat", "/remoteMainModule/base/" + moduleBase.ModuleName + "/go.mod"}),
		),
	)
	assert.Contains(t, respExecuteCommand.Output, "go.opentelemetry.io/otel")

	respExecuteCommand, err = env.c.ExecuteCommand(ctx,
		container.NewExecuteCommandConfig(
			container.WithContainerID(respUseContainer.ContainerID),
			container.WithCommand([]string{"cat", "/remoteMainModule/head/" + moduleHead.ModuleName + "/go.mod"}),
		),
	)
	assert.Contains(t, respExecuteCommand.Output, "go.opentelemetry.io/otel")
}