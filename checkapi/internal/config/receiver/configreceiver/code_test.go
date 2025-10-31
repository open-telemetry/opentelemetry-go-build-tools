// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package configreceiver

import (
	"github.com/open-telemetry/opentelemetry-go-build-tools/checkapi/internal/config/receiver/configreceiver/emb"
	"go.opentelemetry.io/collector/component"
)

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

type DataDefinedElsewhere string

type Config struct {
	Foo              []string             `mapstructure:"foo"`
	Bar              map[string]string    `mapstructure:"bar"`
	Bool             bool                 `mapstructure:"bool"`
	DefinedElsewhere DataDefinedElsewhere `mapstructure:"data_defined_elsewhere"`
	StrChan          chan string
	Sub              SubConfig  `mapstructure:"subconfig"`
	Ptr              *PtrStruct `mapstructure:"ptrStruct"`
	// Embedded struct
	emb.Config `mapstructure:",squash"`
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
	Foo              string               `mapstructure:"foo"`
	Bar              bool                 `mapstructure:"bar"`
	Sub2             SubConfig2           `mapstructure:"sub_config2"`
	DefinedElsewhere DataDefinedElsewhere `mapstructure:",squash"`
}

type SubConfig2 struct {
	FooBar string `mapstructure:"foobar"`
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
