// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package internal

import "go.opentelemetry.io/build-tools/grater/internal/workspace"

// AddDependents writes the list of dependents to the dependents.txt file.
func AddDependents(dependents []string) error {
	ws, err := workspace.NewWorkspace()
	if err != nil {
		return err
	}

	for _, dep := range dependents {
		// TODO: Implement processing for each dependent to be a valid object for runner.

		err := ws.AddDependent(dep)
		if err != nil {
			return err
		}
	}

	return nil
}
