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

func TestPkgPkg(t *testing.T) {
	err := run(filepath.Join("internal", "pkg"), "config.yaml")
	require.NoError(t, err)
}

func TestPkgPkgNoFunctions(t *testing.T) {
	t.Chdir("internal")
	err := run("pkg", filepath.Join("pkg", "config_no_functions.yaml"))
	require.EqualError(t, err, `[pkg/pkg] no functions must be exported under this module, found "OtherFunc,SomeFunc"`)
}

func TestPkgPkgSpecificFunctionsAllowed(t *testing.T) {
	t.Chdir("internal")
	err := run("pkg", filepath.Join("pkg", "config_allowed.yaml"))
	require.NoError(t, err)
}

func TestPkgPkgOnlyOneAllowedFunctionAllowed(t *testing.T) {
	t.Chdir("internal")
	err := run("pkg", filepath.Join("pkg", "config_only_one_allowed.yaml"))
	require.EqualError(t, err, `[pkg/pkg] these functions should not be exported: "OtherFunc"`)
}

func TestPkgPkgAllFunctionsAllowed(t *testing.T) {
	t.Chdir("internal")
	err := run("pkg", filepath.Join("pkg", "config_any_functions.yaml"))
	require.NoError(t, err)
}

func TestPkgPkgMissingFunction(t *testing.T) {
	t.Chdir("internal")
	err := run("pkg", filepath.Join("pkg", "config_missing_function.yaml"))
	require.EqualError(t, err, `[pkg/pkg] no function matching configuration found
[pkg/pkg] these functions should not be exported: "OtherFunc,SomeFunc"`)
}

func TestPkgPkgWrongReturnType(t *testing.T) {
	t.Chdir("internal")
	err := run("pkg", filepath.Join("pkg", "config_wrong_return_type.yaml"))
	require.EqualError(t, err, `[pkg/pkg] no function matching configuration found
[pkg/pkg] these functions should not be exported: "OtherFunc,SomeFunc"`)
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

func TestDefaultConstructorLiteral(t *testing.T) {
	t.Chdir(filepath.Join("internal", "defaultctorpkg", "receiver", "literalreceiver"))
	err := run(".", filepath.Join("..", "..", "config.yaml"))
	require.EqualError(t, err, "[.] code_test.go:12 builds component.BuildInfo as a struct literal, use component.NewDefaultBuildInfo() instead")
}

func TestDefaultConstructorUsed(t *testing.T) {
	t.Chdir(filepath.Join("internal", "defaultctorpkg", "receiver", "ctorreceiver"))
	err := run(".", filepath.Join("..", "..", "config.yaml"))
	require.NoError(t, err)
}

func TestDefaultConstructorDisabled(t *testing.T) {
	t.Chdir(filepath.Join("internal", "defaultctorpkg", "receiver", "literalreceiver"))
	err := run(".", filepath.Join("..", "..", "config_disabled.yaml"))
	require.NoError(t, err, "the check does not run unless it is enabled")
}

func TestEmbeddedConfigFields(t *testing.T) {
	for _, tt := range []struct{ name, folder, config, expectedErr string }{
		{
			name:   "embedded fields at every depth",
			folder: filepath.Join("receiver", "configcallreceiver"),
			config: "config_embedded.yaml",
			expectedErr: `[.] config struct "Config" must not have embedded fields, found "Embedded,EmbeddedPtr"
[.] config struct "SubConfig2" must not have embedded fields, found "DeepEmbedded"`,
		},
		{
			name:        "embedded type from another package",
			folder:      filepath.Join("receiver", "configreceiver"),
			config:      "config_embedded.yaml",
			expectedErr: `[.] config struct "Config" must not have embedded fields, found "emb.Config,EmbeddedPtr"`,
		},
		{
			name:        "ignored types are allowed",
			folder:      filepath.Join("receiver", "configcallreceiver"),
			config:      "config_embedded_ignored.yaml",
			expectedErr: `[.] config struct "Config" must not have embedded fields, found "Embedded"`,
		},
		{
			// createDefaultConfig here returns a variable rather than a literal.
			name:        "config assigned to a variable",
			folder:      filepath.Join("receiver", "varconfigreceiver"),
			config:      "config_embedded.yaml",
			expectedErr: `[.] config struct "Config" must not have embedded fields, found "Embedded"`,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Chdir(filepath.Join("internal", "config", tt.folder))
			require.EqualError(t, run(".", filepath.Join("..", "..", tt.config)), tt.expectedErr)
		})
	}
}

func TestComponentConfigBadStruct(t *testing.T) {
	t.Chdir(filepath.Join("internal", "config", "receiver", "badconfigreceiver"))
	err := run(".", filepath.Join("..", "..", "config.yaml"))
	require.EqualError(t, err, "[.] these structs are not part of config and cannot be exported: ExtraStruct")
}
