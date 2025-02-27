package common

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"golang.org/x/mod/modfile"
)

func newAllModulePathMap(root string) (ModulePathMap, error) {
	modPathMap := make(ModulePathMap)

	findGoMod := func(filePath string, _ fs.FileInfo, err error) error {
		if err != nil {
			fmt.Printf("Warning: file could not be read during filepath.Walk(): %v", err)
			return nil
		}
		if filepath.Base(filePath) == "go.mod" {
			// read go.mod file into mod []byte
			mod, err := os.ReadFile(filepath.Clean(filePath))
			if err != nil {
				return err
			}

			// read path of module from go.mod file
			modPathString := modfile.ModulePath(mod)

			// convert modPath, filePath string to modulePath and moduleFilePath
			modPath := ModulePath(modPathString)
			modFilePath := ModuleFilePath(filePath)

			modPathMap[modPath] = modFilePath
		}
		return nil
	}

	if err := filepath.Walk(root, findGoMod); err != nil {
		return nil, err
	}

	return modPathMap, nil
}
