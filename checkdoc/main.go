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
//
// Deprecated: Please use checkfile with argument --file-name README.md instead.
package main

import (
	"go.opentelemetry.io/build-tools/internal/check"
)

const (
	// The name of the Readme file
	readMeFileName = "README.md"
)

// The main verifies if README.md and proper documentations for the enabled default components
// are existed in OpenTelemetry core and contrib repository.
// Usage in the core repo:
//
//	checkdoc --project-path path/to/project \
//				--component-rel-path service/defaultcomponents/defaults.go \
//				--module-name go.opentelemetry.io/collector
//
// Usage in the contrib repo:
//
//	checkdoc --project-path path/to/project \
//				--component-rel-path cmd/otelcontrib/components.go \
//				--module-name github.com/open-telemetry/opentelemetry-collector-contrib
func main() {
	projectPath, componentPath, moduleName, _ := check.Flags()

	err := check.FileExists(
		*projectPath,
		*componentPath,
		*moduleName,
		readMeFileName,
	)

	if err != nil {
		panic(err)
	}
}
