// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package configreceiver

import "go.opentelemetry.io/collector/component"

func createDefaultConfig() component.Config {
	return &Config{
		Foo: []string{"foo"},
		Sub: SubConfig{
			Foo: "foo",
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
}

type SubConfig struct {
	Foo  string
	Bar  bool
	Sub2 SubConfig2
}

type SubConfig2 struct {
	FooBar string
}

type ExtraStruct struct {
	SomeField string
}
