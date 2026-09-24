// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

// Package emb provides embedded configuration types for testing.
package emb

// Config holds the embedded configuration fields.
type Config struct {
	Foo                    string `mapstructure:"foo"`
	MySpecialEmbeddedField string `mapstructure:"my_special_embedded_field"`
}
