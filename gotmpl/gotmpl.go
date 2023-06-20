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

package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"text/template"
)

// gotmpl uses [text/template] package to generate a file using
// bodyPath as template body filepath,
// jsonData as data in JSON format,
// outPath as output filepath.
func gotmpl(bodyPath, jsonData, outPath string) error {
	if bodyPath == "" {
		return errors.New("gotmpl: template body filepath must be set")
	}
	if outPath == "" {
		return errors.New("gotmpl: output filepath must be set")
	}

	tmpl, err := template.ParseFiles(bodyPath)
	if err != nil {
		return fmt.Errorf("gotmpl: cannot parse template body file: %w", err)
	}

	var data any
	if err = json.Unmarshal(([]byte)(jsonData), &data); err != nil {
		return fmt.Errorf("gotmpl: data must be in JSON format: %w", err)
	}

	outFile, err := os.Create(outPath) //nolint:gosec // This is a file generation tool that takes filepath as output.
	if err != nil {
		return fmt.Errorf("gotmpl: cannot create output file: %w", err)
	}
	defer outFile.Close()

	if err := tmpl.Option("missingkey=error").Execute(outFile, data); err != nil {
		return fmt.Errorf("gotmpl: execution failed: %w", err)
	}
	if err := outFile.Close(); err != nil {
		return fmt.Errorf("gotmpl: cannot close output file: %w", err)
	}
	return nil
}
