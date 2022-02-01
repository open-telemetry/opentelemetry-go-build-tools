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
	defer rc.logger.Sync()

	if rc.rootPath == "" {
		rc.rootPath, err = tools.FindRepoRoot()
		if err != nil {
			panic("Could not find repo root directory")
		}
	}

	if _, err := os.Stat(filepath.Join(rc.rootPath, "go.mod")); err != nil {
		panic("Invalid root directory, could not locate go.mod file")
	}

	// identify and read the root module
	rootModPath := filepath.Join(rc.rootPath, "go.mod")
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

	// check to see if its intra dependency and no longer presenent
	for _, rep := range mfParsed.Replace {
		// skip excluded
		if _, exists := rc.excludedPaths[rep.Old.Path]; exists {
			if rc.verbose {
				rc.logger.Sugar().Infof("Excluded Module %s, ignoring prune", rep.Old.Path)
			}
			continue
		}

		// THOUGHTS ON NAMING CONVENTION REQ:
		// will this cause errors for modules that do not conform to naming conventions?
		// this may unintentially drop replace statements
		// will go mod tidy remove replace statements for you?
		// if not I would want to see if replace is not in the requirements or required replace statements
		// I believe checking to make sure it's not in the requirements also would alleviate the issue.
		// Even with the k,v store in mod info does that account for inter-repository replacements. Do those
		// require transitive replacements that we would drop? This could get messy if we don't enforce the naming convention.
		// IF IT IS INTRA REPOSITORY (ID'D BY REQ'D REPLACE STATEMENT) AND ITS NOT IN REQUIRED MODULES KV STORE == REMOVE
		//		This doesn't account for inter repository transitive dependencies on the local machine.

		if _, ok := module.requiredReplaceStatements[rep.Old.Path]; strings.Contains(rep.Old.Path, rootModulePath) && !ok {
			if rc.verbose {
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
