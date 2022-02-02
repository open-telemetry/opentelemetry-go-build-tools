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

	"golang.org/x/mod/modfile"
)

func Prune(rc runConfig) {
	var err error

	rootModulePath, err := identifyRootModule(rc.RootPath)
	if err != nil {
		panic(fmt.Sprintf("failed to identify root module: %v", err))
	}

	graph, err := buildDepedencyGraph(rc, rootModulePath)
	if err != nil {
		panic(fmt.Sprintf("failed to build dependency graph: %v", err))
	}

	for _, moduleInfo := range graph {
		err = pruneReplace(rootModulePath, &moduleInfo, rc)

		if err != nil {
			panic(fmt.Sprintf("error pruning replace statements: %v", err))
		}

		err = writeModules(moduleInfo)
		if err != nil {
			panic(fmt.Sprintf("error writing go.mod files: %v", err))
		}
	}
	err = rc.logger.Sync()
	if err != nil {
		fmt.Printf("failed to sync logger:  %v", err)
	}
}

func pruneReplace(rootModulePath string, module *moduleInfo, rc runConfig) error {
	mfParsed, err := modfile.Parse("go.mod", module.moduleContents, nil)
	if err != nil {
		return err
	}

	// check to see if its intra dependency and no longer present
	for _, rep := range mfParsed.Replace {
		// skip excluded
		if _, exists := rc.ExcludedPaths[rep.Old.Path]; exists {
			if rc.Verbose {
				rc.logger.Sugar().Infof("Excluded Module %s, ignoring prune", rep.Old.Path)
			}
			continue
		}

		if _, ok := module.requiredReplaceStatements[rep.Old.Path]; strings.Contains(rep.Old.Path, rootModulePath) && !ok {
			if rc.Verbose {
				rc.logger.Sugar().Infof("Pruning replace statement: Module %s: %s => %s", mfParsed.Module.Mod.Path, rep.Old.Path, rep.New.Path)
			}
			err = mfParsed.DropReplace(rep.Old.Path, rep.Old.Version)
			if err != nil {
				rc.logger.Sugar().Errorf("error dropping replace statement: %v", err)
			}

		}
	}
	module.moduleContents, err = mfParsed.Format()
	if err != nil {
		return err
	}

	return nil
}
