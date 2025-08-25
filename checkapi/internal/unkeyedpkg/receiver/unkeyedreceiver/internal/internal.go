// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package internal

// An exported internal struct should not raise any unkeyed literal initialization checks.
type InternalStruct struct {
	Foo string
}
