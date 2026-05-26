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

// GetModuleFromProxy downloads a module via the Go module proxy to the given path.
func GetModuleFromProxy(ctx context.Context, c container.Container, useContainerResp container.UseContainerResponse, module module.Module, modulePath string) error {
	script := fmt.Sprintf(
		`set -e
	GOMODCACHE=$(mktemp -d)
	GOPATH=$(mktemp -d)

	OUT=$(GOMODCACHE="$GOMODCACHE" GOPATH="$GOPATH" \
	GONOSUMCHECK="*" GONOSUMDB="*" \
	go mod download -json %s@%s)

	DIR=$(echo "$OUT" | grep '"Dir"' | awk -F'"' '{print $4}')

	cp -r "$DIR"/. .

	rm -rf "$GOMODCACHE" "$GOPATH"`,
		module.ModulePath,
		module.ModuleVersion,
	)

	_, err := c.ExecuteCommand(ctx,
		container.NewExecuteCommandConfig(
			container.WithContainerID(useContainerResp.ContainerID),
			container.WithWorkingDir(modulePath),
			container.WithCommand([]string{"sh", "-c", script}),
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
			container.WithWorkingDir(modulePath),
			container.WithCommand([]string{"go", "mod", "edit", "-replace", replace}),
		),
	)
	if err != nil {
		return err
	}

	_, err = c.ExecuteCommand(ctx,
		container.NewExecuteCommandConfig(
			container.WithContainerID(useContainerResp.ContainerID),
			container.WithWorkingDir(modulePath),
			container.WithCommand([]string{"go", "mod", "tidy"}),
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
			container.WithWorkingDir(modulePath),
			container.WithCommand([]string{"go", "test", "-v", "./..."}),
		),
	)
	if err != nil {
		return container.ExecuteCommandResponse{}, err
	}
	return resp, nil
}
