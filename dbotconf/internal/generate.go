// Copyright The OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package internal

import (
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"
	"golang.org/x/mod/modfile"
	"gopkg.in/yaml.v3"
)

const header = "# File generated by dbotconf; DO NOT EDIT."

// Allow testing override.
var buildConfigFunc = buildConfig

// buildConfig constructs a dependabotConfig for all modules in the repo.
func buildConfig(root string, mods []*modfile.File, dockerFiles []string, pipFiles []string) (*dependabotConfig, error) {
	c := &dependabotConfig{
		Version: version2,
		Updates: []update{
			{
				PackageEcosystem: ghPkgEco,
				Directory:        "/",
				Labels:           actionLabels,
				Schedule:         weeklySchedule,
			},
		},
	}

	for _, d := range dockerFiles {
		local, err := localPath(root, d)
		if err != nil {
			return nil, err
		}

		c.Updates = append(c.Updates, update{
			PackageEcosystem: dockerPkgEco,
			Directory:        local,
			Labels:           dockerLabels,
			Schedule:         weeklySchedule,
		})
	}

	for _, m := range mods {
		local, err := localModPath(root, m)
		if err != nil {
			return nil, err
		}

		c.Updates = append(c.Updates, update{
			PackageEcosystem: gomodPkgEco,
			Directory:        local,
			Labels:           goLabels,
			Schedule:         weeklySchedule,
		})
	}

	for _, p := range pipFiles {
		local, err := localPath(root, p)
		if err != nil {
			return nil, err
		}

		c.Updates = append(c.Updates, update{
			PackageEcosystem: pipPkgEco,
			Directory:        local,
			Labels:           pipLabels,
			Schedule:         weeklySchedule,
		})
	}

	return c, nil
}

var output io.Writer = os.Stdout

// generate outputs a generated dependabot configuration for all Go modules
// contained in the repository.
func generate() error {
	root, mods, err := allModsFunc()
	if err != nil {
		return err
	}

	dockerFiles, err := allDockerFunc(root)
	if err != nil {
		return err
	}

	pipFiles, err := allPipFunc(root)
	if err != nil {
		return err
	}

	c, err := buildConfigFunc(root, mods, dockerFiles, pipFiles)
	if err != nil {
		return err
	}

	fmt.Fprintln(output, header)
	encoder := yaml.NewEncoder(output)
	encoder.SetIndent(2)
	return encoder.Encode(c)
}

func runGenerate(c *cobra.Command, _ []string) {
	if err := generate(); err != nil {
		fmt.Printf("%s: %v", c.CommandPath(), err)
		os.Exit(1)
	}
}
