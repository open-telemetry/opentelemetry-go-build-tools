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
	"github.com/spf13/cobra"
	cl "go.opentelemetry.io/build-tools/crosslink/internal"
)

// pruneCmd represents the prune command
var pruneCmd = &cobra.Command{
	Use:   "prune",
	Short: "Remove unnecessary replace statements from intra-repository go.mod files",
	Long: `Prune will analyze and remove any uncessary replace statements for intra-repository
	go.mod files that are not direct or transitive dependencies for intra-repository modules. 
	This is a destructive action and will overwrite existing go.mod files. 
	Prune will not remove modules that fall outside of the root module namespace.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return cl.Prune(rc)
	},
}
