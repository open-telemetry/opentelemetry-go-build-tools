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

	"go.opentelemetry.io/build-tools/releaser/internal/tag"
)

var (
	commitHash          string
	deleteModuleSetTags bool
	moduleSetName       string
)

// tagCmd represents the tag command
var tagCmd = &cobra.Command{
	Use:   "tag",
	Short: "Applies Git tags to specified commit",
	Long: `Tag script to add Git tags to a specified commit hash created by prerelease script:
- Creates new Git tags for all modules being updated.
- If tagging fails in the middle of the script, the recently created tags will be deleted.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Using versioning file", versioningFile)

		tag.Run(versioningFile, moduleSetName, commitHash, deleteModuleSetTags)
	},
}

func init() {
	// Plain log output, no timestamps.
	log.SetFlags(0)

	rootCmd.AddCommand(tagCmd)

	tagCmd.Flags().StringVarP(&commitHash, "commit-hash", "c", "",
		"Git commit hash to tag.",
	)
	if err := tagCmd.MarkFlagRequired("commit-hash"); err != nil {
		log.Fatalf("could not mark commit-hash flag as required: %v", err)
	}

	tagCmd.Flags().BoolVarP(&deleteModuleSetTags, "delete-module-set-tags", "d", false,
		"Specify this flag to delete all module tags associated with the version listed for the module set in the versioning file. Should only be used to undo recent tagging mistakes.",
	)
}
