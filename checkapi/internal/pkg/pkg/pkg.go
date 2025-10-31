// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package pkg

type SomeStruct struct {
	OneField string
}

func SomeFunc(foo string) bool {
	return foo == "foo"
}

func OtherFunc(bar string) bool {
	return bar == "bar"
}
