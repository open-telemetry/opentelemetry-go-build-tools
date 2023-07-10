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

package chlog

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_SummaryString(t *testing.T) {
	s := summary{
		Version:         "1.0",
		BreakingChanges: []string{"foo", "bar"},
		Deprecations:    []string{"foo", "bar"},
		NewComponents:   []string{"foo", "bar", "new component"},
		Enhancements:    []string{},
		BugFixes:        []string{"foo", "bar", "foobar"},
	}
	result, err := s.String()
	assert.NoError(t, err)
	assert.Equal(t, `
## 1.0

### 🛑 Breaking changes 🛑

foo
bar

### 🚩 Deprecations 🚩

foo
bar

### 🚀 New components 🚀

foo
bar
new component

### 🧰 Bug fixes 🧰

foo
bar
foobar
`, result)
}

func Test_SummaryStringWithLibraryChanges(t *testing.T) {
	s := summary{
		Version:         "1.0",
		BreakingChanges: []string{"foo", "bar"},
		Deprecations:    []string{"foo", "bar"},
		NewComponents:   []string{"foo", "bar", "new component"},
		Enhancements:    []string{},
		BugFixes:        []string{"foo", "bar", "foobar"},
		LibraryChanges:  []string{"foo"},
	}
	result, err := s.String()
	assert.NoError(t, err)
	assert.Equal(t, `
## 1.0

### 🛑 Breaking changes 🛑

foo
bar

### 🚩 Deprecations 🚩

foo
bar

### 🚀 New components 🚀

foo
bar
new component

### 🧰 Bug fixes 🧰

foo
bar
foobar

### 🛠️ Library changes 🛠️

foo
`, result)
}
