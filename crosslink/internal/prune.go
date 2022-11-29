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
	"strings"

	"go.uber.org/zap"
)

// main entry point for the Prune subcommand.
func Prune(rc RunConfig) error {
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

	for moduleName, moduleInfo := range graph {
		pruneReplace(rootModulePath, moduleInfo, rc)
		logger := rc.Logger.With(zap.String("module", moduleName))

		err = writeModule(moduleInfo)
		if err != nil {
			logger.Error("Failed to write module",
				zap.Error(err))
		}
	}
	return nil
}

// pruneReplace removes any extraneous intra-repository replace statements.
func pruneReplace(rootModulePath string, module *moduleInfo, rc RunConfig) {
	modContents := module.moduleContents

	// check to see if its intra dependency and no longer present
	for _, rep := range modContents.Replace {
		// skip excluded
		if _, exists := rc.ExcludedPaths[rep.Old.Path]; exists {

			rc.Logger.Debug("Excluded Module, ignoring prune", zap.String("excluded_mod", rep.Old.Path))

			continue
		}

		if _, ok := module.requiredReplaceStatements[rep.Old.Path]; strings.Contains(rep.Old.Path, rootModulePath) && !ok {
			if rc.Verbose {
				rc.Logger.Debug("Pruning replace statement",
					zap.String("module", modContents.Module.Mod.Path),
					zap.String("replace_statement", rep.Old.Path+" => "+rep.New.Path))
			}
			err := modContents.DropReplace(rep.Old.Path, rep.Old.Version)
			if err != nil {
				rc.Logger.Error("error dropping replace statement",
					zap.Error(err),
					zap.String("module", modContents.Module.Mod.Path),
					zap.String("replace_statement", rep.Old.Path+" => "+rep.New.Path))
			}

		}
	}
	module.moduleContents = modContents
}
