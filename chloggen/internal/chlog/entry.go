// Copyright The OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package chlog

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"

	"go.opentelemetry.io/build-tools/chloggen/internal/config"
)

const (
	Breaking     = "breaking"
	Deprecation  = "deprecation"
	NewComponent = "new_component"
	Enhancement  = "enhancement"
	BugFix       = "bug_fix"
)

type Entry struct {
	ChangeType string `yaml:"change_type"`
	Component  string `yaml:"component"`
	Note       string `yaml:"note"`
	Issues     []int  `yaml:"issues"`
	SubText    string `yaml:"subtext"`
}

var changeTypes = []string{
	Breaking,
	Deprecation,
	NewComponent,
	Enhancement,
	BugFix,
}

func (e Entry) Validate() error {
	var validType bool
	for _, ct := range changeTypes {
		if e.ChangeType == ct {
			validType = true
			break
		}
	}
	if !validType {
		return fmt.Errorf("'%s' is not a valid 'change_type'. Specify one of %v", e.ChangeType, changeTypes)
	}

	if strings.TrimSpace(e.Component) == "" {
		return fmt.Errorf("specify a 'component'")
	}

	if strings.TrimSpace(e.Note) == "" {
		return fmt.Errorf("specify a 'note'")
	}

	if len(e.Issues) == 0 {
		return fmt.Errorf("specify one or more issues #'s")
	}

	return nil
}

func (e Entry) String() string {
	issueStrs := make([]string, 0, len(e.Issues))
	for _, issue := range e.Issues {
		issueStrs = append(issueStrs, fmt.Sprintf("#%d", issue))
	}
	issueStr := strings.Join(issueStrs, ", ")

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("- `%s`: %s (%s)", e.Component, e.Note, issueStr))
	if e.SubText != "" {
		sb.WriteString("\n  ")
		lines := strings.Split(strings.ReplaceAll(e.SubText, "\r\n", "\n"), "\n")
		sb.WriteString(strings.Join(lines, "\n  "))
	}
	return sb.String()
}

func ReadEntries(cfg config.Config) ([]*Entry, error) {
	entryYAMLs, err := filepath.Glob(filepath.Join(cfg.ChlogsDir, "*.yaml"))
	if err != nil {
		return nil, err
	}

	entries := make([]*Entry, 0, len(entryYAMLs))
	for _, entryYAML := range entryYAMLs {
		if entryYAML == cfg.TemplateYAML || entryYAML == cfg.ConfigYAML {
			continue
		}

		fileBytes, err := os.ReadFile(filepath.Clean(entryYAML))
		if err != nil {
			return nil, err
		}

		entry := &Entry{}
		if err = yaml.Unmarshal(fileBytes, entry); err != nil {
			return nil, err
		}
		entries = append(entries, entry)
	}
	return entries, nil
}

func DeleteEntries(cfg config.Config) error {
	entryYAMLs, err := filepath.Glob(filepath.Join(cfg.ChlogsDir, "*.yaml"))
	if err != nil {
		return err
	}

	for _, entryYAML := range entryYAMLs {
		if filepath.Base(entryYAML) == filepath.Base(cfg.TemplateYAML) {
			continue
		}

		if err := os.Remove(entryYAML); err != nil {
			fmt.Printf("Failed to delete: %s\n", entryYAML)
		}
	}
	return nil
}
