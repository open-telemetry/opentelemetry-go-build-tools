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

package internal

const (
	version2     = 2
	ghPkgEco     = "github-actions"
	dockerPkgEco = "docker"
	gomodPkgEco  = "gomod"
)

var (
	weeklySchedule = schedule{Interval: "weekly", Day: "sunday"}
	actionLabels   = []string{"dependencies", "actions", "Skip Changelog"}
	dockerLabels   = []string{"dependencies", "docker", "Skip Changelog"}
	goLabels       = []string{"dependencies", "go", "Skip Changelog"}
)

type dependabotConfig struct {
	Version int
	Updates []update
}

type update struct {
	PackageEcosystem string `yaml:"package-ecosystem"`
	Directory        string
	Labels           []string `yaml:",omitempty"`
	Schedule         schedule
}

type schedule struct {
	Interval string
	Day      string `yaml:",omitempty"`
}
