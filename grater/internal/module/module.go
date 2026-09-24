// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

// Package module represents a Go module.
package module

import (
	"path"
	"regexp"
)

// Module represents a Go module.
type Module struct {
	ModuleName    string `json:"module_name"`
	ModulePath    string `json:"module_path"`
	ModuleVersion string `json:"module_version"`
}

// NewModule creates a new Module.
func NewModule(modulePath string, moduleVersion string) *Module {
	moduleName := path.Base(modulePath)

	return &Module{
		ModuleName:    moduleName,
		ModulePath:    modulePath,
		ModuleVersion: moduleVersion,
	}
}

// IsRemotePath checks if the module is a remote path.
func (m *Module) IsRemotePath() bool {
	cleanPath := path.Clean(m.ModulePath)
	pattern := `^([a-zA-Z0-9-]+(\.[a-zA-Z0-9-]+)+)/([a-zA-Z0-9-_~.]+)+`
	match, _ := regexp.MatchString(pattern, cleanPath)
	return match
}
