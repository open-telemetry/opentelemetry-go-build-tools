// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package cmd

import "os"

const (
	dirReadWrite = os.FileMode(0o755)
	dirReadOnly  = os.FileMode(0o555)
)
