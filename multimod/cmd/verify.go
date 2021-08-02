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

	"go.opentelemetry.io/build-tools/multimod/internal/verify"
)

const (
	defaultVersionsConfigName = "versions"
	defaultVersionsConfigType = "yaml"
)

// verifyCmd represents the verify command
var verifyCmd = &cobra.Command{
	Use:   "verify",
	Short: "Verifies that the versioning file is valid",
	Long: `verify checks that all modules listed in sets are valid by verifying the following properties:
- All modules are contained in exactly one module set.
- Versions conform to semver semantics.
- No more than one set of modules exists for any non-zero major version.
- Script warns if any stable modules depend on any unstable modules.
`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Using versioning file", versioningFile)

		verify.Run(versioningFile)
	},
}

func init() {
	// Plain log output, no timestamps.
	log.SetFlags(0)

	rootCmd.AddCommand(verifyCmd)
}
