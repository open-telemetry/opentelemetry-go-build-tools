// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

// Package commands provides utilities to execute commands.
package commands

import (
	"context"
	"fmt"

	"go.opentelemetry.io/build-tools/grater/internal/container"
	"go.opentelemetry.io/build-tools/grater/internal/module"
)

// ShallowClone shallow clones a module inside a container to the given path.
func ShallowClone(ctx context.Context, c container.Container, useContainerResp container.UseContainerResponse, module module.Module, branch, modulePath string) error {
	args := []string{"git", "clone", "--depth=1"}
	if branch != "" {
		args = append(args, "--branch", branch)
	}
	args = append(args, "https://"+module.ModulePath+".git", modulePath)

	_, err := c.ExecuteCommand(ctx,
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

// CheckoutBranch checks out the provided branch on the given path.
func CheckoutBranch(ctx context.Context, c container.Container, useContainerResp container.UseContainerResponse, branch, modulePath string) error {
	_, err := c.ExecuteCommand(ctx,
		container.NewExecuteCommandConfig(
			container.WithContainerID(useContainerResp.ContainerID),
			container.WithCommand([]string{
				"git", "-C", modulePath, "checkout", branch,
			}),
		),
	)
	if err != nil {
		return err
	}
	return nil
}

// SetReplaceDirective adds a new replace directive in the go.mod file on the given path.
func SetReplaceDirective(ctx context.Context, c container.Container, useContainerResp container.UseContainerResponse, oldRef, newRef, modulePath string) error {
	replace := fmt.Sprintf("%s=%s", oldRef, newRef)

	_, err := c.ExecuteCommand(ctx,
		container.NewExecuteCommandConfig(
			container.WithContainerID(useContainerResp.ContainerID),
			container.WithCommand([]string{
				"sh", "-c", fmt.Sprintf(`cd %s && go mod edit -replace %s`, modulePath, replace),
			}),
		),
	)
	if err != nil {
		return err
	}

	_, err = c.ExecuteCommand(ctx,
		container.NewExecuteCommandConfig(
			container.WithContainerID(useContainerResp.ContainerID),
			container.WithCommand([]string{
				"sh", "-c", fmt.Sprintf(`cd %s && go mod tidy`, modulePath),
			}),
		),
	)
	if err != nil {
		return err
	}

	return nil
}

// RunModuleTest runs the test of a single module and returns an execute command response.
func RunModuleTest(ctx context.Context, c container.Container, useContainerResp container.UseContainerResponse, modulePath string) (container.ExecuteCommandResponse, error) {
	resp, err := c.ExecuteCommand(ctx,
		container.NewExecuteCommandConfig(
			container.WithContainerID(useContainerResp.ContainerID),
			container.WithCommand([]string{
				"sh", "-c", "cd " + modulePath + " && go test -v ./...",
			}),
		),
	)

	if err != nil {
		return container.ExecuteCommandResponse{}, err
	}

	return resp, nil
}
