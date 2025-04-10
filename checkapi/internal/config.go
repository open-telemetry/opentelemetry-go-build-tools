// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package internal

type Function struct {
	Name        string   `json:"name"`
	Receiver    string   `json:"receiver"`
	ReturnTypes []string `json:"return_types,omitempty"`
	Params      []string `json:"params,omitempty"`
	TypeParams  []string `json:"type_params,omitempty"`
}

type Apistruct struct {
	Name   string   `json:"name"`
	Fields []string `json:"fields"`
}

type API struct {
	Values    []string     `json:"values,omitempty"`
	Structs   []*Apistruct `json:"structs,omitempty"`
	Functions []*Function  `json:"functions,omitempty"`
}

type FunctionDescription struct {
	Classes     []string `yaml:"classes"`
	Name        string   `yaml:"name"`
	Parameters  []string `yaml:"parameters"`
	ReturnTypes []string `yaml:"return_types"`
}

type Config struct {
	IgnoredPaths     []string              `yaml:"ignored_paths"`
	ExcludedFiles    []string              `yaml:"excluded_files"`
	AllowedFunctions []FunctionDescription `yaml:"allowed_functions"`
	IgnoredFunctions []string              `yaml:"ignored_functions"`
	UnkeyedLiteral   UnkeyedLiteral        `yaml:"unkeyed_literal_initialization"`
}

type UnkeyedLiteral struct {
	Enabled bool `yaml:"enabled"`
	Limit   int  `yaml:"limit"`
}
