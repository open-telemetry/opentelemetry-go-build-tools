// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package bad2receiver

import "go.opentelemetry.io/collector/receiver"

func NewFactory() receiver.Factory {
	return nil
}

func ThisFuncWillError() string {
	return "foo"
}
