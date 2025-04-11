// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package internal

import (
	"fmt"
	"strings"
)

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

type Apistruct struct {
	Name   string   `json:"name"`
	Fields []string `json:"fields"`
}

func (f Apistruct) String() string {
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

type API struct {
	Values     []string    `json:"values,omitempty"`
	Structs    []Apistruct `json:"structs,omitempty"`
	Functions  []Function  `json:"functions,omitempty"`
	Interfaces []Interface `json:"interfaces,omitempty"`
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
	Diff             DiffConfig            `yaml:"diff"`
}

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
