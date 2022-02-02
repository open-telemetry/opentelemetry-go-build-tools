package crosslink

import (
	"os"
	"path/filepath"

	tools "go.opentelemetry.io/build-tools"
	"golang.org/x/mod/modfile"
)

func writeModules(module moduleInfo) error {
	mfParsed, err := modfile.Parse("go.mod", module.moduleContents, nil)
	if err != nil {
		return err
	}
	//  now overwrite the existing gomod file
	gomodFile, err := mfParsed.Format()
	if err != nil {
		return err
	}
	//write our updated go.mod file
	err = os.WriteFile(module.moduleFilePath, gomodFile, 0700)
	if err != nil {
		return err
	}

	return nil
}

func identifyRootModule(r string) (string, error) {
	var err error
	rootPath := r
	if rootPath == "" {
		rootPath, err = tools.FindRepoRoot()
		if err != nil {
			return "", err
		}
	}

	if _, err := os.Stat(filepath.Join(rootPath, "go.mod")); err != nil {
		return "", err
	}

	// identify and read the root module
	rootModPath := filepath.Join(rootPath, "go.mod")
	rootModFile, err := os.ReadFile(rootModPath)
	if err != nil {
		return "", err
	}
	return modfile.ModulePath(rootModFile), nil
}
