// Copyright The OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package report

import (
	"os"
	"path"

	"github.com/joshdk/go-junit"
	"go.uber.org/zap"
)

type GeneratorConfig struct {
	ArtifactsPath string
}

type Generator struct {
	logger *zap.Logger
	cfg    GeneratorConfig
}

func NewGenerator(logger *zap.Logger, cfg GeneratorConfig) *Generator {
	return &Generator{
		logger: logger,
		cfg:    cfg,
	}
}

func (rg *Generator) ingestArtifacts() map[string]junit.Suite {
	files, err := os.ReadDir(rg.cfg.ArtifactsPath)
	if err != nil {
		rg.logger.Warn(
			"Failed to read directory with test results",
			zap.Error(err),
			zap.String("path", rg.cfg.ArtifactsPath),
		)
	}

	testSuites := map[string]junit.Suite{}
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		rg.logger.Info("Ingesting test reports", zap.String("path", file.Name()))
		suites, err := junit.IngestFile(path.Join(rg.cfg.ArtifactsPath, file.Name()))
		if err != nil {
			rg.logger.Error(
				"Failed to ingest JUnit xml, omitting test results from report",
				zap.String("path", file.Name()),
				zap.Error(err),
			)
			continue
		}

		for _, s := range suites {
			testSuites[s.Name] = s
		}
	}
	return testSuites
}

// ProcessTestResults iterates over the test results and matches the module
// with the code owners, creating a report.
func (rg *Generator) ProcessTestResults() []Report {
	var reports []Report
	testSuites := rg.ingestArtifacts()
	for module, suite := range testSuites {
		if suite.Totals.Failed == 0 {
			continue
		}

		r := Report{
			Module:      module,
			FailedTests: make(map[string]string, suite.Totals.Failed),
		}
		for _, t := range suite.Tests {
			if t.Status == junit.StatusFailed {
				r.FailedTests[t.Name] = t.Error.Error()
			}
		}
		reports = append(reports, r)
	}
	return reports
}
