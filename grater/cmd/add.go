// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"github.com/spf13/cobra"

	"go.opentelemetry.io/build-tools/grater/internal/add"
	"go.opentelemetry.io/build-tools/grater/internal/workspace"
)

var path string

// addCmd represents the add command
func addCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add",
		Short: "Adds a new dependent to be tested.",
		RunE: func(cmd *cobra.Command, args []string) error {
			ws, err := workspace.NewWorkspace()
			if err != nil {
				return err
			}

			if path != "" {
				if err := add.AddFromFile(ws, path); err != nil {
					return err
				}
			}

			add.Add(ws, args)

			cmd.Printf("Successfully added dependents. \n")
			return nil
		},
	}

	cmd.Flags().StringVarP(&path, "file", "f", "", "path to the dependents file")
	return cmd
}
