// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

// Package environment provides utilities for working with environment to run tests.
package environment

import (
	"path"

	"go.opentelemetry.io/build-tools/grater/internal/container"
)

// Environment struct initialises an environment to run tests.
type Environment struct {
	c container.Container
}

// NewEnvironment creates an instance of an environment.
func NewEnvironment(container container.Container) *Environment {
	return &Environment{
		c: container,
	}
}

// RunTests runs the tests in the environment and returns the results.
func (env *Environment) RunTests(cfg *RunTestsConfig) ([][]int, error) {
	var results [][]int

	remotePaths, hostToContainerPaths := []string{}, []string{}

	remotePathMainModule, hostPathMainModule := env.setUpMainModule(cfg.Module())
	remotePathInjections, hostPathInjections := env.setUpInjections(cfg.Injections())

	for _, r := range remotePathInjections + remotePathMainModule {
		remotePaths = append(remotePaths, path.Join("/modules", path.Base(r)))
	}
	for _, h := range hostPathInjections + hostPathMainModule {
		hostToContainerPaths = append(hostToContainerPaths, path.Join("/modules", path.Base(h)))
	}

	volumeName := "remote-modules-storage"
	env.c.CreateVolume(NewCreateVolumeConfig(WithVolumeName(volumeName)))
	err := setUpVolumeForRemoteModules(remotePaths, volumeName)
	if err != nil {
		return nil, err
	}

	for _, dependent := range cfg.Dependents() {
		env.runTest(dependent, volumeName, hostToContainerPaths)
	}
	return results, nil
}

func (env *Environment) setUpMainModule(module module.Module) (string, string) {
	if module.IsRemotePath() {
		return module.ModulePath(), ""
	}
	return "", module.ModulePath()
}

func (env *Environment) setUpInjections(injections []injection.Injection) ([]string, []string) {
	remotePaths, hostPaths := []string{}, []string{}

	for _, module := range injections {
		if module.IsRemotePath() {
			remotePaths = append(remotePaths, module.ModulePath())
		} else {
			hostPaths = append(hostPaths, module.ModulePath())
		}
	}
	return remotePaths, hostPaths
}

func (env *Environment) setUpVolumeForRemoteModules(remotePaths []string{}, volumeName string) error {
	resp, err := env.c.UseContainer(NewUseContainerConfig(WithImageName("alpine")))
	if err != nil {
		return err
	}
	defer resp.Cleanup()

	for _, r := range remotePaths {
		executeCommandResponse, err := env.c.ExecuteCommand(
			NewExecuteCommandConfig(
				WithContainerID(resp.ContainerID()),
				WithCommand([]string{"git", "clone", "--depth", "1", "https://" + r, "/modules/" + path.Base(r)})
			)
		)
		if err != nil {
			return err
		}
	}

	return nil
}

func (env *Environment) runTest(dependent module.Module, volumeName string, hostToContainerPaths []string, injections []module.Module, main module.Module) {
	resp, err := env.c.UseContainer(
		NewUseContainerConfig(
			WithImageName("go:alpine"),
			WithBindMounts(volumeName, "/"),
			WithHostToContainerPaths(hostToContainerPaths),
		)
	)
	if err != nil {
		return
	}
	defer resp.Cleanup()

	executeCommandResponse, err := env.c.ExecuteCommand(
		NewExecuteCommandConfig(
			WithContainerID(resp.ContainerID()),
			WithCommand([]string{"git", "clone", "--depth", "1", "https://" + dependent.ModulePath(), "/modules/" + path.Base(dependent.ModulePath())})
		)
	)
	if err != nil {
		return
	}

	executeCommandResponse, err := env.c.ExecuteCommand(
			NewExecuteCommandConfig(
				WithContainerID(resp.ContainerID()),
				WithCommand([]string{"cd", "/modules/" + path.Base(injection.ModulePath())})
			)
		)
	if err != nil {
		return
	}

	for _, injection := range injections {
		executeCommandResponse, err := env.c.ExecuteCommand(
			NewExecuteCommandConfig(
				WithContainerID(resp.ContainerID()),
				WithCommand(
					[]string{"echo", "replace", "<injection.ModulePath()>"}
				)
			)
		)
		if err != nil {
			return
		}}

	executeCommandResponse, err := env.c.ExecuteCommand(
			NewExecuteCommandConfig(
				WithContainerID(resp.ContainerID()),
				WithCommand(
					[]string{"echo", "replace", "<mainmodule.ModulePath()>"}
				)
			)
		)
	if err != nil {
		return
	}

	executeCommandResponse, err := env.c.ExecuteCommand(
		NewExecuteCommandConfig(
			WithContainerID(resp.ContainerID()),
			WithCommand(
				[]string{"go", "test", "./.."}
			)
		)
	)
	if err != nil {
		return
	}

	executeCommandResponse, err := env.c.ExecuteCommand(
		NewExecuteCommandConfig(
			WithContainerID(resp.ContainerID()),
			WithCommand(
				[]string{"echo", "replace", "<mainmodule.ModulePath()>"}
			)
		)
	)
	if err != nil {
		return}

			executeCommandResponse, err := env.c.ExecuteCommand(
		NewExecuteCommandConfig(
			WithContainerID(resp.ContainerID()),
			WithCommand(
				[]string{"go", "test", "./.."}
			)
		)
	)
	if err != nil {
		return
	}
}