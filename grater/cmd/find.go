// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0
package cmd

import (
	"github.com/spf13/cobra"
	"go.opentelemetry.io/build-tools/grater/internal/findhelper"
	"go.opentelemetry.io/build-tools/grater/internal/workspace"
)

// findCmd represents the find command
func findCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "find [module@version]",
		Short: "Finds dependents of a module from pkg.go.dev.",
		Long:  "Finds all dependents of a given module from pkg.go.dev and adds them to the workspace.",
		Example: `
grater find go.opentelemetry.io/otel@v1.24.0
`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ws, err := workspace.NewWorkspace()
			if err != nil {
				return err
			}

			if err = findhelper.FindDependents(ws, args[0]); err != nil {
				return err
			}

			cmd.Printf("Successfully found and added dependents.\n")
			return nil
		},
	}
}