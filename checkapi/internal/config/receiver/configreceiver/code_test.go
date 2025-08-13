// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package configreceiver

import "go.opentelemetry.io/collector/component"

func createDefaultConfig() component.Config {
	fooStr := "foo"
	return &Config{
		Foo: []string{"foo"},
		Sub: SubConfig{
			Foo: fooStr,
			Bar: false,
			Sub2: SubConfig2{
				FooBar: "foobar",
			},
		},
	}
}

type Config struct {
	Foo     []string
	Bar     map[string]string
	Bool    bool
	StrChan chan string
	Sub     SubConfig
	Ptr     *PtrStruct
	// Embedded struct
	Embedded
	// Embedded struct pointer
	*EmbeddedPtr
	// Generic type
	Holder GenericHolder[GenericType]
	// Map holding types
	MapOfStructs map[Key]Value
}

type Key struct{}

type Value struct{}

type PtrStruct struct {
	Field string
}

type SubConfig struct {
	Foo  string
	Bar  bool
	Sub2 SubConfig2
}

type SubConfig2 struct {
	FooBar string
}

type Embedded struct {
	Foo string
}

type EmbeddedPtr struct {
	Foo string
}

type GenericHolder[T any] struct {
	Value T
}

type GenericType struct {
	Foo string
}
