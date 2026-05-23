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

func TestGetDependentInContainerRemote(t *testing.T) {
	ctx := context.Background()
	
	dc, err := dockercontainer.NewDockerContainer()
	require.NoError(t, err)

	env := NewEnvironment(dc)

	respUseContainer, err := env.c.UseContainer(ctx,
		container.NewUseContainerConfig(
			container.WithImageName("golang:1.25"),
		),
	)
	require.NoError(t, err)
	defer respUseContainer.Cleanup()

	dependent := *module.NewModule("go.opentelemetry.io/otel", "v1.24.0")
	err = env.getDependentInContainer(ctx, respUseContainer, dependent)
	require.NoError(t, err)

	respExecuteCommand, err := env.c.ExecuteCommand(ctx,
		container.NewExecuteCommandConfig(
			container.WithContainerID(respUseContainer.ContainerID),
			container.WithCommand([]string{"cat", "/dependent/otelv1.24.0/go.mod"}),
		),
	)
	assert.Contains(t, respExecuteCommand.Output, "go.opentelemetry.io/otel")
}

func TestGetDependentInContainerHost(t *testing.T) {
	ctx := context.Background()
	
	dc, err := dockercontainer.NewDockerContainer()
	require.NoError(t, err)

	env := NewEnvironment(dc)

	respUseContainer, err := env.c.UseContainer(ctx,
		container.NewUseContainerConfig(
			container.WithImageName("golang:1.25"),
		),
	)
	require.NoError(t, err)
	defer respUseContainer.Cleanup()

	dependent := *module.NewModule("../testdata/modulePass", "")
	err = env.getDependentInContainer(ctx, respUseContainer, dependent)
	require.NoError(t, err)

	respExecuteCommand, err := env.c.ExecuteCommand(ctx,
		container.NewExecuteCommandConfig(
			container.WithContainerID(respUseContainer.ContainerID),
			container.WithCommand([]string{"cat", "/dependent/modulePass/go.mod"}),
		),
	)
	assert.Contains(t, respExecuteCommand.Output, "go.opentelemetry.io/build-tools/grater/internal/testdata/module")
}

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

	Cleanup, binds, err := env.getReplacementBinds(ctx, replacements)
	require.NoError(t, err)
	defer Cleanup()

	respUseContainer, err := env.c.UseContainer(ctx,
		container.NewUseContainerConfig(
			container.WithImageName("golang:1.25"),
			container.WithBindMounts(binds),
		),
	)
	require.NoError(t, err)
	defer respUseContainer.Cleanup()

	respExecuteCommand, err := env.c.ExecuteCommand(ctx,
		container.NewExecuteCommandConfig(
			container.WithContainerID(respUseContainer.ContainerID),
			container.WithCommand([]string{"cat", "/replacements/modulePass/go.mod"}),
		),
	)
	assert.Contains(t, respExecuteCommand.Output, "go.opentelemetry.io/build-tools/grater/internal/testdata/module")
	
	respExecuteCommand, err = env.c.ExecuteCommand(ctx,
		container.NewExecuteCommandConfig(
			container.WithContainerID(respUseContainer.ContainerID),
			container.WithCommand([]string{"cat", "/replacements/otelv1.24.0/go.mod"}),
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

	Cleanup, binds, err := env.getReplacementBinds(ctx, replacements)
	require.NoError(t, err)
	defer Cleanup()

	respUseContainer, err := env.c.UseContainer(ctx,
		container.NewUseContainerConfig(
			container.WithImageName("golang:1.25"),
			container.WithBindMounts(binds),
		),
	)
	require.NoError(t, err)
	defer respUseContainer.Cleanup()

	respExecuteCommand, err := env.c.ExecuteCommand(ctx,
		container.NewExecuteCommandConfig(
			container.WithContainerID(respUseContainer.ContainerID),
			container.WithCommand([]string{"cat", "/replacements/modulePass/go.mod"}),
		),
	)
	assert.Contains(t, respExecuteCommand.Output, "go.opentelemetry.io/build-tools/grater/internal/testdata/module")

	respExecuteCommand, err = env.c.ExecuteCommand(ctx,
		container.NewExecuteCommandConfig(
			container.WithContainerID(respUseContainer.ContainerID),
			container.WithCommand([]string{"cat", "/replacements/moduleFail/go.mod"}),
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
			*module.NewModule("go.opentelemetry.io/otel", "v1.23.0"),
		},
	}

	Cleanup, binds, err := env.getReplacementBinds(ctx, replacements)
	require.NoError(t, err)
	defer Cleanup()

	respUseContainer, err := env.c.UseContainer(ctx,
		container.NewUseContainerConfig(
			container.WithImageName("golang:1.22"),
			container.WithBindMounts(binds),
		),
	)
	require.NoError(t, err)
	defer respUseContainer.Cleanup()

	respExecuteCommand, err := env.c.ExecuteCommand(ctx,
		container.NewExecuteCommandConfig(
			container.WithContainerID(respUseContainer.ContainerID),
			container.WithCommand([]string{"cat", "/replacements/otelv1.24.0/go.mod"}),
		),
	)
	assert.Contains(t, respExecuteCommand.Output, "go.opentelemetry.io/otel")

	respExecuteCommand, err = env.c.ExecuteCommand(ctx,
		container.NewExecuteCommandConfig(
			container.WithContainerID(respUseContainer.ContainerID),
			container.WithCommand([]string{"cat", "/replacements/otelv1.23.0/go.mod"}),
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

	Cleanup, binds, err := env.getMainModuleBinds(ctx, moduleBase, moduleHead)
	require.NoError(t, err)
	defer Cleanup()

	respUseContainer, err := env.c.UseContainer(ctx,
        container.NewUseContainerConfig(
            container.WithImageName("golang:1.25"),
			container.WithBindMounts(binds),
        ),
    )
	require.NoError(t, err)
	defer respUseContainer.Cleanup()

	respExecuteCommand, err := env.c.ExecuteCommand(ctx,
		container.NewExecuteCommandConfig(
			container.WithContainerID(respUseContainer.ContainerID),
			container.WithCommand([]string{"cat", "/mainModule/" + moduleBase.ModuleName + moduleBase.ModuleVersion + "/go.mod"}),
		),
	)
	assert.Contains(t, respExecuteCommand.Output, "go.opentelemetry.io/build-tools/grater/internal/testdata/dependent")

	respExecuteCommand, err = env.c.ExecuteCommand(ctx,
		container.NewExecuteCommandConfig(
			container.WithContainerID(respUseContainer.ContainerID),
			container.WithCommand([]string{"cat", "/mainModule/" + moduleHead.ModuleName + moduleHead.ModuleVersion + "/go.mod"}),
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

	Cleanup, binds, err := env.getMainModuleBinds(ctx, moduleBase, moduleHead)
	require.NoError(t, err)
	defer Cleanup()

	respUseContainer, err := env.c.UseContainer(ctx,
        container.NewUseContainerConfig(
            container.WithImageName("golang:1.25"),
			container.WithBindMounts(binds),
        ),
    )
	require.NoError(t, err)
	defer respUseContainer.Cleanup()

	respExecuteCommand, err := env.c.ExecuteCommand(ctx,
		container.NewExecuteCommandConfig(
			container.WithContainerID(respUseContainer.ContainerID),
			container.WithCommand([]string{"cat", "/mainModule/" + moduleBase.ModuleName + moduleBase.ModuleVersion + "/go.mod"}),
		),
	)
	assert.Contains(t, respExecuteCommand.Output, "go.opentelemetry.io/build-tools/grater/internal/testdata/dependent")

	respExecuteCommand, err = env.c.ExecuteCommand(ctx,
		container.NewExecuteCommandConfig(
			container.WithContainerID(respUseContainer.ContainerID),
			container.WithCommand([]string{"cat", "/mainModule/" + moduleHead.ModuleName + moduleHead.ModuleVersion + "/go.mod"}),
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
	moduleHead := *module.NewModule("go.opentelemetry.io/otel", "v1.23.0")

	Cleanup, binds, err := env.getMainModuleBinds(ctx, moduleBase, moduleHead)
	require.NoError(t, err)
	defer Cleanup()

	respUseContainer, err := env.c.UseContainer(ctx,
        container.NewUseContainerConfig(
            container.WithImageName("golang:1.22"),
            container.WithBindMounts(binds),
        ),
    )
	require.NoError(t, err)
	defer respUseContainer.Cleanup()

	respExecuteCommand, err := env.c.ExecuteCommand(ctx,
		container.NewExecuteCommandConfig(
			container.WithContainerID(respUseContainer.ContainerID),
			container.WithCommand([]string{"cat", "/mainModule/" + moduleBase.ModuleName + moduleBase.ModuleVersion + "/go.mod"}),
		),
	)
	assert.Contains(t, respExecuteCommand.Output, "go.opentelemetry.io/otel")

	respExecuteCommand, err = env.c.ExecuteCommand(ctx,
		container.NewExecuteCommandConfig(
			container.WithContainerID(respUseContainer.ContainerID),
			container.WithCommand([]string{"cat", "/mainModule/" + moduleHead.ModuleName + moduleHead.ModuleVersion + "/go.mod"}),
		),
	)
	assert.Contains(t, respExecuteCommand.Output, "go.opentelemetry.io/otel")
}

func TestGetModuleInContainerHost(t *testing.T) {
	ctx := context.Background()

	dc, err := dockercontainer.NewDockerContainer()
	require.NoError(t, err)

	env := NewEnvironment(dc)

	respUseContainer, err := env.c.UseContainer(ctx,
        container.NewUseContainerConfig(
            container.WithImageName("golang:1.22"),
        ),
    )
	require.NoError(t, err)
	defer respUseContainer.Cleanup()

	module := *module.NewModule("../testdata/modulePass", "")
	modulePath := "/modulePath/" + module.ModuleName

	err = env.getModuleInContainer(ctx, respUseContainer, module, modulePath)
	require.NoError(t, err)

	respExecuteCommand, err := env.c.ExecuteCommand(ctx,
		container.NewExecuteCommandConfig(
			container.WithContainerID(respUseContainer.ContainerID),
			container.WithCommand([]string{"cat", modulePath + "/go.mod"}),
		),
	)
	assert.Contains(t, respExecuteCommand.Output, "go.opentelemetry.io/build-tools/grater/internal/testdata/module")
}

func TestGetModuleInContainerRemote(t *testing.T) {
	ctx := context.Background()

	dc, err := dockercontainer.NewDockerContainer()
	require.NoError(t, err)

	env := NewEnvironment(dc)

	respUseContainer, err := env.c.UseContainer(ctx,
        container.NewUseContainerConfig(
            container.WithImageName("golang:1.22"),
        ),
    )
	require.NoError(t, err)
	defer respUseContainer.Cleanup()

	module := *module.NewModule("go.opentelemetry.io/otel", "v1.24.0")
	modulePath := "/modulePath/" + module.ModuleName

	err = env.getModuleInContainer(ctx, respUseContainer, module, modulePath)
	require.NoError(t, err)

	respExecuteCommand, err := env.c.ExecuteCommand(ctx,
		container.NewExecuteCommandConfig(
			container.WithContainerID(respUseContainer.ContainerID),
			container.WithCommand([]string{"cat", modulePath + "/go.mod"}),
		),
	)
	assert.Contains(t, respExecuteCommand.Output, "go.opentelemetry.io/otel")
}