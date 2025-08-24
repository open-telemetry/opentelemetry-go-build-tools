// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package internal

type Embedded struct {
	Foo                    string `mapstructure:"foo"`
	MySpecialEmbeddedField string `mapstructure:"my_special_embedded_field"`
}
