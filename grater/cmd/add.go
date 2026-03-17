// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"go.opentelemetry.io/build-tools/grater/internal"
)

// addCmd represents the add command
func addCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "add",
		Short: "Adds a dependent to dependents.txt.",
		RunE: func(cmd *cobra.Command, args []string) error {
			err := internal.AddDependents(args)
			if err != nil {
				fmt.Printf("Error adding dependents: %v\n", err)
				return err
			}
			cmd.Printf("Successfully added dependents: %s\n", args)
			return nil
		},
	}
}
