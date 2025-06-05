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

package shared

import (
	"fmt"
	"strings"
)

// ErrGitTagsAlreadyExist is an error that indicates that all git tags checked
// already exist.
type ErrGitTagsAlreadyExist struct {
	tagNames []string
}

// Error returns a string representation of the error.
func (e ErrGitTagsAlreadyExist) Error() string {
	return fmt.Sprintf("all git tags checked already exist:\n%s", strings.Join(e.tagNames, "\n"))
}

// ErrInconsistentGitTagsExist is an error that indicates that some but not all
// git tags checked exist.
type ErrInconsistentGitTagsExist struct {
	tagNames []string
}

// Error returns a string representation of the error.
func (e ErrInconsistentGitTagsExist) Error() string {
	return fmt.Sprintf("git tags inconsistent for module set (some but not all tags in module set):\n%s", strings.Join(e.tagNames, "\n"))
}

type errGetWorktreeFailed struct {
	reason error
}

func (e *errGetWorktreeFailed) Error() string {
	return fmt.Sprintf("failed to get worktree: %v", e.reason)
}

type errWorkingTreeNotClean struct{}

func (e *errWorkingTreeNotClean) Error() string {
	return "working tree not clean"
}
