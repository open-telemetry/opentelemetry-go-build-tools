// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package internal

import (
	"fmt"
	"strings"
)

// Function represents a function in the codebase.
type Function struct {
	Name        string   `json:"name"`
	Receiver    string   `json:"receiver"`
	ReturnTypes []string `json:"return_types,omitempty"`
	Params      []string `json:"params,omitempty"`
	TypeParams  []string `json:"type_params,omitempty"`
}

func (f Function) String() string {
	receiverPrefix := ""
	if f.Receiver != "" {
		receiverPrefix = f.Receiver + "."
	}
	return fmt.Sprintf("%s%s(%s) %s", receiverPrefix, f.Name, strings.Join(f.TypeParams, ","), strings.Join(f.ReturnTypes, ","))
}

// APIstruct represents a struct in the codebase.
type APIstruct struct {
	Name   string   `json:"name"`
	Fields []string `json:"fields"`
}

func (f APIstruct) String() string {
	return fmt.Sprintf("%s(%s)", f.Name, strings.Join(f.Fields, ","))
}

type Interface struct {
	Name    string     `json:"name"`
	Methods []Function `json:"methods"`
}

func (f Interface) String() string {
	methodsStr := make([]string, len(f.Methods))
	for i, m := range f.Methods {
		methodsStr[i] = m.String()
	}
	return fmt.Sprintf("%s{%s}", f.Name, strings.Join(methodsStr, ","))
}

// API represents the API of the codebase, including functions and structs.
type API struct {
	Values     []string    `json:"values,omitempty"`
	Structs    []APIstruct `json:"structs,omitempty"`
	Functions  []Function  `json:"functions,omitempty"`
	Interfaces []Interface `json:"interfaces,omitempty"`
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
	Diff             DiffConfig            `yaml:"diff"`
}

// UnkeyedLiteral represents the configuration for unkeyed literal initialization.
type UnkeyedLiteral struct {
	Enabled bool `yaml:"enabled"`
	Limit   int  `yaml:"limit"`
}

type DiffConfig struct {
	ErrorOnAddition bool   `yaml:"error_on_addition"`
	ErrorOnRemoval  bool   `yaml:"error_on_removal"`
	WriteDiff       bool   `yaml:"write_diff"`
	Path            string `yaml:"path"`
}
