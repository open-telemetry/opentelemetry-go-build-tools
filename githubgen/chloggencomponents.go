// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"
	"sort"

	"go.opentelemetry.io/build-tools/githubgen/datatype"
)

type chloggenComponentsGenerator struct {
	writeComponents func(rootFolder string, cfg datatype.ChloggenConfig) error
}

func (cg *chloggenComponentsGenerator) Generate(data datatype.GithubData) error {
	components := make([]string, 1, len(data.Components)+1)
	components[0] = "all"
	for _, folder := range data.Folders {
		comp := data.Components[folder]
		if comp.Status != nil && comp.Status.Class != "" && comp.Type != "" {
			components = append(components, fmt.Sprintf("%s/%s", comp.Status.Class, comp.Type))
		} else {
			components = append(components, folder)
		}
	}
	sort.Strings(components)
	cfg := data.Chloggen
	cfg.Components = components
	return cg.writeComponents(data.RootFolder, cfg)
}
