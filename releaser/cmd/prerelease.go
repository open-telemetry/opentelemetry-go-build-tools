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

	"go.opentelemetry.io/build-tools/releaser/internal/prerelease"
)

var (
	skipMake           bool
)

// prereleaseCmd represents the prerelease command
var prereleaseCmd = &cobra.Command{
	Use:   "prerelease",
	Short: "Prepares files for new version release",
	Long: `Updates version numbers and commits to a new branch for release:
- Checks that Git tags do not already exist for the new module set version.
- Checks that the working tree is clean.
- Switches to a new branch called pre_release_<module set name>_<new version>.
- Updates module versions in all go.mod files.
- 'make lint' and 'make ci' are called
- Adds and commits changes to Git`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Using versioning file", versioningFile)

		prerelease.Run(versioningFile, moduleSetName, skipMake)
	},
}

func init() {
	// Plain log output, no timestamps.
	log.SetFlags(0)

	rootCmd.AddCommand(prereleaseCmd)

	prereleaseCmd.Flags().BoolVarP(&skipMake, "skip-make", "s", false,
		"Specify this flag to skip the 'make lint' and 'make ci' steps. "+
			"To be used for debugging purposes. Should not be skipped during actual release.",
	)
}
