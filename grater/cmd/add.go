// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"github.com/spf13/cobra"

	"go.opentelemetry.io/build-tools/grater/internal/addhelper"
	"go.opentelemetry.io/build-tools/grater/internal/workspace"
)

var path string

// addCmd represents the add command
func addCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add [dependents...]",
		Short: "Adds one or more dependents to be tested.",
		Long:  "Adds one or more dependents to be tested. The dependents can be specified as command line arguments or in a .txt file, or both.",
		Example: `
grater add github.com/foo/bar/v bar/foo/v --file dependents.txt
grater add github.com/foo/bar/v
grater add --file dependents.txt
grater add -f dependents.txt
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ws, err := workspace.NewWorkspace()
			if err != nil {
				return err
			}

			if path != "" {
				if err = addhelper.AddFromFile(ws, path); err != nil {
					return err
				}
			}

			if err = addhelper.Add(ws, args); err != nil {
				return err
			}

			cmd.Printf("Successfully added dependents.\n")
			return nil
		},
	}

	cmd.Flags().StringVarP(&path, "file", "f", "", "path to the dependents file")
	return cmd
}
