// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package varconfigreceiver

import "go.opentelemetry.io/collector/component"

// createDefaultConfig assigns the config to a variable, so the returned identifier
// names the variable rather than the config type.
func createDefaultConfig() component.Config { // nolint:unused // we do need that method for tests
	cfg := &Config{Foo: "foo"}
	return cfg
}

type Config struct {
	Foo string
	// Embedded struct
	Embedded
}

type Embedded struct {
	Bar string
}
