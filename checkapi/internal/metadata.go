// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package internal

import (
	"errors"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Status of the component
type Status struct {
	Class string `yaml:"class"`
}

// Metadata of the component
type Metadata struct {
	Status Status `yaml:"status"`
	Config any    `yaml:"config"`
}

// ReadMetadata reads from the metadata.yaml file in the given folder.
func ReadMetadata(folder string) (Metadata, error) {
	if _, err := os.Stat(filepath.Join(folder, "metadata.yaml")); errors.Is(err, os.ErrNotExist) {
		return Metadata{Status: Status{Class: "pkg"}}, nil
	}
	m, err := os.ReadFile(filepath.Join(folder, "metadata.yaml")) // #nosec G304
	if err != nil {
		return Metadata{}, err
	}
	var componentInfo Metadata
	if err = yaml.Unmarshal(m, &componentInfo); err != nil {
		return Metadata{}, err
	}
	return componentInfo, nil
}
