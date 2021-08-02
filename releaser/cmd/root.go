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
	tools "go.opentelemetry.io/build-tools"
)

var (
	moduleSetNames []string
	versioningFile string
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "versions",
	Short: "Enables the release of Go modules with flexible versioning",
	Long: `A Golang release versioning and tagging tool that simplifies and
automates versioning for repos with multiple Go modules.`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	cobra.CheckErr(rootCmd.Execute())
}

func init() {
	cobra.OnInitialize()

	repoRoot, err := tools.FindRepoRoot()
	if err != nil {
		log.Fatalf("could not find repo root: %v", err)
	}
	versioningFile = filepath.Join(repoRoot,
		fmt.Sprintf("%v.%v", defaultVersionsConfigName, defaultVersionsConfigType))

	rootCmd.PersistentFlags().StringVarP(&versioningFile, "versioning-file", "v", versioningFile,
		"Path to versioning file that contains definitions of all module sets. "+
			"If unspecified, defaults to versions.yaml in the Git repo root.")

	rootCmd.PersistentFlags().StringSliceVarP(&moduleSetNames, "module-set-names", "m", nil,
		"Names of module set whose version is being changed. " +
		"Each name be listed in the module set versioning YAML. " +
		"To specify multiple module sets, specify set names as comma-separated values." +
		"For example: --module-set-names=\"mod-set-1,mod-set-2\"",
	)
	if err := rootCmd.MarkPersistentFlagRequired("module-set-names"); err != nil {
		log.Fatalf("could not mark module-set-names flag as required: %v", err)
	}
}
