// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package environment

import (
	"context"
    "fmt"
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
	
}

func (env *Environment) runTest(ctx context.Context, respUseContainer container.UseContainerResponse, mainModuleBase, mainModuleHead, dependent module.Module) ([]container.ExecuteCommandResponse, error) {
    dependentPath := "/dependent/" + dependent.ModuleName + dependent.ModuleVersion

    executeCommandRespBase, err := commands.RunModuleTest(ctx, env.c, respUseContainer, dependentPath)
    if err != nil {
        return nil, err
    }

    oldRef := mainModuleBase.ModulePath
    if mainModuleBase.ModuleVersion != "" {
        oldRef = fmt.Sprintf("%s@%s", mainModuleBase.ModulePath, mainModuleBase.ModuleVersion)
    }
    newRef := "../../mainModule/" + mainModuleHead.ModuleName + mainModuleHead.ModuleVersion

    err = commands.SetReplaceDirective(ctx, env.c, respUseContainer, oldRef, newRef, dependentPath)
    if err != nil {
        return nil, err
    }

	respExecuteCommand, err := env.c.ExecuteCommand(ctx,
		container.NewExecuteCommandConfig(
			container.WithContainerID(respUseContainer.ContainerID),
			container.WithCommand([]string{"cat", "/dependent/dependent/go.mod"}),
		),
	)
    fmt.Println(respExecuteCommand.Output)
    executeCommandRespHead, err := commands.RunModuleTest(ctx, env.c, respUseContainer, dependentPath)
    if err != nil {
        return nil, err
    }

    defer respUseContainer.Cleanup()

    return []container.ExecuteCommandResponse{executeCommandRespBase, executeCommandRespHead}, nil
}

func (env *Environment) getRunTestContainer(ctx context.Context, binds map[string]string, dependent module.Module, replacements [][]module.Module) (container.UseContainerResponse, error) {
    respUseContainer, err := env.c.UseContainer(ctx,
        container.NewUseContainerConfig(
            container.WithImageName("golang:1.25"),
            container.WithBindMounts(binds),
        ),
    )
    if err != nil {
        return container.UseContainerResponse{}, err
    }

    err = env.getDependentInContainer(ctx, respUseContainer, dependent)
    if err != nil {
        return container.UseContainerResponse{}, err
    }

    err = env.setReplaceDirectivesForDependent(ctx, respUseContainer, dependent, replacements)
    if err != nil {
        return container.UseContainerResponse{}, err
    }

    return respUseContainer, nil
}

func (env *Environment) setReplaceDirectivesForDependent(ctx context.Context, respUseContainer container.UseContainerResponse, dependent module.Module, replacements [][]module.Module) error {
    for _, replacementPair := range replacements {
        oldModule, newModule := replacementPair[0], replacementPair[1]
        oldRef := oldModule.ModulePath
        if oldModule.ModuleVersion != "" {
            oldRef = fmt.Sprintf("%s@%s", oldModule.ModulePath, oldModule.ModuleVersion)
        }
        newRef := "../replacements/" + newModule.ModuleName + newModule.ModuleVersion
        err := commands.SetReplaceDirective(ctx, env.c, respUseContainer, oldRef, newRef, "/dependent/" + dependent.ModuleName + dependent.ModuleVersion)
        if err != nil {
            return err
        }
    }
    return nil
}

func (env *Environment) getDependentInContainer(ctx context.Context, respUseContainer container.UseContainerResponse, dependent module.Module) error {
    dependentPath := "/dependent/" + dependent.ModuleName + dependent.ModuleVersion
    err := env.getModuleInContainer(ctx, respUseContainer, dependent, dependentPath)
    if err != nil {
        return err
    }
    return nil
}

func (env *Environment) getReplacementBinds(ctx context.Context, replacements [][]module.Module) (func(), map[string]string, error) {
    binds := make(map[string]string)
    volumeName := "replacements_volume"

    respCreateVolume, err := env.c.CreateVolume(ctx,
        container.NewCreateVolumeConfig(
            container.WithVolumeName(volumeName),
        ),
    )
    if err != nil {
        return nil, map[string]string{}, err
    }

    respUseContainer, err := env.c.UseContainer(ctx,
        container.NewUseContainerConfig(
            container.WithImageName("golang:1.22"),
            container.WithBindMounts(map[string]string{volumeName: "/replacements"}),
        ),
    )
    if err != nil {
        respCreateVolume.Cleanup()
        return nil, map[string]string{}, err
    }
    defer respUseContainer.Cleanup()

    for _, replacementPair := range replacements {
        replacement := replacementPair[1]
        err := env.getModuleInContainer(ctx, respUseContainer, replacement, "/replacements/" + replacement.ModuleName + replacement.ModuleVersion)
        if err != nil {
            respCreateVolume.Cleanup()
            return nil, map[string]string{}, err
        }
    }
    binds[volumeName] = "/replacements"

    return respCreateVolume.Cleanup, binds, nil
}

func (env *Environment) getMainModuleBinds(ctx context.Context, mainModuleHead module.Module) (func(), map[string]string, error) {
    binds := make(map[string]string)
    volumeName := "main_module_volume"
    modulePathHead := "/mainModule/" + mainModuleHead.ModuleName + mainModuleHead.ModuleVersion

    respCreateVolume, err := env.c.CreateVolume(ctx,
        container.NewCreateVolumeConfig(
            container.WithVolumeName(volumeName),
        ),
    )
    if err != nil {
        return nil, map[string]string{}, err
    }

    respUseContainer, err := env.c.UseContainer(ctx,
        container.NewUseContainerConfig(
            container.WithImageName("golang:1.22"),
            container.WithBindMounts(map[string]string{volumeName: "/mainModule"}),
        ),
    )
    if err != nil {
        respCreateVolume.Cleanup()
        return nil, map[string]string{}, err
    }
    defer respUseContainer.Cleanup()

    err = env.getModuleInContainer(ctx, respUseContainer, mainModuleHead, modulePathHead)
    if err != nil {
        respCreateVolume.Cleanup()
        return nil, map[string]string{}, err
    }
    binds[volumeName] = "/mainModule"

    return respCreateVolume.Cleanup, binds, nil
}

func (env *Environment) getModuleInContainer(ctx context.Context, respUseContainer container.UseContainerResponse, module module.Module, modulePath string) error {
    if module.IsRemotePath() {
        _, err := env.c.ExecuteCommand(ctx,
            container.NewExecuteCommandConfig(
                container.WithContainerID(respUseContainer.ContainerID),
                container.WithCommand([]string{"mkdir", "-p", modulePath}),
            ),
        )
        if err != nil {
            return err
        }

        err = commands.GetModuleFromProxy(ctx, env.c, respUseContainer, module, modulePath)
        if err != nil {
            return err
        }
    } else {
        absPath, err := filepath.Abs(module.ModulePath)
        if err != nil {
            return err
        }

        err = env.c.CopyToContainer(ctx, respUseContainer.ContainerID, map[string]string{absPath:modulePath})
        if err != nil {
            return err
        }
    }
    return nil
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