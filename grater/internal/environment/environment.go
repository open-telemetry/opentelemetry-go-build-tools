// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package environment

import (
	"context"
    "path/filepath"

	"go.opentelemetry.io/build-tools/grater/internal/module"
	"go.opentelemetry.io/build-tools/grater/internal/commands"
	"go.opentelemetry.io/build-tools/grater/internal/container"
)

// Environment struct initialises an enviornment to run tests.
type Environment struct {
	c container.Container
}

// NewEnvironment creates an instance of an environment.
func NewEnvironment(c container.Container) *Environment {
	return &Environment{c: c}
}

// RunTests runs tests of dependents of main module with specified replacements.
func (env *Environment) RunTests(ctx context.Context, mainModuleBase, mainModuleHead module.Module, dependents []module.Module, replacements [][]module.Module) {
	// Setup main module
	// setup replacements
	// Set up container with binds
	// for every set up container run tests
}

func (env *Environment) getMainModuleBinds(ctx context.Context, mainModuleBase, headModuleBase module.Module) (func(), map[string]string, map[string]string, error) {
    binds := make(map[string]string)
    hostBinds := make(map[string]string)
    volumeName := "remote_main_module_volume"

    respCreateVolume, err := env.c.CreateVolume(ctx,
        container.NewCreateVolumeConfig(
            container.WithVolumeName(volumeName),
        ),
    )
    if err != nil {
        return nil, map[string]string{}, map[string]string{}, err
    }

    respUseContainer, err := env.c.UseContainer(ctx,
        container.NewUseContainerConfig(
            container.WithImageName("golang:1.22"),
            container.WithBindMounts(map[string]string{volumeName: "/remoteMainModule"}),
        ),
    )
    if err != nil {
        respCreateVolume.Cleanup()
        return nil, map[string]string{}, map[string]string{}, err
    }
    defer respUseContainer.Cleanup()

    if mainModuleBase.IsRemotePath() {
        modulePath := "/remoteMainModule/base/" + mainModuleBase.ModuleName
        _, err := env.c.ExecuteCommand(ctx,
            container.NewExecuteCommandConfig(
                container.WithContainerID(respUseContainer.ContainerID),
                container.WithCommand([]string{"mkdir", "-p", modulePath}),
            ),
        )
        if err != nil {
            respCreateVolume.Cleanup()
            return nil, map[string]string{}, map[string]string{}, err
        }
        _ = commands.GetModuleFromProxy(ctx, env.c, respUseContainer, mainModuleBase, modulePath)
    } else {
        absPath, err := filepath.Abs(mainModuleBase.ModulePath)
        if err != nil {
            respCreateVolume.Cleanup()
            return nil, map[string]string{}, map[string]string{}, err
        }
        hostBinds[absPath] = "/hostMainModule/base/" + mainModuleBase.ModuleName
    }

    if headModuleBase.IsRemotePath() {
        modulePath := "/remoteMainModule/head/" + headModuleBase.ModuleName
        _, err := env.c.ExecuteCommand(ctx,
            container.NewExecuteCommandConfig(
                container.WithContainerID(respUseContainer.ContainerID),
                container.WithCommand([]string{"mkdir", "-p", modulePath}),
            ),
        )
        if err != nil {
            respCreateVolume.Cleanup()
            return nil, map[string]string{}, map[string]string{}, err
        }
        _ = commands.GetModuleFromProxy(ctx, env.c, respUseContainer, headModuleBase, modulePath)
    } else {
        absPath, err := filepath.Abs(headModuleBase.ModulePath)
        if err != nil {
            respCreateVolume.Cleanup()
            return nil, map[string]string{}, map[string]string{}, err
        }
        hostBinds[absPath] = "/hostMainModule/head/" + headModuleBase.ModuleName
    }
    binds[volumeName] = "/remoteMainModule"

    return respCreateVolume.Cleanup, binds, hostBinds, nil
}

func mergeMaps(a, b map[string]string) map[string]string {
	result := make(map[string]string, len(a)+len(b))
	for k, v := range a {
		result[k] = v
	}
	for k, v := range b {
		result[k] = v
	}
	return result
}