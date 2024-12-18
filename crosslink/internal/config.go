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

package crosslink

import (
	"log"

	"go.uber.org/zap"
	"golang.org/x/mod/modfile"
)

type moduleInfo struct {
	moduleContents            modfile.File
	requiredReplaceStatements map[string]struct{}
}

func newModuleInfo(moduleContents modfile.File) *moduleInfo {
	return &moduleInfo{
		requiredReplaceStatements: make(map[string]struct{}),
		moduleContents:            moduleContents,
	}
}

type RunConfig struct {
	RootPath      string
	Verbose       bool
	ExcludedPaths map[string]struct{}
	SkippedPaths  map[string]struct{}
	Overwrite     bool
	Prune         bool
	GoVersion     string
	AllowCircular string
	Validate      bool
	Logger        *zap.Logger
}

func DefaultRunConfig() RunConfig {
	lg, err := zap.NewProduction()
	if err != nil {
		log.Printf("Could not create zap logger: %v", err)
	}
	ep := make(map[string]struct{})
	rc := RunConfig{
		Logger:        lg,
		ExcludedPaths: ep,
	}
	return rc
}
