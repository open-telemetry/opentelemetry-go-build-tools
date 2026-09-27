// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package ctorreceiver

import (
	"go.opentelemetry.io/collector/component"
)

func createDefaultConfig() component.Config { // nolint:unused // we do need that method for tests
	buildInfo := component.NewDefaultBuildInfo()
	buildInfo.Command = "otelcol"
	return &Config{
		BuildInfo: buildInfo,
	}
}

type Config struct {
	BuildInfo component.BuildInfo `mapstructure:"build_info"`
}
