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

package prerelease

import (
	"fmt"
	"strings"
)

type errGitTagAlreadyExistsSlice struct {
	errors []errGitTagAlreadyExists
}

func (e *errGitTagAlreadyExistsSlice) Error() string {
	var s []string

	for _, err := range e.errors {
		s = append(s, err.Error())
	}

	return strings.Join(s, "\n")
}

type errGitTagAlreadyExists struct {
	gitTag string
}

func (e *errGitTagAlreadyExists) Error() string {
	return fmt.Sprintf("git tag %v already exists", e.gitTag)
}
