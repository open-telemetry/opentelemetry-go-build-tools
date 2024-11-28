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
	ChangeLogs   []string `yaml:"change_logs"`
	ChangeType   string   `yaml:"change_type"`
	Component    string   `yaml:"component"`
	Note         string   `yaml:"note"`
	Issues       []int    `yaml:"issues"`
	PullRequests []int    `yaml:"pull_requests"`
	SubText      string   `yaml:"subtext"`
}

var changeTypes = []string{
	Breaking,
	Deprecation,
	NewComponent,
	Enhancement,
	BugFix,
}

func (e Entry) Validate(requireChangelog bool, components []string, validChangeLogs ...string) error {
	if requireChangelog && len(e.ChangeLogs) == 0 {
		return fmt.Errorf("specify one or more 'change_logs'")
	}
	for _, cl := range e.ChangeLogs {
		var valid bool
		for _, vcl := range validChangeLogs {
			if cl == vcl {
				valid = true
			}
		}
		if !valid {
			return fmt.Errorf("'%s' is not a valid value in 'change_logs'. Specify one of %v", cl, validChangeLogs)
		}
	}

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

	found := false
	for _, validComp := range components {
		if e.Component == validComp {
			found = true
			break
		}
	}
	// only apply component validation if one or more values are present.
	if len(components) > 0 && !found {
		return fmt.Errorf("%s is not a valid 'component'. It must be one of %v", e.Component, components)
	}

	if strings.TrimSpace(e.Note) == "" {
		return fmt.Errorf("specify a 'note'")
	}

	if len(e.Issues) == 0 && len(e.PullRequests) == 0 {
		return fmt.Errorf("specify one or more issues or pull requests #'s")
	}

	return nil
}

func ReadEntries(cfg *config.Config) (map[string][]*Entry, error) {
	yamlFiles, err := filepath.Glob(filepath.Join(cfg.EntriesDir, "*.yaml"))
	if err != nil {
		return nil, err
	}

	entries := make(map[string][]*Entry)
	for key := range cfg.ChangeLogs {
		entries[key] = make([]*Entry, 0)
	}

	for _, file := range yamlFiles {
		if file == cfg.TemplateYAML || file == cfg.ConfigYAML {
			continue
		}

		fileBytes, err := os.ReadFile(filepath.Clean(file))
		if err != nil {
			return nil, err
		}

		entry := &Entry{}
		if err = yaml.Unmarshal(fileBytes, entry); err != nil {
			return nil, err
		}
		entry.SubText = strings.ReplaceAll(entry.SubText, "\r\n", "\n")

		if len(entry.ChangeLogs) == 0 {
			for _, cl := range cfg.DefaultChangeLogs {
				entries[cl] = append(entries[cl], entry)
			}
		} else {
			for _, cl := range entry.ChangeLogs {
				entries[cl] = append(entries[cl], entry)
			}
		}
	}
	return entries, nil
}

func DeleteEntries(cfg *config.Config) error {
	yamlFiles, err := filepath.Glob(filepath.Join(cfg.EntriesDir, "*.yaml"))
	if err != nil {
		return err
	}

	for _, file := range yamlFiles {
		if file == cfg.TemplateYAML || file == cfg.ConfigYAML {
			continue
		}

		if err := os.Remove(file); err != nil {
			fmt.Printf("Failed to delete: %s\n", file)
		}
	}
	return nil
}
