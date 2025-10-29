// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

//go:build !windows

package main

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestValidPkg(t *testing.T) {
	err := run(filepath.Join("internal", "testpkg"), "config.yaml")
	require.NoError(t, err)
}

func TestBadPkg(t *testing.T) {
	err := run(filepath.Join("internal", "badpkg"), "config.yaml")
	require.ErrorContains(t, err, "no function matching configuration found")
}

func TestAltConfig(t *testing.T) {
	err := run(filepath.Join("internal", "altpkg"), filepath.Join("internal", "altpkg", "config.yaml"))
	require.NoError(t, err)
}

func TestAltConfigWithOriginalConfig(t *testing.T) {
	err := run(filepath.Join("internal", "altpkg"), "config.yaml")
	require.ErrorContains(t, err, "[internal/altpkg/receiver/altreceiver] no function matching configuration found\n[internal/altpkg/receiver/badreceiver] no function matching configuration found")
}

func TestUnkeyedPkg(t *testing.T) {
	t.Chdir(filepath.Join("internal", "unkeyedpkg"))
	err := run(".", filepath.Join("..", "..", "config.yaml"))
	require.EqualError(t, err, `[receiver/unkeyedreceiver] struct "UnkeyedConfig" does not prevent unkeyed literal initialization`)
}

func TestMissingConfigFile(t *testing.T) {
	err := run(filepath.Join("internal", "unkeyedpkg"), "badconfig.yaml")
	require.EqualError(t, err, `open badconfig.yaml: no such file or directory`)
}

func TestComponentConfig(t *testing.T) {
	t.Chdir(filepath.Join("internal", "config", "receiver", "configreceiver"))
	err := run(".", filepath.Join("..", "..", "config.yaml"))
	require.NoError(t, err, "all config structs are valid")
}

func TestComponentCallConfig(t *testing.T) {
	t.Chdir(filepath.Join("internal", "config", "receiver", "configcallreceiver"))
	err := run(".", filepath.Join("..", "..", "config.yaml"))
	require.NoError(t, err, "all config structs are valid")
}

func TestComponentConfigBadStruct(t *testing.T) {
	t.Chdir(filepath.Join("internal", "config", "receiver", "badconfigreceiver"))
	err := run(".", filepath.Join("..", "..", "config.yaml"))
	require.EqualError(t, err, "[.] these structs are not part of config and cannot be exported: ExtraStruct")
}
