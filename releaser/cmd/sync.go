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

	"github.com/spf13/cobra"
)

var (
	allModuleSets_sync bool
	moduleSetNames_sync []string
	noCommit_sync bool
	skipMake_sync bool
)

// syncCmd represents the sync command
var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Syncs the versions of a repo's dependencies",
	Long: `Updates version numbers of :
- Checks that the working tree is clean.
- Switches to a new branch called prerelease_<module set name>_<new version>.
- Updates module versions in all go.mod files.
- Attempts to call go mod tidy on the files.
- Adds and commits changes to Git branch`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("sync called")
	},
}

func init() {
	rootCmd.AddCommand(syncCmd)


}
