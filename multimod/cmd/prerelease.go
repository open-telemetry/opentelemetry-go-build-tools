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

package cmd

import (
	"fmt"
	"log"

	"github.com/spf13/cobra"

	"go.opentelemetry.io/build-tools/multimod/internal/prerelease"
)

var (
	allModuleSets bool
	moduleSetNames []string
	skipMake bool
)

// prereleaseCmd represents the prerelease command
var prereleaseCmd = &cobra.Command{
	Use:   "prerelease",
	Short: "Prepares files for new version release",
	Long: `Updates version numbers and commits to a new branch for release:
- Checks that the working tree is clean.
- Checks that Git tags do not already exist for the new module set version.
- Switches to a new branch called prerelease_<module set name>_<new version>.
- Updates version.go files, if they exist.
- Updates module versions in all go.mod files.
- Attempts to call go mod tidy on the files.
- Adds and commits changes to Git branch`,
	PreRun: func(cmd *cobra.Command, args []string) {
		if allModuleSets {
			// do not require commit-hash flag if deleting module set tags
			if err := cmd.Flags().SetAnnotation(
				"module-set-names",
				cobra.BashCompOneRequiredFlag,
				[]string{"false"},
			); err != nil {
				log.Fatalf("could not set module-set-names flag as not required flag: %v", err)
			}
		}
	},
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Using versioning file", versioningFile)

		prerelease.Run(versioningFile, moduleSetNames, allModuleSets, skipMake)
	},
}

func init() {
	// Plain log output, no timestamps.
	log.SetFlags(0)

	rootCmd.AddCommand(prereleaseCmd)

	prereleaseCmd.Flags().BoolVarP(&allModuleSets, "all-module-sets", "a", false,
		"Specify this flag to update versions of modules in all sets listed in the versioning file.",
	)

	prereleaseCmd.Flags().StringSliceVarP(&moduleSetNames, "module-set-names", "m", nil,
		"Names of module set whose version is being changed. " +
			"Each name be listed in the module set versioning YAML. " +
			"To specify multiple module sets, specify set names as comma-separated values." +
			"For example: --module-set-names=\"mod-set-1,mod-set-2\"",
	)
	if err := prereleaseCmd.MarkFlagRequired("module-set-names"); err != nil {
		log.Fatalf("could not mark module-set-names flag as required: %v", err)
	}
	prereleaseCmd.Flags().BoolVarP(&skipMake, "skip-make", "s", false,
		"Specify this flag to skip the 'make lint' and 'make ci' steps. "+
			"To be used for debugging purposes. Should not be skipped during actual release.",
	)
}
