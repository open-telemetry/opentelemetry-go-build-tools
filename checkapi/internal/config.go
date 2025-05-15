// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package internal

// Function represents a function in the codebase.
type Function struct {
	Name        string   `json:"name"`
	Receiver    string   `json:"receiver"`
	ReturnTypes []string `json:"return_types,omitempty"`
	Params      []string `json:"params,omitempty"`
	TypeParams  []string `json:"type_params,omitempty"`
}

// APIstructField represents a struct field in the codebase.
type APIstructField struct {
	Name string
	Type string
}

// APIstruct represents a struct in the codebase.
type APIstruct struct {
	Name   string           `json:"name"`
	Fields []APIstructField `json:"fields"`
}

// API represents the API of the codebase, including functions and structs.
type API struct {
	Values           []string    `json:"values,omitempty"`
	Structs          []APIstruct `json:"structs,omitempty"`
	Functions        []Function  `json:"functions,omitempty"`
	ConfigStructName string
}

// FunctionDescription represents a function description.
type FunctionDescription struct {
	Classes     []string `yaml:"classes"`
	Name        string   `yaml:"name"`
	Parameters  []string `yaml:"parameters"`
	ReturnTypes []string `yaml:"return_types"`
}

// Config represents the configuration for the codebase analysis.
type Config struct {
	IgnoredPaths     []string              `yaml:"ignored_paths"`
	ExcludedFiles    []string              `yaml:"excluded_files"`
	AllowedFunctions []FunctionDescription `yaml:"allowed_functions"`
	IgnoredFunctions []string              `yaml:"ignored_functions"`
	UnkeyedLiteral   UnkeyedLiteral        `yaml:"unkeyed_literal_initialization"`
	ComponentAPI     bool                  `yaml:"component_api"`
}

// UnkeyedLiteral represents the configuration for unkeyed literal initialization.
type UnkeyedLiteral struct {
	Enabled bool `yaml:"enabled"`
	Limit   int  `yaml:"limit"`
}
