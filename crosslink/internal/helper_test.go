// Copyright The OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package crosslink

// Helper functions that are used by multiple test files

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	cp "github.com/otiai10/copy"
)

var (
	testDataDir, _ = filepath.Abs("./test_data")
	mockDataDir, _ = filepath.Abs("./mock_test_data")
)

// the odd naming convention and renaming function is required to avoid dependabot
// failures. If the mock gomod files were named go.mod by default our precommit
// dependabot check would fail. Dependabot does not allow us to ignore directories
// so instead we rename the gomod files to go.mod after directories are copied.
func renameGoMod(fp string) error {
	renameFunc := func(filePath string, info fs.FileInfo, err error) error {
		if err != nil {
			fmt.Printf("Warning: file could not be read during filepath.Walk: %v", err)
			return nil
		}

		if filepath.Base(filePath) == "gomod" {
			dir, _ := filepath.Split(filePath)
			err = os.Rename(filePath, filepath.Join(dir, "go.mod"))
			if err != nil {
				return fmt.Errorf("failed to rename go.mod file: %w", err)
			}
		}
		return nil
	}
	err := filepath.Walk(fp, renameFunc)
	if err != nil {
		return fmt.Errorf("failed during file walk: %w", err)
	}
	return nil
}

// Copies the mocked gomod files into a tempory directory in the test_data folder.
// testName must match the name of a directory within the mock_test_data folder.
func createTempTestDir(testName string) (string, error) {
	tmpRootDir, err := os.MkdirTemp(testDataDir, testName)
	if err != nil {
		return "", fmt.Errorf("failed to make temp director: %w", err)
	}

	mockDir := filepath.Join(mockDataDir, testName)
	err = cp.Copy(mockDir, tmpRootDir)
	if err != nil {
		return "", fmt.Errorf("failed to copy mock data into temp: %w", err)
	}

	return tmpRootDir, nil

}
