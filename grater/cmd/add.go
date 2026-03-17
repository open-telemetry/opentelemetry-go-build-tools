// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"go.opentelemetry.io/build-tools/grater/internal/workspace"
)

// addCmd represents the add command
func addCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "add",
		Short: "Adds a dependent to dependents.txt.",
		Run: func(cmd *cobra.Command, args []string) {
			ws, err := workspace.GetWorkspace()
			if err != nil {
				fmt.Printf("Error getting workspace: %v\n", err)
				return
			}

			err = ws.AddDependent(args[0])
			if err != nil {
				fmt.Printf("Error adding dependent: %v\n", err)
				return
			}

			cmd.Printf("Successfully added dependent: %s\n", args[0])
			return
		},
	}
}
