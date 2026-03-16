// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"go.opentelemetry.io/build-tools/grater/internal/workspace"
)

func initCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize .grater/ workspace in the current working directory",
		RunE: func(cmd *cobra.Command, _ []string) error {
			var err error
			graterDir := filepath.Join(".", ".grater")

			err = workspace.GraterExists(".")
			if err == nil {
				return nil // If .grater already exists, skip.
			}

			err = os.Mkdir(graterDir, dirReadWrite)
			if err != nil {
				return fmt.Errorf("failed to create .grater/ directory: %w", err)
			}

			cmd.Printf("Initialized .grater/ workspace at %s\n", graterDir)

			return nil
		},
	}
	return cmd
}
