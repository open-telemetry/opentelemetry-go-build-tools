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

	"go.opentelemetry.io/build-tools/internal/repo"
)

var (
	versioningFile string
)

const (
	defaultVersionsConfigName = "versions"
	defaultVersionsConfigType = "yaml"
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

	repoRoot, err := repo.FindRoot()
	if err != nil {
		log.Fatalf("could not find repo root: %v", err)
	}

	versioningFileDefault := filepath.Join(repoRoot,
		fmt.Sprintf("%v.%v", defaultVersionsConfigName, defaultVersionsConfigType))
	rootCmd.PersistentFlags().StringVarP(&versioningFile, "versioning-file", "v", versioningFileDefault,
		"Path to versioning file that contains definitions of all module sets. "+
			"If unspecified, defaults to versions.yaml in the Git repo root.")
}
