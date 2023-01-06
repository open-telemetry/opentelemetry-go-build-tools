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

package internal

import (
	"path/filepath"
	"strings"

	"golang.org/x/mod/modfile"

	"go.opentelemetry.io/build-tools/internal/repo"
)

// Allow test overrides and validation.
var (
	allModsFunc           = allMods
	allDockerFunc         = allDocker
	configuredUpdatesFunc = configuredUpdates
)

// allMods returns the repo root and all module files within it.
func allMods() (string, []*modfile.File, error) {
	root, err := repo.FindRoot()
	if err != nil {
		return "", nil, err
	}
	root, err = filepath.Abs(root)
	if err != nil {
		return "", nil, err
	}
	mods, err := repo.FindModules(root)
	if err != nil {
		return "", nil, err
	}
	return root, mods, nil
}

func allDocker(root string) ([]string, error) {
	return repo.FindFilePatternDirs(root, "*Dockerfile*")
}

// localModPath returns the dependabot appropriate directory name for module
// mod that resides in a repo with root.
func localModPath(root string, mod *modfile.File) (string, error) {
	return localPath(root, mod.Syntax.Name)
}

// localPath returns the dependabot appropriate directory name for file at path
// that resides in a repo with root.
func localPath(root, path string) (string, error) {
	absPath, err := filepath.Abs(filepath.Dir(path))
	if err != nil {
		return "", err
	}
	local := strings.TrimPrefix(absPath, root)
	if local == "" {
		local = "/"
	}
	return local, nil
}
