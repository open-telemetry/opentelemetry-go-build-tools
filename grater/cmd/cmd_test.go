// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"bytes"
	"fmt"
	"io"
	"testing"

	"github.com/stretchr/testify/require"
)

func runCobra(t *testing.T, args ...string) (string, error) {
	cmd := rootCmd()

	outBytes := bytes.NewBufferString("")
	cmd.SetOut(outBytes)

	cmd.SetArgs(args)
	err := cmd.Execute()

	out, ioErr := io.ReadAll(outBytes)
	require.NoError(t, ioErr, "read stdout")
	return string(out), err
}
