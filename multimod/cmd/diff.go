// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"log"
	"strings"

	"github.com/spf13/cobra"

	"go.opentelemetry.io/build-tools/internal/repo"
	"go.opentelemetry.io/build-tools/multimod/internal/diff"
)

var (
	previousVersion string
)

var diffCmd = &cobra.Command{
	Use:   "diff",
	Short: "Determines if any files in a module have changed",
	Run: func(cmd *cobra.Command, args []string) {
		repoRoot, err := repo.FindRoot()
		if err != nil {
			log.Fatalf("could not find repo root: %v", err)
		}

		changed, changedFiles, err := diff.HasChanged(repoRoot, versioningFile, previousVersion, moduleSetName)
		if err != nil {
			log.Fatalf("error running diff: %v", err)
		}

		if changed {
			log.Fatalf("The following files changed in %s modules since %s: \n%s\nRelease is required for %s modset", moduleSetName, previousVersion, strings.Join(changedFiles, "\n"), moduleSetName)
		}
		log.Printf("No %s modules have changed since %s", moduleSetName, previousVersion)
	},
}

func init() {
	// Plain log output, no timestamps.
	log.SetFlags(0)

	rootCmd.AddCommand(diffCmd)
	diffCmd.Flags().StringVarP(&moduleSetName, "module-set-name", "m", "",
		"Name of module set being diff'd. "+
			"Name must be listed in the module set versioning YAML. ",
	)
	if err := diffCmd.MarkFlagRequired("module-set-name"); err != nil {
		log.Fatalf("could not mark module-set-name flag as required: %v", err)
	}

	diffCmd.Flags().StringVarP(&previousVersion, "previous-version", "p", "",
		"Previously released version."+
			"Version must be a tag in the repository. ",
	)
	if err := diffCmd.MarkFlagRequired("previous-version"); err != nil {
		log.Fatalf("could not mark previous-version flag as required: %v", err)
	}
}
