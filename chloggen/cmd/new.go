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
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"go.opentelemetry.io/build-tools/chloggen/internal/chlog"
)

var (
	filename string
)

var newCmd = &cobra.Command{
	Use:   "new",
	Short: "Creates new change file",
	RunE: func(cmd *cobra.Command, args []string) error {
		return initialize(chlog.DefaultCtx, filename)
	},
}

func initialize(ctx chlog.Context, filename string) error {
	path := filepath.Join(ctx.UnreleasedDir, cleanFileName(filename))
	var pathWithExt string
	switch ext := filepath.Ext(path); ext {
	case ".yaml":
		pathWithExt = path
	case ".yml":
		pathWithExt = strings.TrimSuffix(path, ".yml") + ".yaml"
	case "":
		pathWithExt = path + ".yaml"
	default:
		return fmt.Errorf("non-yaml extension: %s", ext)
	}

	templateBytes, err := os.ReadFile(filepath.Clean(ctx.TemplateYAML))
	if err != nil {
		return err
	}
	err = os.WriteFile(pathWithExt, templateBytes, os.FileMode(0755))
	if err != nil {
		return err
	}
	fmt.Printf("Changelog entry template copied to: %s\n", pathWithExt)
	return nil
}

func cleanFileName(filename string) string {
	replace := strings.NewReplacer("/", "_", "\\", "_")
	return replace.Replace(filename)
}

func init() {
	newCmd.Flags().StringVarP(&filename, "filename", "f", "", "name of the file to add")
	if err := newCmd.MarkFlagRequired("filename"); err != nil {
		log.Fatalf("could not mark filename flag as required: %v", err)
	}
}
