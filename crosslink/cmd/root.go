/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"os"

	"github.com/spf13/cobra"
	cl "go.opentelemetry.io/build-tools/crosslink/internal"
)

var (
	root string
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "crosslink",
	Short: "Automatically instert replace statements for intra-repository dependencies",
	Long: `
	Crosslink is a tool to assist with go.mod file management for repositories containing
	mulitple go modules. Crosslink automatically inserts replace directives into go.mod files
	for all intra-repository dependencies including transitive dependencies.`,
	Run: func(cmd *cobra.Command, args []string) {
		// probably need to handle errors here if they get funneled up
		cl.Crosslink(root)
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
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	rootCmd.PersistentFlags().StringVar(&root, "root", "", "path to root directory of multi module repository")
}
