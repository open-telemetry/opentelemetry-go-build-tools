// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

// Package cmd provides the command line interface for the grater tool.
package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

func rootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "grater",
		Short: "Detects regressions in downstream dependents",
		Long:  `Grater is a tool to detect regressions introduced in our downstream dependents by our changes.`,
	}
	cmd.SetOut(os.Stdout)
	return cmd
}

// Execute executes the root command.
func Execute() {
	cobra.CheckErr(rootCmd().Execute())
}

func init() {
	// TODO: add flags and subcommands.
}
