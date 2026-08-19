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
	Internal    bool     `json:"internal"`
}

// APIstructField represents a struct field in the codebase.
type APIstructField struct {
	Name string
	Type string
	Tag  string
}

// APIstruct represents a struct in the codebase.
type APIstruct struct {
	Name     string           `json:"name"`
	Fields   []APIstructField `json:"fields"`
	Internal bool             `json:"internal"`
}

// API represents the API of the codebase, including functions and structs.
type API struct {
	Values           []string    `json:"values,omitempty"`
	Structs          []APIstruct `json:"structs,omitempty"`
	Functions        []Function  `json:"functions,omitempty"`
	ConfigStructName string
	// DefaultConfigLiterals holds the literals built in the default config function.
	DefaultConfigLiterals []TypeLiteral
}

// TypeLiteral is a composite literal of a type belonging to another package.
type TypeLiteral struct {
	// Type is the type as written in the source, for example "configgrpc.ClientConfig".
	Type string
	File string
	Line int
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
	IgnoredPaths       []string              `yaml:"ignored_paths"`
	ExcludedFiles      []string              `yaml:"excluded_files"`
	AllowedFunctions   []FunctionDescription `yaml:"allowed_functions"`
	IgnoredFunctions   []string              `yaml:"ignored_functions"`
	UnkeyedLiteral     UnkeyedLiteral        `yaml:"unkeyed_literal_initialization"`
	ComponentAPI       bool                  `yaml:"component_api"`
	ComponentAPIStrict bool                  `yaml:"component_api_strict"`
	JSONSchema         JSONSchemaConfig      `yaml:"json_schema"`
	EmbeddedConfigFields EmbeddedConfigFields  `yaml:"embedded_config_fields"`
	DefaultCtors       DefaultConstructors   `yaml:"default_constructors"`
}

// EmbeddedConfigFields represents the configuration for the embedded config fields check.
type EmbeddedConfigFields struct {
	Enabled      bool     `yaml:"enabled"`
	IgnoredTypes []string `yaml:"ignored_types"`
}

// DefaultConstructors configures the default constructor check: config structs borrowed
// from other packages must be built with their NewDefault* constructor.
type DefaultConstructors struct {
	Enabled bool `yaml:"enabled"`
	// Types maps a type, as written in the source, to the constructor building it,
	// for example "configgrpc.ClientConfig": "NewDefaultClientConfig".
	Types map[string]string `yaml:"types"`
}

// JSONSchemaConfig represents the configuration of JSON schema validation and mapping
type JSONSchemaConfig struct {
	CheckPresent bool              `yaml:"check_present"`
	CheckValid   bool              `yaml:"check_valid"`
	TypeMappings map[string]string `yaml:"type_mappings"`
}

// UnkeyedLiteral represents the configuration for unkeyed literal initialization.
type UnkeyedLiteral struct {
	Enabled bool `yaml:"enabled"`
	Limit   int  `yaml:"limit"`
}
