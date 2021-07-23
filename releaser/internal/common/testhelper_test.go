package common

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMockModuleVersioning(t *testing.T) {
	modSetMap := ModuleSetMap{
		"mod-set-1": ModuleSet{
			Version: "v1.2.3-RC1+meta",
			Modules: []ModulePath{
				"go.opentelemetry.io/test/test1",
				"go.opentelemetry.io/test/test2",
			},
		},
		"mod-set-2": ModuleSet{
			Version: "v0.1.0",
			Modules: []ModulePath{
				"go.opentelemetry.io/test3",
			},
		},
	}

	modPathMap := ModulePathMap{
		"go.opentelemetry.io/test/test1": "root/path/to/mod/test/test1/go.mod",
		"go.opentelemetry.io/test/test2": "root/path/to/mod/test/test2/go.mod",
		"go.opentelemetry.io/test3":      "root/test3/go.mod",
	}

	expected := ModuleVersioning{
		ModSetMap:  modSetMap,
		ModPathMap: modPathMap,
		ModInfoMap: ModuleInfoMap{
			"go.opentelemetry.io/test/test1": ModuleInfo{
				ModuleSetName: "mod-set-1",
				Version:       "v1.2.3-RC1+meta",
			},
			"go.opentelemetry.io/test/test2": ModuleInfo{
				ModuleSetName: "mod-set-1",
				Version:       "v1.2.3-RC1+meta",
			},
			"go.opentelemetry.io/test3": ModuleInfo{
				ModuleSetName: "mod-set-2",
				Version:       "v0.1.0",
			},
		},
	}

	actual, err := MockModuleVersioning(modSetMap, modPathMap)

	require.NoError(t, err)
	assert.Equal(t, expected, actual)
}

func TestWriteGoModFiles(t *testing.T) {
	fmt.Println(filepath.Abs(testDataDir))
	tmpRootDir, err := os.MkdirTemp(testDataDir, "NewModuleVersioning")
	if err != nil {
		t.Fatal("creating temp dir:", err)
	}

	defer os.RemoveAll(tmpRootDir)

	modFiles := map[ModuleFilePath][]byte{
		ModuleFilePath(filepath.Join(tmpRootDir, "test", "test1", "go.mod")): []byte("module \"go.opentelemetry.io/test/test1\"\n\ngo 1.16\n\n" +
			"require (\n\t\"go.opentelemetry.io/testroot/v2\" v2.0.0\n)\n"),
		ModuleFilePath(filepath.Join(tmpRootDir, "test", "go.mod")):          []byte("module go.opentelemetry.io/test3\n\ngo 1.16\n"),
		ModuleFilePath(filepath.Join(tmpRootDir, "go.mod")):                  []byte("module go.opentelemetry.io/testroot/v2\n\ngo 1.16\n"),
		ModuleFilePath(filepath.Join(tmpRootDir, "test", "test2", "go.mod")): []byte("module \"go.opentelemetry.io/test/testexcluded\"\n\ngo 1.16\n"),
	}

	if err := WriteGoModFiles(modFiles); err != nil {
		t.Fatal("could not create go mod file tree", err)
	}

	// check all mod files have been written correctly
	for modPath, expectedModFile := range modFiles {
		actual, err := ioutil.ReadFile(string(modPath))

		require.NoError(t, err)
		assert.Equal(t, expectedModFile, actual)
	}
}
