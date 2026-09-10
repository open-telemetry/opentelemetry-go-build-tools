// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package literalreceiver

import (
	"go.opentelemetry.io/collector/component"
)

func createDefaultConfig() component.Config { // nolint:unused // we do need that method for tests
	return &Config{
		BuildInfo: component.BuildInfo{
			Command: "otelcol",
		},
		// A type of the same package that the configuration does not cover.
		Unchecked: component.ID{},
	}
}

type Config struct {
	BuildInfo component.BuildInfo `mapstructure:"build_info"`
	Unchecked component.ID        `mapstructure:"unchecked"`
}
