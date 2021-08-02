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

package commontest

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

// WriteGoModFiles is a helper function to dynamically write go.mod files used for testing.
// This func is duplicated from the commontest package to avoid a cyclic dependency.
func WriteGoModFiles(modFiles map[string][]byte) error {
	perm := os.FileMode(0700)

	for modFilePath, file := range modFiles {
		path := filepath.Dir(modFilePath)
		err := os.MkdirAll(path, perm)
		if err != nil {
			return fmt.Errorf("error calling os.MkdirAll(%v, %v): %v", path, perm, err)
		}

		if err := ioutil.WriteFile(modFilePath, file, perm); err != nil {
			return fmt.Errorf("could not write temporary mod file %v", err)
		}
	}

	return nil
}

func RemoveAll(t *testing.T, dir string) {
	err := os.RemoveAll(dir)
	if err != nil {
		t.Fatalf("error removing dir %v: %v", dir, err)
	}
}
