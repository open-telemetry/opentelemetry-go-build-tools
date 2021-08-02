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

package tag

import (
	"fmt"
	"strings"

	"github.com/go-git/go-git/v5/plumbing"
)

type errGitTagsNotOnCommit struct {
	commitHash plumbing.Hash
	tagNames   []string
}

func (e *errGitTagsNotOnCommit) Error() string {
	return fmt.Sprintf("some git tags are not on commit %s:\n%s", e.commitHash, strings.Join(e.tagNames, "\n"))
}
