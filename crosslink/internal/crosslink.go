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
	"strings"

	"go.uber.org/zap"
	"golang.org/x/mod/modfile"
)

func Crosslink(rc RunConfig) error {
	var err error

	rc.Logger.Debug("Crosslink run config", zap.Any("run_config", rc))

	rootModulePath, err := identifyRootModule(rc.RootPath)
	if err != nil {
		return fmt.Errorf("failed to identify root module: %w", err)
	}

	graph, err := buildDepedencyGraph(rc, rootModulePath)
	if err != nil {
		return fmt.Errorf("failed to build dependency graph: %w", err)
	}

	// update go.mod files
	for moduleName, moduleInfo := range graph {
		err = insertReplace(moduleInfo, rc)
		logger := rc.Logger.With(zap.String("module", moduleName))
		if err != nil {
			logger.Error("Failed to insert replace statements",
				zap.Error(err))
			continue
		}

		if rc.Prune {
			pruneReplace(rootModulePath, moduleInfo, rc)
		}

		err = writeModule(moduleInfo)
		if err != nil {
			logger.Error("Failed to write module",
				zap.Error(err))
		}
	}

	// update go.work file
	var modules []string
	for module, _ := range graph {
		// skip excluded
		if _, exists := rc.ExcludedPaths[module]; exists {
			rc.Logger.Debug("Excluded Module, ignoring use",
				zap.String("module", module))
			continue
		}

		localPath, err := filepath.Rel(rootModulePath, module)
		if err != nil {
			return fmt.Errorf("failed to retrieve relative path: %w", err)
		}
		if localPath == "." || localPath == ".." {
			localPath += "/"
		} else if !strings.HasPrefix(localPath, "..") {
			localPath = "./" + localPath
		}
		modules = append(modules, localPath)
	}
	return updateGoWork(modules, rc)
}

func insertReplace(module *moduleInfo, rc RunConfig) error {
	// modfile type that we will work with then write to the mod file in the end
	modContents := module.moduleContents

	for reqModule := range module.requiredReplaceStatements {
		// skip excluded
		if _, exists := rc.ExcludedPaths[reqModule]; exists {
			rc.Logger.Debug("Excluded Module, ignoring replace",
				zap.Any("required_module", reqModule))
			continue
		}

		localPath, err := filepath.Rel(modContents.Module.Mod.Path, reqModule)
		if err != nil {
			return fmt.Errorf("failed to retrieve relative path: %w", err)
		}
		if localPath == "." || localPath == ".." {
			localPath += "/"
		} else if !strings.HasPrefix(localPath, "..") {
			localPath = "./" + localPath
		}

		if oldReplace, exists := containsReplace(modContents.Replace, reqModule); exists {
			if rc.Overwrite {
				rc.Logger.Debug("Overwriting Module",
					zap.String("module", modContents.Module.Mod.Path),
					zap.String("old_replace", reqModule+" => "+oldReplace.New.Path),
					zap.String("new_replace", reqModule+" => "+localPath))

				err = modContents.AddReplace(reqModule, "", localPath, "")

				if err != nil {
					rc.Logger.Error("failed to add replace statement", zap.Error(err),
						zap.String("module", modContents.Module.Mod.Path),
						zap.String("old_replace", reqModule+" => "+oldReplace.New.Path),
						zap.String("new_replace", reqModule+" => "+localPath))
				}
			} else {
				rc.Logger.Debug("Replace statement already exists -run with overwrite to update if desired",
					zap.String("module", modContents.Module.Mod.Path),
					zap.String("current_replace", reqModule+" => "+oldReplace.New.Path))
			}
		} else {
			// does not contain a replace statement. Insert it
			rc.Logger.Debug("Inserting Replace Statement",
				zap.String("module", modContents.Module.Mod.Path),
				zap.String("statement", reqModule+" => "+localPath))
			err = modContents.AddReplace(reqModule, "", localPath, "")
			if err != nil {
				rc.Logger.Error("Failed to add replace statement", zap.Error(err),
					zap.String("module", modContents.Module.Mod.Path),
					zap.String("statement", reqModule+" => "+localPath))
			}
		}
	}
	module.moduleContents = modContents

	return nil
}

// Identifies if a replace statement already exists for a given module name
func containsReplace(replaceStatments []*modfile.Replace, modName string) (*modfile.Replace, bool) {
	for _, repStatement := range replaceStatments {
		if repStatement.Old.Path == modName {
			return repStatement, true
		}
	}
	return nil, false
}

func updateGoWork(modules []string, rc RunConfig) error {
	goWorkPath := filepath.Join(rc.RootPath, "go.work")
	content, err := os.ReadFile(filepath.Clean(goWorkPath))
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}

	goWork, err := modfile.ParseWork(goWorkPath, content, nil)
	if err != nil {
		return err
	}

	// add missing uses
	existingGoWorkUses := make(map[string]bool, len(goWork.Use))
	for _, use := range goWork.Use {
		existingGoWorkUses[use.Path] = true
	}

	for _, useToAdd := range modules {
		if existingGoWorkUses[useToAdd] {
			continue
		}
		err := goWork.AddUse(useToAdd, "")
		if err != nil {
			rc.Logger.Error("Failed to add use statement", zap.Error(err),
				zap.String("module", useToAdd))
		}
	}

	content = modfile.Format(goWork.Syntax)
	return os.WriteFile(goWorkPath, content, 0600)
}
