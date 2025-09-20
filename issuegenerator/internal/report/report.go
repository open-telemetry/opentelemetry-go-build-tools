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

// Package report handles [Report] generation via the [Generator].
package report

import "strings"

// Report on failing tests.
type Report struct {
	Module      string
	FailedTests map[string]string
}

// FailedTestsMD returns information about failed tests if available, otherwise
// an empty string.
func (r *Report) FailedTestsMD() string {
	if len(r.FailedTests) == 0 {
		return ""
	}

	var sb strings.Builder
	sb.WriteString("#### Test Failures\n")

	for testName, testError := range r.FailedTests {
		sb.WriteString("-  `" + testName + "`\n```\n" + testError + "\n```\n")
	}

	return sb.String()
}
