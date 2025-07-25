// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"gopkg.in/yaml.v3"

	"go.opentelemetry.io/build-tools/githubgen/datatype"
)

type distributionsGenerator struct {
	writeDistribution func(rootFolder string, distName string, distData distOutput) error
}

type distOutput struct {
	Name        string              `yaml:"name"`
	URL         string              `yaml:"url"`
	Maintainers []string            `yaml:"maintainers"`
	Components  map[string][]string `yaml:"components"`
}

func (cg *distributionsGenerator) Generate(data datatype.GithubData) error {
	for _, dist := range data.Distributions {
		components := map[string][]string{}
		if dist.None {
			for _, c := range data.Components {
				if len(c.Status.Distributions) == 0 && c.Status.Class != "" && c.Status.Class != "pkg" && c.Type != "" {
					array, ok := components[c.Status.Class]
					if !ok {
						array = []string{}
					}
					components[c.Status.Class] = append(array, c.Type)
				}
			}
		} else {
			for _, c := range data.Components {
				inDistro := false
				for _, componentDistro := range c.Status.Distributions {
					if dist.Name == componentDistro {
						inDistro = true
						break
					}
				}
				if inDistro {
					array, ok := components[c.Status.Class]
					if !ok {
						array = []string{}
					}
					components[c.Status.Class] = append(array, c.Type)
				}
			}
		}
		for _, comps := range components {
			sort.Strings(comps)
		}
		output := distOutput{
			Name:        dist.Name,
			URL:         dist.URL,
			Maintainers: dist.Maintainers,
			Components:  components,
		}
		err := cg.writeDistribution(data.RootFolder, dist.Name, output)
		if err != nil {
			return err
		}
	}
	return nil
}

// WriteDistribution writes the distribution data to a YAML file in the specified root folder.
func WriteDistribution(rootFolder string, distName string, distData distOutput) error {
	b, err := yaml.Marshal(distData)
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(rootFolder, "reports", "distributions", fmt.Sprintf("%s.yaml", distName)), b, 0o600)
}

// WriteChloggenComponents writes the Chloggen configuration to a YAML file in the specified root folder.
func WriteChloggenComponents(rootFolder string, cfg datatype.ChloggenConfig) error {
	b, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(rootFolder, ".chloggen", "config.yaml"), b, 0o600)
}
