// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

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
	require.ErrorContains(t, err, "[internal\\altpkg\\receiver\\altreceiver] no function matching configuration found\n[internal\\altpkg\\receiver\\badreceiver] no function matching configuration found")
}

func TestUnkeyedPkg(t *testing.T) {
	t.Chdir(filepath.Join("internal", "unkeyedpkg"))
	err := run(".", filepath.Join("..", "..", "config.yaml"))
	require.EqualError(t, err, `[receiver\unkeyedreceiver] struct "UnkeyedConfig" does not prevent unkeyed literal initialization`)
}
