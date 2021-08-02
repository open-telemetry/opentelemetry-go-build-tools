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

package common

import (
	"fmt"
	"log"
	"os"

	"golang.org/x/mod/semver"

	tools "go.opentelemetry.io/build-tools"
)

// IsStableVersion returns true if modSet.Version is stable (i.e. version major greater than
// or equal to v1), else false.
func IsStableVersion(v string) bool {
	return semver.Compare(semver.Major(v), "v1") >= 0
}

// ChangeToRepoRoot changes to the root of the Git repository the script is called from and returns it as a string.
func ChangeToRepoRoot() (string, error) {
	repoRoot, err := tools.FindRepoRoot()
	if err != nil {
		return "", fmt.Errorf("unable to find repo root: %v", err)
	}

	log.Println("Changing to root directory...")
	err = os.Chdir(repoRoot)
	if err != nil {
		return "", fmt.Errorf("unable to change to repo root: %v", err)
	}

	return repoRoot, nil
}
