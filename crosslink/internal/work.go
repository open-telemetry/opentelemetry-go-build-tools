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

package crosslink

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"go.uber.org/zap"
	"golang.org/x/mod/modfile"
)

// Work is the main entry point for the work subcommand.
func Work(rc RunConfig) error {
	rc.Logger.Debug("Crosslink run config", zap.Any("run_config", rc))

	if err := validateGoVersion(rc.GoVersion); err != nil {
		return err
	}

	uses, err := intraRepoUses(rc)
	if err != nil {
		return fmt.Errorf("failed to find Go modules: %w", err)
	}

	goWork, err := openGoWork(rc)
	if errors.Is(err, os.ErrNotExist) {
		goWork = &modfile.WorkFile{
			Syntax: &modfile.FileSyntax{},
		}
		if addErr := goWork.AddGoStmt(rc.GoVersion); addErr != nil {
			return fmt.Errorf("failed to create go.work: %w", addErr)
		}
	} else if err != nil {
		return err
	}

	insertUses(goWork, uses, rc)
	pruneUses(goWork, uses, rc)

	return writeGoWork(goWork, rc)
}

// validateGoVersion checks if goVersion is valid Go release version
// according to the go.work file syntax:
// a positive integer followed by a dot and a non-negative integer
// (for example, 1.19, 1.20).
// More: https://go.dev/ref/mod#workspaces.
func validateGoVersion(goVersion string) error {
	matched, err := regexp.MatchString(`^[1-9]+[0-9]*\.(0|[1-9]+[0-9]*)$`, goVersion)
	if err != nil {
		return err
	}
	if !matched {
		return fmt.Errorf("%q is not a valid Go version", goVersion)
	}
	return nil
}

func intraRepoUses(rc RunConfig) ([]string, error) {
	var uses []string
	err := forGoModules(rc.Logger, rc.RootPath, func(path string) error {
		// normalize use statement (path)
		use := filepath.Dir(path)
		if use == "." {
			use = "./"
		} else {
			use = "./" + use
		}

		uses = append(uses, use)
		return nil
	})
	return uses, err
}

func openGoWork(rc RunConfig) (*modfile.WorkFile, error) {
	goWorkPath := filepath.Join(rc.RootPath, "go.work")
	content, err := os.ReadFile(filepath.Clean(goWorkPath))
	if err != nil {
		return nil, err
	}
	return modfile.ParseWork(goWorkPath, content, nil)
}

func writeGoWork(goWork *modfile.WorkFile, rc RunConfig) error {
	goWorkPath := filepath.Join(rc.RootPath, "go.work")
	content := modfile.Format(goWork.Syntax)
	return os.WriteFile(goWorkPath, content, 0600)
}

// insertUses adds any missing intra-repository use statements.
func insertUses(goWork *modfile.WorkFile, uses []string, rc RunConfig) {
	existingGoWorkUses := make(map[string]bool, len(goWork.Use))
	for _, use := range goWork.Use {
		existingGoWorkUses[use.Path] = true
	}

	for _, useToAdd := range uses {
		if existingGoWorkUses[useToAdd] {
			continue
		}
		err := goWork.AddUse(useToAdd, "")
		if err != nil {
			rc.Logger.Error("Failed to add use statement", zap.Error(err),
				zap.String("path", useToAdd))
		}
	}
}

// pruneUses removes any extraneous intra-repository use statements.
func pruneUses(goWork *modfile.WorkFile, uses []string, rc RunConfig) {
	requiredUses := make(map[string]bool, len(uses))
	for _, use := range uses {
		requiredUses[use] = true
	}

	usesToKeep := make(map[string]bool, len(goWork.Use))
	for _, use := range goWork.Use {
		usesToKeep[use.Path] = true
	}

	for use := range usesToKeep {
		// check to see if its intra dependency
		if !strings.HasPrefix(use, "./") {
			continue
		}

		// check if the intra dependency is still used
		if requiredUses[use] {
			continue
		}

		usesToKeep[use] = false
	}

	// remove unnecessary uses
	for use, needed := range usesToKeep {
		if needed {
			continue
		}

		err := goWork.DropUse(use)
		if err != nil {
			rc.Logger.Error("Failed to drop use statement", zap.Error(err),
				zap.String("path", use))
		}
	}
}
