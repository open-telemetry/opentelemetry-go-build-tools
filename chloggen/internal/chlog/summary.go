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
	"bytes"
	_ "embed"
	"fmt"
	"sort"
	"text/template"

	"go.opentelemetry.io/build-tools/chloggen/internal/config"
)

//go:embed summary.tmpl
var tmpl []byte

type summary struct {
	Version         string
	BreakingChanges []string
	Deprecations    []string
	NewComponents   []string
	Enhancements    []string
	BugFixes        []string
}

func GenerateSummary(version string, entries []*Entry, cfg *config.Config) (string, error) {
	s := summary{
		Version: version,
	}

	for _, entry := range entries {
		switch entry.ChangeType {
		case Breaking:
			s.BreakingChanges = append(s.BreakingChanges, entry.String(cfg))
		case Deprecation:
			s.Deprecations = append(s.Deprecations, entry.String(cfg))
		case NewComponent:
			s.NewComponents = append(s.NewComponents, entry.String(cfg))
		case Enhancement:
			s.Enhancements = append(s.Enhancements, entry.String(cfg))
		case BugFix:
			s.BugFixes = append(s.BugFixes, entry.String(cfg))
		}
	}

	s.BreakingChanges = sort.StringSlice(s.BreakingChanges)
	s.Deprecations = sort.StringSlice(s.Deprecations)
	s.NewComponents = sort.StringSlice(s.NewComponents)
	s.Enhancements = sort.StringSlice(s.Enhancements)
	s.BugFixes = sort.StringSlice(s.BugFixes)

	return s.String()
}

func (s summary) String() (string, error) {
	tmpl := template.Must(
		template.
			New("summary.tmpl").
			Option("missingkey=error").
			Parse(string(tmpl)))

	buf := bytes.Buffer{}
	if err := tmpl.Execute(&buf, s); err != nil {
		return "", fmt.Errorf("failed executing template: %w", err)
	}

	return buf.String(), nil
}
