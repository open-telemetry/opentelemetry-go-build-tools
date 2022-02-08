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
	"io/fs"
	"io/ioutil"
	"path/filepath"
	"strings"

	"go.uber.org/zap"
	"golang.org/x/mod/modfile"
)

// Creates a dependency graph for all intra-repository go.mod files. Only adds
// modules that fall under the root module namespace.
func buildDepedencyGraph(rc RunConfig, rootModulePath string) (map[string]moduleInfo, error) {
	moduleMap := make(map[string]moduleInfo)
	goModFunc := func(filePath string, info fs.FileInfo, err error) error {
		if err != nil {
			rc.Logger.Error("File could not be read during filePath.Walk",
				zap.Error(err),
				zap.String("file path", filePath))

			return nil
		}

		if filepath.Base(filePath) == "go.mod" {
			modFile, err := ioutil.ReadFile(filePath)
			if err != nil {
				return fmt.Errorf("failed to read file: %w", err)
			}

			modInfo := newModuleInfo(modFile, filePath)

			moduleMap[modfile.ModulePath(modFile)] = *modInfo
		}
		return nil
	}

	err := filepath.Walk(rc.RootPath, goModFunc)
	if err != nil {
		return nil, fmt.Errorf("failed during file walk: %w", err)
	}

	for _, modInfo := range moduleMap {
		// reqStack contains a list of module paths that are required to have local replace statements
		// reqStack should only contain intra-repository modules
		reqStack := make([]string, 0)
		alreadyInsertedRepSet := make(map[string]struct{})

		// modfile type that we will work with then write to the mod file in the end
		mfParsed, err := modfile.Parse("go.mod", modInfo.moduleContents, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to parse go.mod file: %w", err)
		}

		// populate initial list of requirements
		// Modules should only be queued for replacement if they meet the following criteria
		// 1. They exist within the set of go.mod files discovered during the filepath walk
		//		- This prevents uneccessary or erroneous replace statements from being added.
		//		- Crosslink will not make an assumption that a module exists even though it falls under the module path.
		// 2. They fall under the module path of the root module
		// 3. They are not the same module that we are currently working with.
		for _, req := range mfParsed.Require {
			if _, existsInPath := moduleMap[req.Mod.Path]; strings.Contains(req.Mod.Path, rootModulePath) &&
				req.Mod.Path != mfParsed.Module.Mod.Path && existsInPath {
				reqStack = append(reqStack, req.Mod.Path)
				alreadyInsertedRepSet[req.Mod.Path] = struct{}{}
			}
		}

		// iterate through stack adding replace directives and transitive requirements as needed
		// if the replace directive already exists for the module path then ensure that it is pointing to the correct location
		for len(reqStack) > 0 {
			var reqModule string
			reqModule, reqStack = reqStack[len(reqStack)-1], reqStack[:len(reqStack)-1]
			modInfo.requiredReplaceStatements[reqModule] = struct{}{}

			// now find all transitive dependencies for the current required module. Only add to stack if they
			// have not already been added and they are not the current module we are working in.
			if value, ok := moduleMap[reqModule]; ok {
				m, err := modfile.Parse("go.mod", value.moduleContents, nil)
				if err != nil {
					return nil, fmt.Errorf("failed to parse go.mod file: %w", err)
				}
				for _, transReq := range m.Require {
					_, existsInPath := moduleMap[transReq.Mod.Path]
					_, alreadyInserted := alreadyInsertedRepSet[transReq.Mod.Path]
					if transReq.Mod.Path != mfParsed.Module.Mod.Path &&
						strings.Contains(transReq.Mod.Path, rootModulePath) &&
						!alreadyInserted && existsInPath {
						reqStack = append(reqStack, transReq.Mod.Path)
						alreadyInsertedRepSet[transReq.Mod.Path] = struct{}{}
					}
				}
			}

		}
	}
	return moduleMap, nil
}
