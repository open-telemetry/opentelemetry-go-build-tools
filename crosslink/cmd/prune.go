/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"github.com/spf13/cobra"
	cl "go.opentelemetry.io/build-tools/crosslink/internal"
)

// pruneCmd represents the prune command
var pruneCmd = &cobra.Command{
	Use:   "prune",
	Short: "Remove unnecessary replace statements from intra-repository go.mod files",
	Long: `Prune will analyze and remove any uncessary replace statements that are still
	included in the intra-repository go.mod files. This is a destructive action and will overwrite
	existing go.mod files. Prune will not remove modules that fall outside of the root
	module namespace.`,
	Run: func(cmd *cobra.Command, args []string) {
		cl.Prune(rc)
	},
}

func init() {
	rootCmd.AddCommand(pruneCmd)
}
