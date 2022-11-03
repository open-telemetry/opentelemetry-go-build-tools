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
	"errors"
	"fmt"
	"log"
	"os"
	"syscall"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"go.uber.org/zap"

	cl "go.opentelemetry.io/build-tools/crosslink/internal"
	"go.opentelemetry.io/build-tools/internal/repo"
)

type commandConfig struct {
	runConfig    cl.RunConfig
	excludeFlags []string
	rootCommand  cobra.Command
	pruneCommand cobra.Command
}

func newCommandConfig() *commandConfig {
	c := &commandConfig{
		runConfig: cl.DefaultRunConfig(),
	}

	preRunSetup := func(cmd *cobra.Command, args []string) error {
		c.runConfig.ExcludedPaths = transformExclude(c.excludeFlags)

		if c.runConfig.RootPath == "" {
			rp, err := repo.FindRoot()
			if err != nil {
				return fmt.Errorf("could not find a valid repository: %w", err)
			}
			c.runConfig.RootPath = rp
		}

		// enable verbosity on overwrite if user has not supplied another value
		vExists := false
		cmd.Flags().Visit(func(input *pflag.Flag) {
			if input.Name == "verbose" {
				vExists = true
			}
		})
		if c.runConfig.Overwrite && !vExists {
			c.runConfig.Verbose = true
		}
		var err error
		if c.runConfig.Verbose {
			c.runConfig.Logger, err = zap.NewDevelopment()
			if err != nil {
				return fmt.Errorf("could not create zap logger: %w", err)
			}
		}
		return nil

	}

	postRunSetup := func(cmd *cobra.Command, args []string) error {
		err := c.runConfig.Logger.Sync()
		if err != nil && !errors.Is(err, syscall.ENOTTY) {
			return fmt.Errorf("failed to sync logger: %w \n", err)
		}
		return nil
	}

	c.rootCommand = cobra.Command{
		Use:   "crosslink",
		Short: "Automatically insert replace statements for intra-repository dependencies",
		Long: `Crosslink is a tool to assist with go.mod file management for repositories containing
		multiple go modules. Crosslink automatically inserts replace statements into go.mod files
		for all intra-repository dependencies including transitive dependencies so the local module is used.`,
		PersistentPreRunE:  preRunSetup,
		PersistentPostRunE: postRunSetup,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cl.Crosslink(c.runConfig)
		},
	}

	c.pruneCommand = cobra.Command{
		Use:   "prune",
		Short: "Remove unnecessary replace statements from intra-repository go.mod files",
		Long: `Prune will analyze and remove any uncessary replace statements for intra-repository
		go.mod files that are not direct or transitive dependencies for intra-repository modules. 
		This is a destructive action and will overwrite existing go.mod files. 
		Prune will not remove modules that fall outside of the root module namespace.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cl.Prune(c.runConfig)
		},
	}
	c.rootCommand.AddCommand(&c.pruneCommand)
	return c
}

var (
	comCfg = newCommandConfig()
)

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := comCfg.rootCommand.Execute()
	if err != nil {
		log.Printf("failed execute: %v", err)
		os.Exit(1)
	}
}

func init() {

	comCfg.rootCommand.PersistentFlags().StringVar(&comCfg.runConfig.RootPath, "root", "", `path to root directory of multi-module repository. If --root flag is not provided crosslink will attempt to find a
	git repository in the current or a parent directory.`)
	comCfg.rootCommand.PersistentFlags().StringSliceVar(&comCfg.excludeFlags, "exclude", []string{}, "list of comma separated go modules that crosslink will ignore in operations."+
		"multiple calls of --exclude can be made")
	comCfg.rootCommand.PersistentFlags().BoolVarP(&comCfg.runConfig.Verbose, "verbose", "v", false, "verbose output")
	comCfg.rootCommand.Flags().BoolVar(&comCfg.runConfig.Overwrite, "overwrite", false, "overwrite flag allows crosslink to make destructive (replacing or updating) actions to existing go.mod files")
	comCfg.rootCommand.Flags().BoolVarP(&comCfg.runConfig.Prune, "prune", "p", false, "enables pruning operations on all go.mod files inside root repository")
}

// transform array slice into map
func transformExclude(ef []string) map[string]struct{} {
	output := make(map[string]struct{})
	for _, val := range ef {
		output[val] = struct{}{}
	}
	return output
}
