// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

//go:build integration
// +build integration

package commands

import (
	"testing"
	"context"
	"fmt"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.opentelemetry.io/build-tools/grater/internal/container"
	"go.opentelemetry.io/build-tools/grater/internal/dockercontainer"
	"go.opentelemetry.io/build-tools/grater/internal/module"
)

func TestGetModuleFromProxy(t *testing.T) {
	ctx := context.Background()
	var c container.Container
	c, err := dockercontainer.NewDockerContainer()
	require.NoError(t, err)

	useContainerResp, err := c.UseContainer(ctx,
		container.NewUseContainerConfig(
			container.WithImageName("golang:1.25"),
		),
	)
	require.NoError(t, err)

	_, err = c.ExecuteCommand(ctx,
		container.NewExecuteCommandConfig(
			container.WithContainerID(useContainerResp.ContainerID),
			container.WithCommand([]string{
				"mkdir", "-p", "/module/",
			}),
		),
	)
	require.NoError(t, err)

	mod := module.NewModule("go.opentelemetry.io/otel", "v1.24.0")

	err = GetModuleFromProxy(ctx, c, useContainerResp, *mod, "/module/")
	require.NoError(t, err)

	resp, err := c.ExecuteCommand(ctx,
		container.NewExecuteCommandConfig(
			container.WithContainerID(useContainerResp.ContainerID),
			container.WithCommand([]string{"cat", "/module/go.mod"}),
		),
	)
	require.NoError(t, err)
	assert.Contains(t, resp.Output, "go.opentelemetry.io/otel")
}

func TestSetReplaceDirective(t *testing.T) {
	ctx := context.Background()

	var c container.Container

	c, err := dockercontainer.NewDockerContainer()
	require.NoError(t, err)

	binds := map[string]string{
		"../testdata/dependent": "/dependent/",
		"../testdata/moduleFail": "/moduleFail/",
	}

	useContainerResp, err := c.UseContainer(ctx,
		container.NewUseContainerConfig(
			container.WithImageName("golang:1.25.0"),
			container.WithHostToContainerPaths(binds),
		),
	)
	require.NoError(t, err)

	oldModule := module.NewModule("go.opentelemetry.io/build-tools/grater/internal/testdata/module", "",)
	newModule := module.NewModule("../moduleFail", "")

	oldRef := oldModule.ModulePath
	if oldModule.ModuleVersion != "" {
		oldRef = fmt.Sprintf("%s@%s", oldModule.ModulePath, oldModule.ModuleVersion)
	}

	newRef := newModule.ModulePath
	if newModule.ModuleVersion != "" {
		newRef = fmt.Sprintf("%s@%s", newModule.ModulePath, newModule.ModuleVersion)
	}

	err = SetReplaceDirective(ctx, c, useContainerResp, oldRef, newRef, "/dependent/")
	require.NoError(t, err)

	resp, err := c.ExecuteCommand(
		ctx,
		container.NewExecuteCommandConfig(
			container.WithContainerID(useContainerResp.ContainerID),
			container.WithCommand([]string{"cat", "/dependent/go.mod"}),
		),
	)

	assert.Contains(t, resp.Output, "replace go.opentelemetry.io/build-tools/grater/internal/testdata/module => ../moduleFail")
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