// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package workspace

import "os"

var workspaces = make(map[string]*Workspace)

// GetWorkspace returns a workspace instance for the current directory.
func GetWorkspace() (*Workspace, error) {
	root, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	if ws, ok := workspaces[root]; ok {
		return ws, nil
	}

	ws, err := Init(root)
	if err != nil {
		return nil, err
	}

	workspaces[root] = ws
	return ws, nil
}