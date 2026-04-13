// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

// Package pkg is a test package used by checkapi tests.
package pkg

// SomeStruct is a test struct.
type SomeStruct struct {
	OneField string
}

// SomeFunc is a test function.
func SomeFunc(foo string) bool {
	return foo == "foo"
}

// OtherFunc is a test function.
func OtherFunc(bar string) bool {
	return bar == "bar"
}
