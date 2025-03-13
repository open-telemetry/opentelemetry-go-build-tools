// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package datatype

//go:generate moq -pkg fake -skip-ensure -out ./fake/mock_generator.go . Generator:MockGenerator
type Generator interface {
	Generate(data GithubData) error
}

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
	Chloggen          ChloggenConfig
}

type Codeowners struct {
	// Active codeowners
	Active []string `mapstructure:"active"`
	// Emeritus codeowners
	Emeritus []string `mapstructure:"emeritus"`
}

type Status struct {
	Stability     map[string][]string `mapstructure:"stability"`
	Distributions []string            `mapstructure:"distributions"`
	Class         string              `mapstructure:"class"`
	Warnings      []string            `mapstructure:"warnings"`
	Codeowners    *Codeowners         `mapstructure:"codeowners"`
}
type Metadata struct {
	// Type of the component.
	Type string `mapstructure:"type"`
	// Type of the parent component (applicable to subcomponents).
	Parent string `mapstructure:"parent"`
	// Status information for the component.
	Status *Status `mapstructure:"status"`
}

type DistributionData struct {
	Name        string   `yaml:"name"`
	URL         string   `yaml:"url"`
	Maintainers []string `yaml:"maintainers,omitempty"`
}

type ChloggenConfig struct {
	ChangeLogs        map[string]string `yaml:"change_logs"`
	DefaultChangeLogs []string          `yaml:"default_change_logs"`
	EntriesDir        string            `yaml:"entries_dir"`
	TemplateYAML      string            `yaml:"template_yaml"`
	SummaryTemplate   string            `yaml:"summary_template"`
	Components        []string          `yaml:"components"`
}
