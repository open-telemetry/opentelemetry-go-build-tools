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
	components := make([]string, 0, len(data.Components)+1)
	components = append(components, "all")
	for _, comp := range data.Components {
		if comp.Status != nil {
			components = append(components, fmt.Sprintf("%s/%s", comp.Status.Class, comp.Type))
		} else {
			components = append(components, comp.Type)
		}
	}
	sort.Strings(components)
	cfg := data.Chloggen
	cfg.Components = components
	return cg.writeComponents(data.RootFolder, cfg)
}
