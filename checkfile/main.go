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

package main

import (
	"go.opentelemetry.io/build-tools/internal/check"
)

// The main verifies if filename exists for the enabled default components
// in OpenTelemetry core and contrib repository.
// Usage in the core repo:
//
//	checkfile --project-path path/to/project \
//				--component-rel-path service/defaultcomponents/defaults.go \
//				--module-name go.opentelemetry.io/collector
//				--file-name README.md
//
// Usage in the contrib repo:
//
//	checkfile --project-path path/to/project \
//				--component-rel-path cmd/otelcontrib/components.go \
//				--module-name github.com/open-telemetry/opentelemetry-collector-contrib
//				--file-name metadata.yaml
func main() {
	projectPath, componentPath, moduleName, fileName := check.Flags()
	if *fileName == "" {
		panic("Missing required argument: --file-name")
	}

	err := check.FileExists(
		*projectPath,
		*componentPath,
		*moduleName,
		*fileName,
	)

	if err != nil {
		panic(err)
	}
}
