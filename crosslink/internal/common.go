package crosslink

import (
	"os"

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
