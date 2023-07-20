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

package cmd

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"

	"go.opentelemetry.io/build-tools/chloggen/internal/chlog"
	"go.opentelemetry.io/build-tools/chloggen/internal/config"
)

func getSampleEntries() []*chlog.Entry {
	return []*chlog.Entry{
		enhancementEntry(),
		bugFixEntry(),
		deprecationEntry(),
		newComponentEntry(),
		breakingEntry(),
		entryWithSubtext(),
	}
}

func enhancementEntry() *chlog.Entry {
	return &chlog.Entry{
		ChangeType: chlog.Enhancement,
		Component:  "receiver/foo",
		Note:       "Add some bar",
		Issues:     []int{12345},
	}
}

func bugFixEntry() *chlog.Entry {
	return &chlog.Entry{
		ChangeType: chlog.BugFix,
		Component:  "testbed",
		Note:       "Fix blah",
		Issues:     []int{12346, 12347},
	}
}

func deprecationEntry() *chlog.Entry {
	return &chlog.Entry{
		ChangeType: chlog.Deprecation,
		Component:  "exporter/old",
		Note:       "Deprecate old",
		Issues:     []int{12348},
	}
}

func newComponentEntry() *chlog.Entry {
	return &chlog.Entry{
		ChangeType: chlog.NewComponent,
		Component:  "exporter/new",
		Note:       "Add new exporter ...",
		Issues:     []int{12349},
	}
}

func breakingEntry() *chlog.Entry {
	return &chlog.Entry{
		ChangeType: chlog.Breaking,
		Component:  "processor/oops",
		Note:       "Change behavior when ...",
		Issues:     []int{12350},
	}
}

func entryWithSubtext() *chlog.Entry {
	lines := []string{"- foo\n  - bar\n- blah\n  - 1234567"}

	return &chlog.Entry{
		ChangeType: chlog.Breaking,
		Component:  "processor/oops",
		Note:       "Change behavior when ...",
		Issues:     []int{12350},
		SubText:    strings.Join(lines, "\n"),
	}
}

func setupTestDir(t *testing.T, entries []*chlog.Entry) config.Config {
	cfg := config.New(t.TempDir())

	// Create a known dummy changelog which may be updated by the test
	changelogBytes, err := os.ReadFile(filepath.Join("testdata", config.DefaultChangelogMD))
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(cfg.ChangelogMD, changelogBytes, os.FileMode(0755)))

	require.NoError(t, os.Mkdir(cfg.ChlogsDir, os.FileMode(0755)))

	// Copy the entry template, for tests that ensure it is not deleted
	templateInRootDir := config.New("testdata").TemplateYAML
	templateBytes, err := os.ReadFile(filepath.Clean(templateInRootDir))
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(cfg.TemplateYAML, templateBytes, os.FileMode(0755)))

	for i, entry := range entries {
		require.NoError(t, writeEntryYAML(cfg, fmt.Sprintf("%d.yaml", i), entry))
	}

	return cfg
}

func writeEntryYAML(cfg config.Config, filename string, entry *chlog.Entry) error {
	entryBytes, err := yaml.Marshal(entry)
	if err != nil {
		return err
	}
	path := filepath.Join(cfg.ChlogsDir, filename)
	return os.WriteFile(path, entryBytes, os.FileMode(0755))
}

func runCobra(t *testing.T, args ...string) (string, string) {
	cmd := rootCmd()

	outBytes := bytes.NewBufferString("")
	cmd.SetOut(outBytes)

	errBytes := bytes.NewBufferString("")
	cmd.SetErr(errBytes)

	cmd.SetArgs(args)
	cmd.Execute() // nolint:errcheck

	out, ioErr := io.ReadAll(outBytes)
	require.NoError(t, ioErr, "read stdout")

	err, ioErr := io.ReadAll(errBytes)
	require.NoError(t, ioErr, "read stderr")

	return string(out), string(err)
}
