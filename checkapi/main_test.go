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
	err := run(filepath.Join("internal", "unkeyedpkg"), "config.yaml")
	require.EqualError(t, err, `internal/unkeyedpkg/receiver/unkeyedreceiver struct "UnkeyedConfig" does not prevent unkeyed literal initialization`)
}

func TestMissingConfigFile(t *testing.T) {
	err := run(filepath.Join("internal", "unkeyedpkg"), "badconfig.yaml")
	require.EqualError(t, err, `open badconfig.yaml: no such file or directory`)
}

func TestComponentConfig(t *testing.T) {
	err := run(filepath.Join("internal", "config", "receiver", "configreceiver"), filepath.Join("internal", "config", "config.yaml"))
	require.NoError(t, err, "all config structs are valid")
}

func TestComponentConfigBadStruct(t *testing.T) {
	err := run(filepath.Join("internal", "config", "receiver", "badconfigreceiver"), filepath.Join("internal", "config", "config.yaml"))
	require.EqualError(t, err, "[internal/config/receiver/badconfigreceiver] these structs are not part of config and cannot be exported: ExtraStruct")
}
