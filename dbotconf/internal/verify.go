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
	errInvalid      = errors.New("invalid dependabot configuration")
	errMissing      = errors.New("missing update check(s)")
	errNotEnoughArg = errors.New("path argument required")
	errTooManyArg   = errors.New("only single path argument allowed")
)

// configuredUpdates returns the set of Go modules dependabot is configured to
// check updates for.
func configuredUpdates(path string) (map[string]struct{}, error) {
	f, err := os.Open(filepath.Clean(path))
	if errors.Is(err, os.ErrNotExist) {
		return nil, fmt.Errorf("dependabot configuration file does not exist: %s", path)
	} else if err != nil {
		return nil, fmt.Errorf("failed to read dependabot configuration file: %s", path)
	}

	var c dependabotConfig
	if err := yaml.NewDecoder(f).Decode(&c); err != nil {
		return nil, fmt.Errorf("%w: %v", errInvalid, err)
	}

	updates := make(map[string]struct{})
	for _, u := range c.Updates {
		if u.PackageEcosystem != gomodPkgEco {
			continue
		}
		updates[u.Directory] = struct{}{}
	}
	return updates, nil
}

// verify ensures dependabot configuration contains a check for all modules.
func verify(args []string) error {
	switch len(args) {
	case 0:
		return errNotEnoughArg
	case 1:
		// Valid case.
	default:
		return fmt.Errorf("%w, received %v", errTooManyArg, args)
	}

	root, mods, err := allModsFunc()
	if err != nil {
		return err
	}

	updates, err := configuredUpdatesFunc(args[0])
	if err != nil {
		return err
	}

	var missing []string
	for _, m := range mods {
		local, err := localPath(root, m)
		if err != nil {
			return err
		}

		if _, ok := updates[local]; !ok {
			missing = append(missing, local)
		}
	}

	if len(missing) > 0 {
		return fmt.Errorf("%w: %s", errMissing, strings.Join(missing, ", "))
	}
	return nil
}

func runVerify(c *cobra.Command, args []string) {
	if err := verify(args); err != nil {
		fmt.Printf("%s: %v", c.CommandPath(), err)
		os.Exit(1)
	}
}
