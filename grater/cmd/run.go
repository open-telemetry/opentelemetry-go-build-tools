// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"github.com/spf13/cobra"

	"go.opentelemetry.io/build-tools/grater/internal"
)

// runCmd represents the run command
func runCmd() *cobra.Command {
	var repo string
	var base string
	var head string

	cmd := &cobra.Command{
		Use:   "run",
		Short: "Runs the tests for the added dependents.",
		RunE: func(cmd *cobra.Command, args []string) error {
			var err error
			err = internal.RunTests(repo, base, head)
			return err
		},
	}

	cmd.Flags().StringVarP(&repo, "repo", "r", "", "repository to run tests for")
	cmd.Flags().StringVarP(&base, "base", "B", "", "base branch to compare against")
	cmd.Flags().StringVarP(&head, "head", "H", "", "head branch to compare")
	return cmd
}
