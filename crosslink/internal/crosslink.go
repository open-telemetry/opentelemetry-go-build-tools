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
	"fmt"
	"path/filepath"
	"strings"

	"go.uber.org/zap"
	"golang.org/x/mod/modfile"
)

func Crosslink(rc RunConfig) {
	var err error

	rootModulePath, err := identifyRootModule(rc.RootPath)
	if err != nil {
		rc.Logger.Sugar().Panic("Failed to identify root Module",
			zap.Error(err),
			zap.Any("run config", rc))
	}

	graph, err := buildDepedencyGraph(rc, rootModulePath)
	if err != nil {
		rc.Logger.Sugar().Panic("Failed to build dependency graph",
			zap.Any("Run Config", rc),
			zap.String("Root Module Path", rootModulePath))
	}

	for moduleName, moduleInfo := range graph {
		err = insertReplace(&moduleInfo, rc)
		if err != nil {
			rc.Logger.Sugar().Error("Failed to insert replace statements",
				zap.Error(err),
				zap.String("Module Name", moduleName),
				zap.Any("Module Info", moduleInfo),
				zap.Any("Run config", rc))
			continue
		}

		if rc.Prune {
			err = pruneReplace(rootModulePath, &moduleInfo, rc)

			if err != nil {
				rc.Logger.Sugar().Error("Failed to prune replace statements",
					zap.Error(err),
					zap.String("Module Name", moduleName),
					zap.Any("Module Info", moduleInfo),
					zap.Any("Run config", rc))
				continue
			}

		}

		err = writeModule(moduleInfo)
		if err != nil {
			rc.Logger.Sugar().Error("Failed to write module",
				zap.Error(err),
				zap.String("Module Name", moduleName),
				zap.Any("Module Info", moduleInfo),
				zap.Any("Run config", rc))
		}
	}
	err = rc.Logger.Sync()
	if err != nil {
		fmt.Printf("failed to sync logger:  %v \n", err)
	}
}

func insertReplace(module *moduleInfo, rc RunConfig) error {
	// modfile type that we will work with then write to the mod file in the end
	mfParsed, err := modfile.Parse("gomod", module.moduleContents, nil)
	if err != nil {
		return fmt.Errorf("failed to parse go.mod file: %w", err)
	}

	for reqModule := range module.requiredReplaceStatements {
		// skip excluded
		if _, exists := rc.ExcludedPaths[reqModule]; exists {
			rc.Logger.Sugar().Debug("Excluded Module, ignoring replace",
				zap.Any("Required Module", reqModule))
			continue
		}

		localPath, err := filepath.Rel(mfParsed.Module.Mod.Path, reqModule)
		if err != nil {
			return fmt.Errorf("failed to retrieve relative path: %w", err)
		}
		if localPath == "." || localPath == ".." {
			localPath += "/"
		} else if !strings.HasPrefix(localPath, "..") {
			localPath = "./" + localPath
		}

		if oldReplace, exists := containsReplace(mfParsed.Replace, reqModule); exists {
			if rc.Overwrite {
				rc.Logger.Sugar().Debug("Overwriting Module",
					zap.String("Module", mfParsed.Module.Mod.Path),
					zap.String("Old Replace", reqModule+" => "+oldReplace.New.Path),
					zap.String("New Replace", reqModule+" => "+localPath))

				err = mfParsed.AddReplace(reqModule, "", localPath, "")

				if err != nil {
					rc.Logger.Sugar().Error("failed to add replace statement", zap.Error(err),
						zap.String("Module", mfParsed.Module.Mod.Path),
						zap.String("Old Replace", reqModule+" => "+oldReplace.New.Path),
						zap.String("New Replace", reqModule+" => "+localPath))
				}
			} else {
				rc.Logger.Sugar().Debug("Replace statement already exists -run with overwrite to update if desired",
					zap.String("Module", mfParsed.Module.Mod.Path),
					zap.String("Current Replace", reqModule+" => "+oldReplace.New.Path))
			}
		} else {
			// does not contain a replace statement. Insert it
			rc.Logger.Sugar().Debug("Inserting Replace Statement",
				zap.String("Module", mfParsed.Module.Mod.Path),
				zap.String("Statement", reqModule+" => "+localPath))
			err = mfParsed.AddReplace(reqModule, "", localPath, "")
			if err != nil {
				rc.Logger.Sugar().Error("Failed to add replace statement", zap.Error(err),
					zap.String("Module", mfParsed.Module.Mod.Path),
					zap.String("Statement", reqModule+" => "+localPath))
			}
		}
	}
	module.moduleContents, err = mfParsed.Format()
	if err != nil {
		return fmt.Errorf("failed to format go.mod file: %w", err)
	}

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
