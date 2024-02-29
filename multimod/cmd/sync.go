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
	"path/filepath"

	"github.com/spf13/cobra"

	"go.opentelemetry.io/build-tools/multimod/internal/sync"
)

var (
	otherVersioningFile string
	otherRepoRoot       string
	otherVersionCommit  string
	allModuleSetsSync   bool
	moduleSetNamesSync  []string
	skipGoModTidySync   bool
)

// syncCmd represents the sync command
var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Syncs the versions of a repo's dependencies",
	Long: `Updates version numbers of module sets from another repo:
- Checks that the working tree is clean.
- Switches to a new branch called prerelease_<module set name>_<new version>.
- Updates module versions in all go.mod files.
- Attempts to call go mod tidy on the files.
- Adds and commits changes to Git branch`,
	PreRun: func(cmd *cobra.Command, _ []string) {
		if allModuleSetsSync {
			// do not require module set names if operating on all module sets
			if err := cmd.Flags().SetAnnotation(
				"module-set-names",
				cobra.BashCompOneRequiredFlag,
				[]string{"false"},
			); err != nil {
				log.Fatalf("could not set module-set-names flag as not required flag: %v", err)
			}
		}
	},
	Run: func(*cobra.Command, []string) {
		log.Println("Using versioning file", versioningFile)

		if otherVersioningFile == "" {
			otherVersioningFile = filepath.Join(otherRepoRoot,
				fmt.Sprintf("%v.%v", defaultVersionsConfigName, defaultVersionsConfigType))
		}
		sync.Run(versioningFile, otherVersioningFile, otherRepoRoot, moduleSetNamesSync, otherVersionCommit, allModuleSetsSync, skipGoModTidySync)
	},
}

func init() {
	// Plain log output, no timestamps.
	log.SetFlags(0)

	rootCmd.AddCommand(syncCmd)

	syncCmd.Flags().StringVarP(&otherRepoRoot, "other-repo-root", "o", "",
		"File path of other repository root whose modules' versions need to be updated.")
	if err := syncCmd.MarkFlagRequired("other-repo-root"); err != nil {
		log.Fatalf("could not mark other-repo-root flag as required: %v", err)
	}

	syncCmd.Flags().StringVar(&otherVersioningFile, "other-versioning-file", "",
		"Path to other versioning file that contains all module set versions to sync. "+
			"If unspecified, defaults to versions.yaml in the other Git repo root.")

	syncCmd.Flags().StringVar(&otherVersionCommit, "commit-hash", "",
		"Supports specifying to a commit hash or branch name to sync the modules to. "+
			"If unspecified, defaults to version in versions.yaml.")

	syncCmd.Flags().BoolVarP(&allModuleSetsSync, "all-module-sets", "a", false,
		"Specify this flag to update versions of modules in all sets listed in the versioning file.",
	)

	syncCmd.Flags().StringSliceVarP(&moduleSetNamesSync, "module-set-names", "m", nil,
		"Names of module set whose version is being changed. "+
			"Each name be listed in the module set versioning YAML. "+
			"To specify multiple module sets, specify set names as comma-separated values."+
			"For example: --module-set-names=\"mod-set-1,mod-set-2\"",
	)
	if err := syncCmd.MarkFlagRequired("module-set-names"); err != nil {
		log.Fatalf("could not mark module-set-names flag as required: %v", err)
	}

	syncCmd.Flags().BoolVarP(&skipGoModTidySync, "skip-go-mod-tidy", "s", false,
		"Specify this flag to skip invoking `go mod tidy`. "+
			"To be used for debugging purposes. Should not be skipped during actual release.",
	)
}
