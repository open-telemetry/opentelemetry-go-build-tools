// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

// Grater is a tool to detect regressions introduced in our downstream dependents by our changes.
package main

import "go.opentelemetry.io/build-tools/grater/cmd"

func main() {
	cmd.Execute()
}
