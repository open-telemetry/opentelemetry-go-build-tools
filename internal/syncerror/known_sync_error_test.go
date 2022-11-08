// Copyright The OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//go:build !windows
// +build !windows

package syncerror

import (
	"errors"
	"syscall"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestKnownSyncError(t *testing.T) {
	tests := []struct {
		testName   string
		assertion  assert.BoolAssertionFunc
		errorValue error
	}{
		{
			testName:   "einval",
			assertion:  assert.True,
			errorValue: syscall.EINVAL,
		},
		{
			testName:   "enotsup",
			assertion:  assert.True,
			errorValue: syscall.ENOTSUP,
		},
		{
			testName:   "enotty",
			assertion:  assert.True,
			errorValue: syscall.EBADF,
		},
		{
			testName:   "fake error",
			assertion:  assert.False,
			errorValue: errors.New("invalid error"),
		},
	}
	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			test.assertion(t, KnownSyncError(test.errorValue))
		})
	}
}
