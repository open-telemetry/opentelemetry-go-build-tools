package crosslink

import (
	"fmt"
	"path/filepath"
	"strings"

	"golang.org/x/mod/modfile"
)

func Crosslink(rc runConfig) {
	var err error
	defer rc.logger.Sync()

	rootModulePath, err := identifyRootModule(rc.RootPath)
	if err != nil {
		panic(fmt.Sprintf("failed to identify root module: %v", err))
	}

	graph, err := buildDepedencyGraph(rc, rootModulePath)
	if err != nil {
		panic(fmt.Sprintf("failed to build dependency graph: %v", err))
	}

	for _, moduleInfo := range graph {
		err = insertReplace(&moduleInfo, rc)
		if err != nil {
			panic(fmt.Sprintf("failed to insert replace statements: %v", err))
		}

		err = pruneReplace(rootModulePath, &moduleInfo, rc)

		if err != nil {
			panic(fmt.Sprintf("error pruning replace statements: %v", err))
		}

		err = writeModules(moduleInfo)
		if err != nil {
			panic(fmt.Sprintf("error writing gomod files: %v", err))
		}
	}
}

func insertReplace(module *moduleInfo, rc runConfig) error {
	// modfile type that we will work with then write to the mod file in the end
	mfParsed, err := modfile.Parse("gomod", module.moduleContents, nil)
	if err != nil {
		return err
	}

	for reqModule := range module.requiredReplaceStatements {
		// skip excluded
		if _, exists := rc.ExcludedPaths[reqModule]; exists {
			if rc.Verbose {
				rc.logger.Sugar().Infof("Excluded Module %s, ignoring replace", reqModule)
			}
			continue
		}

		localPath, err := filepath.Rel(mfParsed.Module.Mod.Path, reqModule)
		if err != nil {
			return err
		}
		if localPath == "." || localPath == ".." {
			localPath += "/"
		} else if !strings.HasPrefix(localPath, "..") {
			localPath = "./" + localPath
		}
		var loggerStr string
		// see if replace statement already exists for module. Verify if it's the same. If it does not exist then add it.
		// AddReplace should handle all of these conditions in terms of add and/or verifying
		// https://cs.opensource.google/go/go/+/master:src/cmd/vendor/golang.org/x/mod/modfile/rule.go;l=1296?q=addReplace
		if oldReplace, exists := containsReplace(mfParsed.Replace, reqModule); exists {
			if rc.Overwrite {
				loggerStr = fmt.Sprintf("Overwriting: Module: %s Old: %s => %s New: %s => %s", mfParsed.Module.Mod.Path, reqModule, oldReplace.New.Path, reqModule, localPath)
				mfParsed.AddReplace(reqModule, "", localPath, "")
			} else {
				loggerStr = fmt.Sprintf("Replace already exists: Module: %s : %s => %s \n run with -overwrite flag if update is desired", mfParsed.Module.Mod.Path, reqModule, oldReplace.New.Path)
			}
		} else {
			// does not contain a replace statement. Insert it
			loggerStr = fmt.Sprintf("Inserting replace: Module: %s : %s => %s", mfParsed.Module.Mod.Path, reqModule, localPath)
			mfParsed.AddReplace(reqModule, "", localPath, "")
		}
		if rc.Verbose {
			rc.logger.Sugar().Info(loggerStr)
		}

	}
	module.moduleContents, err = mfParsed.Format()
	if err != nil {
		return err
	}

	return nil
}

// Identifies if a replace statement already exists for a given module name
func containsReplace(replaceStatments []*modfile.Replace, modName string) (*modfile.Replace, bool) {
	for _, repStatement := range replaceStatments {
		if repStatement.Old.Path == modName {
			return repStatement, true
		}
	}
	return nil, false
}
