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
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"go.uber.org/zap"
	"golang.org/x/mod/modfile"
)

// Work is the main entry point for the work subcommand.
func Work(rc RunConfig) error {
	rc.Logger.Debug("Crosslink run config", zap.Any("run_config", rc))

	uses, err := intraRepoUses(rc.RootPath)
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

func intraRepoUses(rootPath string) ([]string, error) {
	var uses []string
	err := fs.WalkDir(os.DirFS(rootPath), ".", func(path string, dir fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// find Go module
		if dir.Name() != "go.mod" {
			return nil
		}

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

// pruneUses removes any missing intra-repository use statements.
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
