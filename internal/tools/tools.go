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

//go:build tools
// +build tools

package tools

import (
	_ "github.com/client9/misspell/cmd/misspell"
	_ "github.com/gogo/protobuf/protoc-gen-gogofast"
	_ "github.com/golangci/golangci-lint/v2/cmd/golangci-lint"
	_ "github.com/matryer/moq"
	_ "go.opentelemetry.io/build-tools/chloggen"
	_ "go.opentelemetry.io/build-tools/crosslink"
	_ "go.opentelemetry.io/build-tools/dbotconf"
	_ "go.opentelemetry.io/build-tools/multimod"
	_ "golang.org/x/vuln/cmd/govulncheck"
)
