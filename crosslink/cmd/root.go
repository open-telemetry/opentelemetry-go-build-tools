/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	cl "go.opentelemetry.io/build-tools/crosslink/internal"
)

var (
	excludeFlags = make([]string, 0)
	rc           = cl.DefaultRunConfig()
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "crosslink",
	Short: "Automatically insert replace statements for intra-repository dependencies",
	Long: `
	Crosslink is a tool to assist with go.mod file management for repositories containing
	mulitple go modules. Crosslink automatically inserts replace directives into go.mod files
	for all intra-repository dependencies including transitive dependencies.`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
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
	},
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
	rootCmd.PersistentFlags().StringVar(&rc.RootPath, "root", "", "path to root directory of multi module repository")
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
