// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"github.com/spf13/cobra"

	"go.opentelemetry.io/build-tools/grater/internal/replacehelper"
	"go.opentelemetry.io/build-tools/grater/internal/workspace"
)

var replacePath string

// replaceCmd represents the replace command
func replaceCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "replace [replacements...]",
		Short: "Adds one or more replacements to be tested.",
		Long: "Adds one or more replacements to be tested. " +
			"The replacements can be specified as command line arguments or in a .txt file, or both.",
		Example: `
grater replace github.com/foo/bar github.com/foo/bar@v1.0.0
grater replace github.com/foo/bar@v1.0.0 ../local/module
grater replace --file replacements.txt
grater replace -f replacements.txt
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ws, err := workspace.NewWorkspace()
			if err != nil {
				return err
			}

			if replacePath != "" {
				if err = replacehelper.AddFromFile(ws, replacePath); err != nil {
					return err
				}
			}

			if len(args) > 0 {
				if err = replacehelper.Replace(ws, []string{
					args[0] + " " + args[1],
				}); err != nil {
					return err
				}
			}

			cmd.Printf("Successfully added replacements.\n")
			return nil
		},
	}

	cmd.Flags().StringVarP(&replacePath, "file", "f", "", "path to the replacements file")

	return cmd
}