// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

// Package datatype contains the data structures used in the generator.
package datatype

// Generator is an interface for generating metadata files.
//
//go:generate moq -pkg fake -skip-ensure -out ./fake/mock_generator.go . Generator:MockGenerator
type Generator interface {
	Generate(data GithubData) error
}

// GithubData represents the data structure for GitHub metadata.
type GithubData struct {
	RootFolder        string
	Folders           []string
	Codeowners        []string
	AllowlistFilePath string
	MaxLength         int
	Components        map[string]Metadata
	Distributions     []DistributionData
	DefaultCodeOwner  string
	GitHubOrg         string
	GitHubTeam        string
	Chloggen          ChloggenConfig
}

// Codeowners represents the code owners for a component.
type Codeowners struct {
	// Active codeowners
	Active []string `yaml:"active"`
	// Emeritus codeowners
	Emeritus []string `yaml:"emeritus"`
	// Whether to leave go.mod and go.sum assigned to default codeowner
	ExcludeGoMod bool `yaml:"exclude_go_mod"`
}

// Status represents the status of a component.
type Status struct {
	Stability     map[string][]string `yaml:"stability"`
	Distributions []string            `yaml:"distributions"`
	Class         string              `yaml:"class"`
	Warnings      []string            `yaml:"warnings"`
	Codeowners    *Codeowners         `yaml:"codeowners"`
}

// Metadata represents the metadata for a component.
type Metadata struct {
	// Type of the component.
	Type string `yaml:"type"`
	// Status information for the component.
	Status *Status `yaml:"status"`
}

// DistributionData represents the distribution data for a component.
type DistributionData struct {
	Name        string   `yaml:"name"`
	URL         string   `yaml:"url"`
	Maintainers []string `yaml:"maintainers,omitempty"`
	None        bool     `yaml:"none"` // denotes a special distribution of all the components that are not in any distribution.
}

// ChloggenConfig represents the configuration for the changelog generator.
type ChloggenConfig struct {
	ChangeLogs        map[string]string `yaml:"change_logs"`
	DefaultChangeLogs []string          `yaml:"default_change_logs"`
	EntriesDir        string            `yaml:"entries_dir"`
	TemplateYAML      string            `yaml:"template_yaml"`
	SummaryTemplate   string            `yaml:"summary_template"`
	Components        []string          `yaml:"components"`
}
