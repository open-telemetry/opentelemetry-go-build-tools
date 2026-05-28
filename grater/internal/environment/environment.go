// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package environment

import (
	"context"
	"fmt"
	"path"

	"go.opentelemetry.io/build-tools/grater/internal/container"
	"go.opentelemetry.io/build-tools/grater/internal/module"
)

// Environment struct initialises an environment to run tests.
type Environment struct {
	c container.Container
}

// NewEnvironment creates an instance of an environment.
func NewEnvironment(c container.Container) *Environment {
	return &Environment{c: c}
}

// Report is the report for a test run.
type Report struct {
	ExitCode int
	Out      string
}

// NewReport creates a new Report.
func NewReport(exitCode int, out string) Report {
	return Report{
		ExitCode: exitCode,
		Out:      out,
	}
}

// SetUpMainModuleResponse represents the response for setting up the main module.
type SetUpMainModuleResponse struct {
	Binds   map[string]string
	Cleanup func()
}

// NewSetUpMainModuleResponse creates a new SetUpMainModuleResponse.
func NewSetUpMainModuleResponse(binds map[string]string, cleanup func()) SetUpMainModuleResponse {
	return SetUpMainModuleResponse{
		Binds:   binds,
		Cleanup: cleanup,
	}
}

// RunTestResponse is the response for running a test of a single dependent.
type RunTestResponse struct {
	baseReport Report
	headReport Report
}

// NewRunTestResponse creates a new RunTestResponse.
func NewRunTestResponse(baseReport, headReport Report) RunTestResponse {
	return RunTestResponse{
		baseReport: baseReport,
		headReport: headReport,
	}
}

// RunTestsResponse is the response for running tests.
type RunTestsResponse struct {
	responses []RunTestResponse
}

// NewRunTestsResponse creates a new RunTestsResponse.
func NewRunTestsResponse() RunTestsResponse {
	return RunTestsResponse{
		responses: []RunTestResponse{},
	}
}

// AddResponse adds a response to the RunTestsResponse.
func (resp *RunTestsResponse) AddResponse(response RunTestResponse) {
	resp.responses = append(resp.responses, response)
}

// RunTests executes tests of dependents of the main module.
func (env *Environment) RunTests(ctx context.Context, cfg RunTestsConfig) (RunTestsResponse, error) {
	binds := map[string]string{}

	// TODO: Set up injection binds for the main module.

	respMainModule, err := env.setUpMainModule(ctx, cfg.MainModule(), cfg.BaseRef(), cfg.HeadRef())
	if err != nil {
		return NewRunTestsResponse(), err
	}
	defer respMainModule.Cleanup()
	binds = mergeMaps(binds, respMainModule.Binds)

	runTestsResponse := NewRunTestsResponse()
	for _, dep := range cfg.Dependents() {
		runTestResp, err := env.runTest(ctx, cfg, dep, binds)
		if err != nil {
			return NewRunTestsResponse(), err
		}
		runTestsResponse.AddResponse(runTestResp)
	}

	return runTestsResponse, nil
}

func (env *Environment) runTest(ctx context.Context, cfg RunTestsConfig, dependent module.Module, binds map[string]string) (RunTestResponse, error) {
	respUseContainer, err := env.setUpContainerForTest(ctx, dependent, binds)
	if err != nil {
		return RunTestResponse{}, err
	}
	defer respUseContainer.Cleanup()

	dependentWorkDir := "/dependent/" + dependent.ModuleName
	mainModulePath := "/mainModule/" + cfg.MainModule().ModuleName

	if err := env.checkoutBranch(ctx, respUseContainer, cfg.BaseRef(), mainModulePath); err != nil {
		return RunTestResponse{}, err
	}
	if err := env.inject(ctx, respUseContainer, cfg.MainModule().ModulePath, mainModulePath, dependentWorkDir); err != nil {
		return RunTestResponse{}, err
	}
	baseReport, err := env.getTestReport(ctx, respUseContainer, dependentWorkDir)
	if err != nil {
		return RunTestResponse{}, err
	}

	if err := env.checkoutBranch(ctx, respUseContainer, cfg.HeadRef(), mainModulePath); err != nil {
		return RunTestResponse{}, err
	}
	if err := env.inject(ctx, respUseContainer, cfg.MainModule().ModulePath, mainModulePath, dependentWorkDir); err != nil {
		return RunTestResponse{}, err
	}
	headReport, err := env.getTestReport(ctx, respUseContainer, dependentWorkDir)
	if err != nil {
		return RunTestResponse{}, err
	}

	return NewRunTestResponse(baseReport, headReport), nil
}

func (env *Environment) setUpContainerForTest(ctx context.Context, dependent module.Module, binds map[string]string) (container.UseContainerResponse, error) {
	if dependent.IsRemotePath() {
		respUseContainer, err := env.c.UseContainer(ctx,
			container.NewUseContainerConfig(
				container.WithImageName("golang:1.22"),
				container.WithBindMounts(binds),
			),
		)
		if err != nil {
			return container.UseContainerResponse{}, err
		}
		if err := env.shallowClone(ctx, respUseContainer, dependent, dependent.ModuleVersion, "/dependent"); err != nil {
			respUseContainer.Cleanup()
			return container.UseContainerResponse{}, err
		}
		return respUseContainer, nil
	}

	respUseContainer, err := env.c.UseContainer(ctx,
		container.NewUseContainerConfig(
			container.WithImageName("golang:1.22"),
			container.WithBindMounts(binds),
			container.WithHostToContainerPaths(map[string]string{
				dependent.ModulePath: "/dependent/" + dependent.ModuleName,
			}),
		),
	)
	if err != nil {
		return container.UseContainerResponse{}, err
	}
	return respUseContainer, nil
}

func (env *Environment) setUpMainModule(ctx context.Context, mainModule module.Module, baseRef, headRef string) (SetUpMainModuleResponse, error) {
	binds := make(map[string]string)
	volumeName := "main_module_volume"

	respCreateVolume, err := env.c.CreateVolume(ctx,
		container.NewCreateVolumeConfig(
			container.WithVolumeName(volumeName),
		),
	)
	if err != nil {
		return SetUpMainModuleResponse{}, err
	}

	respUseContainer, err := env.c.UseContainer(ctx,
		container.NewUseContainerConfig(
			container.WithImageName("golang:1.22"),
			container.WithBindMounts(map[string]string{volumeName: "/mainModule"}),
		),
	)
	if err != nil {
		respCreateVolume.Cleanup()
		return SetUpMainModuleResponse{}, err
	}
	defer respUseContainer.Cleanup()

	if mainModule.IsRemotePath() {
		modulePath := path.Join("/mainModule", mainModule.ModuleName, baseRef)
		if err := env.shallowClone(ctx, respUseContainer, mainModule, baseRef, modulePath); err != nil {
			respCreateVolume.Cleanup()
			return SetUpMainModuleResponse{}, err
		}
		modulePath = path.Join("/mainModule", mainModule.ModuleName, headRef)
		if err := env.shallowClone(ctx, respUseContainer, mainModule, headRef, modulePath); err != nil {
			respCreateVolume.Cleanup()
			return SetUpMainModuleResponse{}, err
		}
		binds[volumeName] = "/mainModule"
	} else {
		binds[mainModule.ModulePath] = path.Join("/mainModule", mainModule.ModuleName)
	}

	return NewSetUpMainModuleResponse(binds, respCreateVolume.Cleanup), nil
}

func (env *Environment) inject(ctx context.Context, useContainerResp container.UseContainerResponse, old, new, path string) error {
	replace := fmt.Sprintf("%s=%s", old, new)

	_, err := env.c.ExecuteCommand(ctx,
		container.NewExecuteCommandConfig(
			container.WithContainerID(useContainerResp.ContainerID),
			container.WithCommand([]string{
				"sh", "-c",
				fmt.Sprintf(`cd %s && go mod edit -replace %s`, path, replace),
			}),
		),
	)
	if err != nil {
		return err
	}

	_, err = env.c.ExecuteCommand(ctx,
		container.NewExecuteCommandConfig(
			container.WithContainerID(useContainerResp.ContainerID),
			container.WithCommand([]string{
				"sh", "-c",
				fmt.Sprintf(`cd %s && go mod tidy`, path),
			}),
		),
	)
	if err != nil {
		return err
	}

	return nil
}

func (env *Environment) checkoutBranch(ctx context.Context, useContainerResp container.UseContainerResponse, branch, path string) error {
	_, err := env.c.ExecuteCommand(ctx,
		container.NewExecuteCommandConfig(
			container.WithContainerID(useContainerResp.ContainerID),
			container.WithCommand([]string{
				"git", "-C", path, "checkout", branch,
			}),
		),
	)
	if err != nil {
		return err
	}
	return nil
}

func (env *Environment) shallowClone(ctx context.Context, useContainerResp container.UseContainerResponse, module module.Module, branch, path string) error {
	args := []string{"git", "clone", "--depth=1"}
	if branch != "" {
		args = append(args, "--branch", branch)
	}
	args = append(args, "https://" + module.ModulePath + ".git", path)

	_, err := env.c.ExecuteCommand(ctx,
		container.NewExecuteCommandConfig(
			container.WithContainerID(useContainerResp.ContainerID),
			container.WithCommand(args),
		),
	)
	if err != nil {
		return err
	}
	return nil
}

func (env *Environment) getTestReport(ctx context.Context, useContainerResp container.UseContainerResponse, path string) (Report, error) {
	resp, err := env.c.ExecuteCommand(ctx,
		container.NewExecuteCommandConfig(
			container.WithContainerID(useContainerResp.ContainerID),
			container.WithCommand([]string{
				"sh", "-c", "cd " + path + " && go test -v ./...",
			}),
		),
	)

	if err != nil {
		return Report{}, err
	}

	return NewReport(resp.ExitCode, resp.Output), nil
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
