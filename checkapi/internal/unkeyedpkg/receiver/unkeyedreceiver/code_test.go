// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package unkeyedreceiver

import (
	"context"

	"go.opentelemetry.io/collector/receiver"
)

func NewFactory() receiver.Factory {
	return nil
}

type MyInterface interface {
	Foo() string
}

type UnkeyedConfig struct {
	Foo     []string
	Bar     map[string]string
	Bool    bool
	StrChan chan string
}

type ShutdownFunc func(context.Context) error

type Metadata struct {
	data map[string][]string
}

type FooWithAnonymousField struct {
	Data string
	_    struct{}
}

type EmptyStruct struct{}

type StructWithTooManyFields struct {
	Foo  []string
	Bar  map[string]string
	Foo2 string
	Bar2 string
	Foo3 string
	Bar3 string
	Foo4 string
	Bar4 string
	Foo5 string
	Bar5 string
}

type privateStruct struct {
	Foo string
}
