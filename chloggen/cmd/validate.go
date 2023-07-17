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
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"go.opentelemetry.io/build-tools/chloggen/internal/chlog"
)

var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validates the files in the changelog directory",
	RunE: func(cmd *cobra.Command, args []string) error {
		return validate(chlogCtx)
	},
}

func validate(ctx chlog.Context) error {
	if _, err := os.Stat(ctx.ChloggenDir); err != nil {
		return err
	}

	entries, err := chlog.ReadEntries(ctx)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		if err = entry.Validate(); err != nil {
			return err
		}
	}
	fmt.Printf("PASS: all files in %s/ are valid\n", ctx.ChloggenDir)
	return nil
}
