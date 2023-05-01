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
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var (
	errNotEnoughArg = errors.New("path argument required")
)

// configuredUpdates returns the set of Go modules, Dockerfiles, Pip requirements dependabot
// is configured to check updates for.
func configuredUpdates(path string) (mods map[string]struct{}, docker map[string]struct{}, pip map[string]struct{}, err error) {
	f, err := os.Open(filepath.Clean(path))
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil, nil, fmt.Errorf("dependabot configuration file does not exist: %s", path)
	}
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to read dependabot configuration file: %s", path)
	}

	var c dependabotConfig
	if err := yaml.NewDecoder(f).Decode(&c); err != nil {
		return nil, nil, nil, fmt.Errorf("invalid dependabot configuration: %w", err)
	}

	mods = make(map[string]struct{})
	docker = make(map[string]struct{})
	pip = make(map[string]struct{})
	for _, u := range c.Updates {
		if u.PackageEcosystem == dockerPkgEco {
			docker[u.Directory] = struct{}{}
		}
		if u.PackageEcosystem == gomodPkgEco {
			mods[u.Directory] = struct{}{}
		}
		if u.PackageEcosystem == pipPkgEco {
			pip[u.Directory] = struct{}{}
		}
	}
	return mods, docker, pip, nil
}

// verify ensures dependabot configuration contains a check for all modules,
// Dockerfiles, and requirements.txt files.
func verify(args []string) error {
	switch len(args) {
	case 0:
		return errNotEnoughArg
	case 1:
		// Valid case.
	default:
		return fmt.Errorf("only single path argument allowed, received: %v", args)
	}

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

	modUp, dockerUp, pipUp, err := configuredUpdatesFunc(args[0])
	if err != nil {
		return err
	}

	var missingMod []string
	for _, m := range mods {
		local, err := localModPath(root, m)
		if err != nil {
			return err
		}

		if _, ok := modUp[local]; !ok {
			missingMod = append(missingMod, local)
		}
	}
	var missingDocker []string
	for _, d := range dockerFiles {
		local, err := localPath(root, d)
		if err != nil {
			return err
		}

		if _, ok := dockerUp[local]; !ok {
			missingDocker = append(missingDocker, local)
		}
	}
	var missingPip []string
	for _, p := range pipFiles {
		local, err := localPath(root, p)
		if err != nil {
			return err
		}

		if _, ok := pipUp[local]; !ok {
			missingPip = append(missingPip, local)
		}
	}

	if len(missingMod) > 0 || len(missingDocker) > 0 || len(missingPip) > 0 {
		msg := "missing update check(s):"
		if len(missingMod) > 0 {
			msg = fmt.Sprintf("%s\n- Go mod files: %s", msg, strings.Join(missingMod, ", "))
		}
		if len(missingDocker) > 0 {
			msg = fmt.Sprintf("%s\n- Dockerfiles: %s", msg, strings.Join(missingDocker, ", "))
		}
		if len(missingPip) > 0 {
			msg = fmt.Sprintf("%s\n- Pip files: %s", msg, strings.Join(missingPip, ", "))
		}
		msg += "\n"
		return errors.New(msg)
	}
	return nil
}

func runVerify(c *cobra.Command, args []string) {
	if err := verify(args); err != nil {
		fmt.Printf("%s: %v", c.CommandPath(), err)
		os.Exit(1)
	}
}
