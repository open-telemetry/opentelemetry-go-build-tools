// Copyright The OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"go.opentelemetry.io/build-tools/chloggen/internal/chlog"
)

const (
	insertPoint = "<!-- next version -->\n"
)

var (
	version string
	dry     bool
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Updates CHANGELOG.MD to include all new changes",
	RunE: func(cmd *cobra.Command, args []string) error {
		return update(chlogCtx, version, dry)
	},
}

func update(ctx chlog.Context, version string, dry bool) error {
	entries, err := chlog.ReadEntries(ctx)
	if err != nil {
		return err
	}

	if len(entries) == 0 {
		return fmt.Errorf("no entries to add to the changelog")
	}

	chlogUpdate, err := chlog.GenerateSummary(version, entries)
	if err != nil {
		return err
	}

	if dry {
		fmt.Printf("Generated changelog updates:")
		fmt.Println(chlogUpdate)
		return nil
	}

	oldChlogBytes, err := os.ReadFile(filepath.Clean(ctx.ChangelogMD))
	if err != nil {
		return err
	}
	chlogParts := bytes.Split(oldChlogBytes, []byte(insertPoint))
	if len(chlogParts) != 2 {
		return fmt.Errorf("expected one instance of %s", insertPoint)
	}

	chlogHeader, chlogHistory := string(chlogParts[0]), string(chlogParts[1])

	var chlogBuilder strings.Builder
	chlogBuilder.WriteString(chlogHeader)
	chlogBuilder.WriteString(insertPoint)
	chlogBuilder.WriteString(chlogUpdate)
	chlogBuilder.WriteString(chlogHistory)

	tmpMD := ctx.ChangelogMD + ".tmp"
	if err = os.WriteFile(filepath.Clean(tmpMD), []byte(chlogBuilder.String()), 0600); err != nil {
		return err
	}

	if err = os.Rename(tmpMD, ctx.ChangelogMD); err != nil {
		return err
	}

	fmt.Printf("Finished updating %s\n", ctx.ChangelogMD)

	return chlog.DeleteEntries(ctx)
}

func init() {
	updateCmd.Flags().StringVarP(&version, "version", "v", "vTODO", "will be rendered directly into the update text")
	updateCmd.Flags().BoolVarP(&dry, "dry", "d", false, "will generate the update text and print to stdout")
}
