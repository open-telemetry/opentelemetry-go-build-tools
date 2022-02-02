package crosslink

import (
	"fmt"
	"strings"

	"golang.org/x/mod/modfile"
)

func Prune(rc runConfig) {
	defer rc.logger.Sync()
	var err error

	rootModulePath, err := identifyRootModule(rc.RootPath)
	if err != nil {
		panic(fmt.Sprintf("failed to identify root module: %v", err))
	}

	graph, err := buildDepedencyGraph(rc, rootModulePath)
	if err != nil {
		panic(fmt.Sprintf("failed to build dependency graph: %v", err))
	}

	for _, moduleInfo := range graph {
		err = pruneReplace(rootModulePath, &moduleInfo, rc)

		if err != nil {
			panic(fmt.Sprintf("error pruning replace statements: %v", err))
		}

		err = writeModules(moduleInfo)
		if err != nil {
			panic(fmt.Sprintf("error writing go.mod files: %v", err))
		}
	}
}

func pruneReplace(rootModulePath string, module *moduleInfo, rc runConfig) error {
	mfParsed, err := modfile.Parse("go.mod", module.moduleContents, nil)
	if err != nil {
		return err
	}

	// check to see if its intra dependency and no longer present
	for _, rep := range mfParsed.Replace {
		// skip excluded
		if _, exists := rc.ExcludedPaths[rep.Old.Path]; exists {
			if rc.Verbose {
				rc.logger.Sugar().Infof("Excluded Module %s, ignoring prune", rep.Old.Path)
			}
			continue
		}

		if _, ok := module.requiredReplaceStatements[rep.Old.Path]; strings.Contains(rep.Old.Path, rootModulePath) && !ok {
			if rc.Verbose {
				rc.logger.Sugar().Infof("Pruning replace statement: Module %s: %s => %s", mfParsed.Module.Mod.Path, rep.Old.Path, rep.New.Path)
			}
			mfParsed.DropReplace(rep.Old.Path, rep.Old.Version)
		}
	}
	module.moduleContents, err = mfParsed.Format()
	if err != nil {
		return err
	}

	return nil
}
