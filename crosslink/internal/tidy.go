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
	"bufio"
	"fmt"
	"os"
	"slices"
	"strings"

	"go.uber.org/zap"
	"golang.org/x/mod/modfile"
)

type graphNode struct {
	file    *modfile.File
	path    string
	name    string
	deps    []string
	index   int
	sccRoot int
	onStack bool
}

func Tidy(rc RunConfig, outputPath string) error {
	rc.Logger.Debug("crosslink run config", zap.Any("run_config", rc))

	rootModule, err := identifyRootModule(rc.RootPath)
	if err != nil {
		return fmt.Errorf("failed to identify root module: %w", err)
	}

	// Read circular dependency allowlist

	var allowCircular []string
	if rc.AllowCircular != "" {
		var file *os.File
		file, err = os.Open(rc.AllowCircular)
		if err != nil {
			return fmt.Errorf("failed to open cicular dependency allowlist: %w", err)
		}
		defer file.Close()
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := scanner.Text()
			if line != "" {
				allowCircular = append(allowCircular, line)
			}
		}
		if err = scanner.Err(); err != nil {
			return fmt.Errorf("failed to read cicular dependency allowlist: %w", err)
		}
	}

	// Read intra-repository dependency graph

	graph := make(map[string]*graphNode)
	var modsAlpha []string
	err = forGoModFiles(rc, func(filePath string, name string, file *modfile.File) error {
		if !strings.HasPrefix(name, rootModule) {
			rc.Logger.Debug("ignoring module outside root module namespace", zap.String("mod_name", name))
			return nil
		}
		modsAlpha = append(modsAlpha, name)
		modPath, _ := strings.CutSuffix(filePath, "/go.mod")
		graph[name] = &graphNode{
			file:    file,
			path:    modPath,
			name:    name,
			deps:    nil,
			index:   -1,
			sccRoot: -1,
			onStack: false,
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("failed during file walk: %w", err)
	}

	for _, deps := range graph {
		for _, req := range deps.file.Require {
			if _, ok := graph[req.Mod.Path]; ok {
				deps.deps = append(deps.deps, req.Mod.Path)
			}
		}
	}

	rc.Logger.Debug("read module graph", zap.Int("mod_cnt", len(graph)))

	// Compute tidying schedule
	// We use Tarjan's algorithm to identify the topological order of strongly-
	// connected components of the graph, then apply a naive solution to each.

	var modsTopo []string
	nextIdx := 0
	unauthorizedRec := false
	var stack []*graphNode

	var visit func(mod *graphNode)
	visit = func(mod *graphNode) {
		mod.index = nextIdx
		mod.sccRoot = nextIdx
		nextIdx++
		stack = append(stack, mod)
		mod.onStack = true

		for _, mod2Name := range mod.deps {
			mod2 := graph[mod2Name]
			if mod2.index == -1 {
				visit(mod2)
			} else if !mod2.onStack {
				continue
			}
			mod.sccRoot = min(mod.sccRoot, mod2.sccRoot)
		}

		if mod.index == mod.sccRoot {
			var scc []string
			for {
				mod2 := stack[len(stack)-1]
				stack = stack[:len(stack)-1]
				mod2.onStack = false
				scc = append(scc, mod2.path)
				if mod2 == mod {
					break
				}
			}
			if len(scc) > 1 { // circular dependencies
				rc.Logger.Debug("found SCC in module graph", zap.Any("scc", scc))
				for _, mod2 := range scc {
					if !slices.Contains(allowCircular, mod2) {
						fmt.Printf("module depends on itself: %s\n", mod2)
						unauthorizedRec = true
					}
				}
			}

			// Apply a naive solution for each SCC
			// (quadratic in the number of modules, but optimal for 1 or 2)
			for i := 0; i < len(scc)-1; i++ {
				modsTopo = append(modsTopo, scc...)
			}
			modsTopo = append(modsTopo, scc[0])
		}
	}
	for _, modName := range modsAlpha {
		visit(graph[modName])
	}

	rc.Logger.Debug("computed tidy schedule", zap.Int("schedule_len", len(modsTopo)))

	if unauthorizedRec {
		return fmt.Errorf("circular dependencies were found that are not allowlisted")
	}

	// Writing out schedule
	err = os.WriteFile(outputPath, []byte(strings.Join(modsTopo, "\n")), 0600)
	if err != nil {
		return fmt.Errorf("failed to write tidy schedule file: %w", err)
	}

	if rc.Validate {
		// Check validity of solution
		// (iterate over possible paths and check they are all subsequences of modsTopo)

		var queue [][]string
		for _, modName := range modsAlpha {
			queue = append(queue, []string{modName})
		}
		for len(queue) > 0 {
			path := queue[0]
			queue = queue[1:]
			i := 0
			for _, modName := range path {
				i = slices.Index(path[i:], modName)
				if i == -1 {
					return fmt.Errorf("tidy schedule is invalid; changes may not be propagated along path: %v", path)
				}
			}
			for _, dep := range graph[path[0]].deps {
				if !slices.Contains(path, dep) {
					path2 := slices.Clone(path)
					queue = append(queue, slices.Insert(path2, 0, dep))
				}
			}
		}
	}

	return nil
}
