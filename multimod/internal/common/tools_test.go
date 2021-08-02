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
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsStableVersion(t *testing.T) {
	testCases := []struct {
		Version  string
		Expected bool
	}{
		{Version: "v1.0.0", Expected: true},
		{Version: "v1.2.3", Expected: true},
		{Version: "v1.0.0-RC1", Expected: true},
		{Version: "v1.0.0-RC2+MetaData", Expected: true},
		{Version: "v10.10.10", Expected: true},
		{Version: "v0.0.0", Expected: false},
		{Version: "v0.1.2", Expected: false},
		{Version: "v0.20.0", Expected: false},
		{Version: "v0.0.0-RC1", Expected: false},
		{Version: "not-valid-semver", Expected: false},
	}

	for _, tc := range testCases {
		actual := IsStableVersion(tc.Version)

		assert.Equal(t, tc.Expected, actual)
	}
}
