// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"fmt"

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
grater replace old/module1 new/module1 old/module2 new/module2
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

			if len(args)%2 != 0 {
				return fmt.Errorf("replacements must be provided in pairs")
			}

			var replacements []string

			for i := 0; i < len(args); i += 2 {
				replacements = append(replacements,
					args[i]+" "+args[i+1],
				)
			}

			if len(replacements) > 0 {
				if err = replacehelper.Replace(ws, replacements); err != nil {
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
