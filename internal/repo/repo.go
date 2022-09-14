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

// Package tools provides helper functions used in multiple build tools.
package repo

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"golang.org/x/mod/modfile"
)

// FindRoot retrieves the root of the repository containing the current working directory.
// Beginning at the current working directory (dir), the algorithm checks if joining the ".git"
// suffix, such as "dir.get", is a valid file. Otherwise, it will continue checking the dir's
// parent directory until it reaches the repo root or returns an error if it cannot be found.
func FindRoot() (string, error) {
	start, err := os.Getwd()
	if err != nil {
		return "", err
	}

	dir := start
	for {
		_, err := os.Stat(filepath.Join(dir, ".git"))
		if errors.Is(err, os.ErrNotExist) {
			dir = filepath.Dir(dir)
			// From https://golang.org/pkg/path/filepath/#Dir:
			// The returned path does not end in a separator unless it is the root directory.
			if strings.HasSuffix(dir, string(filepath.Separator)) {
				return "", fmt.Errorf("unable to find git repository enclosing working dir %s", start)
			}
			continue
		}

		if err != nil {
			return "", err
		}

		return dir, nil
	}
}

// FindModules returns all Go modules in the file tree rooted at root.
func FindModules(root string) ([]*modfile.File, error) {
	var results []*modfile.File
	err := filepath.Walk(root, func(path string, info os.FileInfo, walkErr error) error {
		if walkErr != nil {
			// Walk failed to walk into this directory. Stop walking and
			// signal this error.
			return walkErr
		}

		if !info.IsDir() {
			return nil
		}

		goMod := filepath.Join(path, "go.mod")
		f, err := os.Open(goMod)
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		if err != nil {
			return err
		}

		var b bytes.Buffer
		_, err = io.Copy(&b, f)
		if err != nil {
			// Best attempt at cleanup.
			_ = f.Close()
			return err
		}
		if err = f.Close(); err != nil {
			return err
		}

		mFile, err := modfile.Parse(goMod, b.Bytes(), nil)
		if err != nil {
			return err
		}
		results = append(results, mFile)
		return nil
	})

	sort.SliceStable(results, func(i, j int) bool {
		return filepath.Dir(results[i].Syntax.Name) < filepath.Dir(results[j].Syntax.Name)
	})

	return results, err
}
