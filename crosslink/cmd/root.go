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
	"log"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	cl "go.opentelemetry.io/build-tools/crosslink/internal"
	"go.uber.org/zap"
)

var (
	excludeFlags = make([]string, 0)
	rc           = cl.DefaultRunConfig()
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "crosslink",
	Short: "Automatically insert replace statements for intra-repository dependencies",
	Long: `Crosslink is a tool to assist with go.mod file management for repositories containing
	mulitple go modules. Crosslink automatically inserts replace statements into go.mod files
	for all intra-repository dependencies including transitive dependencies so the local module is used.`,
	PersistentPreRun: preRunSetup,
	Run: func(cmd *cobra.Command, args []string) {
		cl.Crosslink(rc)
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&rc.RootPath, "root", "", "path to root directory of multi-module repository")
	rootCmd.PersistentFlags().StringSliceVar(&excludeFlags, "exclude", []string{}, "list of comma separated go modules that crosslink will ignore in operations."+
		"multiple calls of --exclude can be made")
	rootCmd.PersistentFlags().BoolVarP(&rc.Verbose, "verbose", "v", false, "verbose output")
	rootCmd.Flags().BoolVar(&rc.Overwrite, "overwrite", false, "overwrite flag allows crosslink to make destructive (replacing or updating) actions to existing go.mod files")
	rootCmd.Flags().BoolVarP(&rc.Prune, "prune", "p", false, "enables pruning operations on all go.mod files inside root repository")
}

// transform array slice into map
func transformExclude(ef []string) map[string]struct{} {
	output := make(map[string]struct{})
	for _, val := range ef {
		output[val] = struct{}{}
	}
	return output
}

func preRunSetup(cmd *cobra.Command, args []string) {
	rc.ExcludedPaths = transformExclude(excludeFlags)

	// enable verbosity on overwrite if user has not supplied another value
	vExists := false
	cmd.Flags().Visit(func(input *pflag.Flag) {
		if input.Name == "verbose" {
			vExists = true
		}
	})
	if rc.Overwrite && !vExists {
		rc.Verbose = true
	}
	var err error
	if rc.Verbose {
		rc.Logger, err = zap.NewDevelopment()
		if err != nil {
			log.Printf("Could not create zap logger: %v", err)
		}

	}

}
