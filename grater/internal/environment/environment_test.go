// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

//go:build integration
// +build integration

package environment

import (
	"testing"
	"context"
	"path/filepath"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/build-tools/grater/internal/dockercontainer"
	"go.opentelemetry.io/build-tools/grater/internal/module"
	"go.opentelemetry.io/build-tools/grater/internal/container"
	"go.opentelemetry.io/build-tools/grater/internal/commands"
)

func TestRunTests(t *testing.T) {
	// TODO: Add an E2E test.
}

func TestRunTest(t *testing.T) {
    ctx := context.Background()

    dc, err := dockercontainer.NewDockerContainer()
    require.NoError(t, err)

    env := NewEnvironment(dc)

    mainModuleBase := *module.NewModule("go.opentelemetry.io/build-tools/grater/internal/testdata/module", "")
    mainModuleHead := *module.NewModule("../testdata/moduleFail", "")

    replacements := [][]module.Module{}

    CleanupMainModule, bindsMainModule, err := env.getMainModuleBinds(ctx, mainModuleHead)
    require.NoError(t, err)
    defer CleanupMainModule()

    CleanupReplacements, bindsReplacements, err := env.getReplacementBinds(ctx, replacements)
    require.NoError(t, err)
    defer CleanupReplacements()

    dependent := *module.NewModule("../testdata/dependent", "")
	dependentPath := "/dependent/" + dependent.ModuleName + dependent.ModuleVersion
    binds := mergeMaps(bindsMainModule, bindsReplacements)

    respUseContainer, err := env.getRunTestContainer(ctx, binds, dependent, replacements)
    require.NoError(t, err)
    defer respUseContainer.Cleanup()

	// Route the local path dependencies to container's local paths.
	externalModule := *module.NewModule("../testdata/modulePass", "")
	absPath, err := filepath.Abs(externalModule.ModulePath)
	require.NoError(t, err)
	err = env.c.CopyToContainer(ctx, respUseContainer.ContainerID, map[string]string{absPath:"/external/modulePass"})
	require.NoError(t, err)
	err = commands.SetReplaceDirective(ctx, env.c, respUseContainer, "go.opentelemetry.io/build-tools/grater/internal/testdata/module", "../../external/modulePass", dependentPath)
	require.NoError(t, err)

    results, err := env.runTest(ctx, respUseContainer, mainModuleBase, mainModuleHead, dependent)
    require.NoError(t, err)

    assert.Equal(t, 0, results[0].ExitCode)
    assert.Equal(t, 1, results[1].ExitCode)
}

func TestSetUpRunTestContainer(t *testing.T) {
	ctx := context.Background()
	
	dc, err := dockercontainer.NewDockerContainer()
	require.NoError(t, err)

	env := NewEnvironment(dc)

	replacements := [][]module.Module{
		{
			*module.NewModule("go.opentelemetry.io/build-tools/grater/internal/testdata/module", "",),
			*module.NewModule("../testdata/modulePass", ""),
		},
		{
			*module.NewModule("go.opentelemetry.io/otel", "v1.24.0"),
			*module.NewModule("go.opentelemetry.io/otel", "v1.23.0"),
		},
	}
	CleanupReplacements, bindsReplacements, err := env.getReplacementBinds(ctx, replacements)
	require.NoError(t, err)
	defer CleanupReplacements()

	mainModuleHead := *module.NewModule("../testdata/moduleFail", "")
	CleanupMainModule, bindsMainModule, err := env.getMainModuleBinds(ctx, mainModuleHead)
	require.NoError(t, err)
	defer CleanupMainModule()

	dependent := *module.NewModule("../testdata/dependent", "")
	binds := mergeMaps(bindsMainModule, bindsReplacements)

	respUseContainer, err := env.getRunTestContainer(ctx, binds, dependent, replacements)
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
			container.WithCommand([]string{"cat", "/replacements/otelv1.23.0/go.mod"}),
		),
	)
	assert.Contains(t, respExecuteCommand.Output, "go.opentelemetry.io/otel")

	respExecuteCommand, err = env.c.ExecuteCommand(ctx,
		container.NewExecuteCommandConfig(
			container.WithContainerID(respUseContainer.ContainerID),
			container.WithCommand([]string{"cat", "/dependent/dependent/go.mod"}),
		),
	)
	assert.Contains(t, respExecuteCommand.Output, "go.opentelemetry.io/build-tools/grater/internal/testdata/dependent")

	respExecuteCommand, err = env.c.ExecuteCommand(ctx,
		container.NewExecuteCommandConfig(
			container.WithContainerID(respUseContainer.ContainerID),
			container.WithCommand([]string{"cat", "/mainModule/moduleFail/go.mod"}),
		),
	)
	assert.Contains(t, respExecuteCommand.Output, "go.opentelemetry.io/build-tools/grater/internal/testdata/module")

	resp, err := env.c.ExecuteCommand(ctx,
		container.NewExecuteCommandConfig(
			container.WithContainerID(respUseContainer.ContainerID),
			container.WithCommand([]string{"cat", "/dependent/" + dependent.ModuleName + dependent.ModuleVersion + "/go.mod"}),
		),
	)

	assert.Contains(t, resp.Output, "replace go.opentelemetry.io/build-tools/grater/internal/testdata/module => ../../replacements/modulePass")
	assert.Contains(t, resp.Output, "replace go.opentelemetry.io/otel v1.24.0 => ../../replacements/otelv1.23.0")
}

func TestSetReplaceDirectivesForDependent(t *testing.T) {
	ctx := context.Background()
	
	dc, err := dockercontainer.NewDockerContainer()
	require.NoError(t, err)

	env := NewEnvironment(dc)

	replacements := [][]module.Module{
		{
			*module.NewModule("go.opentelemetry.io/build-tools/grater/internal/testdata/module", "",),
			*module.NewModule("../testdata/modulePass", ""),
		},
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
			container.WithImageName("golang:1.25"),
			container.WithBindMounts(binds),
		),
	)
	require.NoError(t, err)
	defer respUseContainer.Cleanup()

	dependent := *module.NewModule("../testdata/dependent", "")
	err = env.getDependentInContainer(ctx, respUseContainer, dependent)
	require.NoError(t, err)

	err = env.setReplaceDirectivesForDependent(ctx, respUseContainer, dependent, replacements)
	require.NoError(t, err)

	resp, err := env.c.ExecuteCommand(ctx,
		container.NewExecuteCommandConfig(
			container.WithContainerID(respUseContainer.ContainerID),
			container.WithCommand([]string{"cat", "/dependent/" + dependent.ModuleName + dependent.ModuleVersion + "/go.mod"}),
		),
	)

	assert.Contains(t, resp.Output, "replace go.opentelemetry.io/build-tools/grater/internal/testdata/module => ../../replacements/modulePass")
	assert.Contains(t, resp.Output, "replace go.opentelemetry.io/otel v1.24.0 => ../../replacements/otelv1.23.0")
}

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
	assert.Equal(t, respExecuteCommand.ExitCode, 1)
	
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
	assert.Equal(t, 1, respExecuteCommand.ExitCode)

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
	assert.Equal(t, respExecuteCommand.ExitCode, 1)

	respExecuteCommand, err = env.c.ExecuteCommand(ctx,
		container.NewExecuteCommandConfig(
			container.WithContainerID(respUseContainer.ContainerID),
			container.WithCommand([]string{"cat", "/replacements/otelv1.23.0/go.mod"}),
		),
	)
	assert.Contains(t, respExecuteCommand.Output, "go.opentelemetry.io/otel")
}

func TestGetMainModuleBindsHostModules(t *testing.T) {
	ctx := context.Background()

	dc, err := dockercontainer.NewDockerContainer()
	require.NoError(t, err)

	env := NewEnvironment(dc)

	moduleHead := *module.NewModule("../testdata/modulePass", "")

	Cleanup, binds, err := env.getMainModuleBinds(ctx, moduleHead)
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

	moduleHead := *module.NewModule("go.opentelemetry.io/otel", "v1.23.0")

	Cleanup, binds, err := env.getMainModuleBinds(ctx, moduleHead)
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

func TestMergeMaps(t *testing.T) {
    a := map[string]string{"x": "1", "y": "2"}
    b := map[string]string{"z": "3"}
    result := mergeMaps(a, b)
    assert.Equal(t, map[string]string{"x": "1", "y": "2", "z": "3"}, result)
}