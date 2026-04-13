// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

// Package internal contains test fixtures for the unkeyedpkg receiver.
package internal

// Struct is an exported type that should not raise any unkeyed literal initialization checks.
type Struct struct {
	Foo string
}
