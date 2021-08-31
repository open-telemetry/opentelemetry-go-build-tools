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
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReplaceModVersion(t *testing.T) {
	for _, s := range []struct {
		name     string
		input    []byte
		expected []byte
		err      bool
	}{
		{
			name: "simple",
			input: []byte(`module test
go 1.17

require (
	foo.bar/baz v1.2.3
)
`),
			expected: []byte(`module test
go 1.17

require (
	foo.bar/baz v1.2.4
)
`),
		},
		{
			name: "indirect",
			input: []byte(`module test
go 1.17

require (
	foo.bar/baz v1.2.3 // indirect
)
`),
			expected: []byte(`module test
go 1.17

require (
	foo.bar/baz v1.2.4 // indirect
)
`),
		},
		{
			name: "1.17 style",
			input: []byte(`module test
go 1.17

require (
	bar.baz/quux v0.1.2
)

require (
	foo.bar/baz v1.2.3 // indirect
)
`),
			expected: []byte(`module test
go 1.17

require (
	bar.baz/quux v0.1.2
)

require (
	foo.bar/baz v1.2.4 // indirect
)
`),
		},
	} {
		t.Run(s.name, func(t *testing.T) {
			got, err := replaceModVersion("foo.bar/baz", "v1.2.4", s.input)
			assert.Equal(t, string(s.expected), string(got))
			if s.err {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
