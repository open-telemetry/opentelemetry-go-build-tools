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

package verify

import (
	"fmt"

	"go.opentelemetry.io/build-tools/releaser/internal/common"
)

type errModuleNotInSet struct {
	modPath     common.ModulePath
	modFilePath common.ModuleFilePath
}

func (e *errModuleNotInSet) Error() string {
	return fmt.Sprintf("Module %v (defined in %v) is not listed in any module set.", e.modPath, e.modFilePath)
}

type errModuleNotInRepo struct {
	modPath    common.ModulePath
	modSetName string
}

func (e *errModuleNotInRepo) Error() string {
	return fmt.Sprintf("Module %v in module set %v does not exist in the current repo.", e.modPath, e.modSetName)
}

type errInvalidVersion struct {
	modSetName    string
	modSetVersion string
}

func (e *errInvalidVersion) Error() string {
	return fmt.Sprintf("Module set %v has invalid version string: %v", e.modSetName, e.modSetVersion)
}

type errMultipleSetSameVersion struct {
	modSetNames   []string
	modSetVersion string
}

func (e *errMultipleSetSameVersion) Error() string {
	return fmt.Sprintf("Multiple module sets have the same major version (%v): %v",
		e.modSetVersion, e.modSetNames)
}

// errDependency is logged upon discovery that a stable module depends on an unstable module.
type errDependency struct {
	modPath    common.ModulePath
	modVersion string
	depPath    common.ModulePath
	depVersion string
}

func (e *errDependency) Error() string {
	return fmt.Sprintf("WARNING: Stable module %v (%v) depends on unstable module %v (%v).\n",
		e.modPath, e.modVersion,
		e.depPath, e.depVersion)
}
