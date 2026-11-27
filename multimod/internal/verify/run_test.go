// Copyright The OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package verify

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/go-git/go-git/v5"
	"github.com/stretchr/testify/require"

	"go.opentelemetry.io/build-tools/multimod/internal/shared/sharedtest"
)

// setupRun creates a git repo containing the given versions.yaml and go.mod
// files, makes it the working directory, and returns the versions.yaml path.
func setupRun(t *testing.T, versions string, goMods map[string]string) string {
	t.Helper()
	root := t.TempDir()
	_, err := git.PlainInit(root, false)
	require.NoError(t, err)

	files := map[string][]byte{filepath.Join(root, "versions.yaml"): []byte(versions)}
	for p, content := range goMods {
		files[filepath.Join(root, p)] = []byte(content)
	}
	require.NoError(t, sharedtest.WriteTempFiles(files))

	t.Chdir(root)
	return filepath.Join(root, "versions.yaml")
}

func TestRun(t *testing.T) {
	versions := `module-sets:
  a:
    version: v1.0.0
    modules:
      - example.com/a
  b:
    version: v0.1.0
    modules:
      - example.com/a/b
`
	goMods := map[string]string{
		"go.mod":   "module example.com/a\n\ngo 1.21\n\nrequire example.com/a/b v0.1.0\n",
		"b/go.mod": "module example.com/a/b\n\ngo 1.21\n",
	}
	// A stable module depending on an unstable one is only logged, not an error.
	require.NoError(t, Run(setupRun(t, versions, goMods)))
}

func TestRunErrors(t *testing.T) {
	const oneMod = "module example.com/a\n\ngo 1.21\n"

	t.Run("not in a repository", func(t *testing.T) {
		t.Chdir(t.TempDir())
		err := Run("versions.yaml")
		require.ErrorContains(t, err, "unable to find repo root")
	})

	t.Run("missing versioning file", func(t *testing.T) {
		root := setupRun(t, "", nil)
		require.NoError(t, os.Remove(root))
		err := Run(root)
		require.ErrorContains(t, err, "Error creating new verification struct")
	})

	t.Run("module not in a set", func(t *testing.T) {
		versions := "module-sets:\n  a:\n    version: v1.0.0\n    modules:\n      - example.com/a\n"
		err := Run(setupRun(t, versions, map[string]string{
			"go.mod":   oneMod,
			"b/go.mod": "module example.com/b\n\ngo 1.21\n",
		}))
		require.ErrorContains(t, err, "verifyAllModulesInSet failed")
	})

	t.Run("two stable sets with the same major version", func(t *testing.T) {
		versions := `module-sets:
  a:
    version: v1.0.0
    modules:
      - example.com/a
  b:
    version: v1.1.0
    modules:
      - example.com/a/b
`
		err := Run(setupRun(t, versions, map[string]string{
			"go.mod":   oneMod,
			"b/go.mod": "module example.com/a/b\n\ngo 1.21\n",
		}))
		require.ErrorContains(t, err, "verifyVersions failed")
	})

	t.Run("unparsable go.mod", func(t *testing.T) {
		versions := "module-sets:\n  a:\n    version: v1.0.0\n    modules:\n      - example.com/a\n"
		root := setupRun(t, versions, map[string]string{"go.mod": oneMod})
		// Valid while the versioning is loaded, broken by the time dependencies are read.
		require.NoError(t, os.WriteFile(filepath.Join(filepath.Dir(root), "go.mod"), []byte("module example.com/a\n\nrequire (\n"), 0o600))
		require.Error(t, Run(root))
	})
}
