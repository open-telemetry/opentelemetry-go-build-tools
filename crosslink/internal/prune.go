package crosslink

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	tools "go.opentelemetry.io/build-tools"
	"golang.org/x/mod/modfile"
)

func Prune(rc runConfig) {
	defer rc.logger.Sync()
	var err error
	if rc.RootPath == "" {
		rc.RootPath, err = tools.FindRepoRoot()
		if err != nil {
			panic("Could not find repo root directory")
		}
	}

	if _, err := os.Stat(filepath.Join(rc.RootPath, "go.mod")); err != nil {
		panic("Invalid root directory, could not locate go.mod file")
	}

	// identify and read the root module
	rootModPath := filepath.Join(rc.RootPath, "go.mod")
	rootModFile, err := os.ReadFile(rootModPath)
	if err != nil {
		panic(fmt.Sprintf("Could not read go.mod file in root path: %v", err))
	}
	rootModulePath := modfile.ModulePath(rootModFile)

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
