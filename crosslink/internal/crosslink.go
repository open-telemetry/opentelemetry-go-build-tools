package crosslink

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	tools "go.opentelemetry.io/build-tools"
	"golang.org/x/mod/modfile"
)

// TODO: Print warning if there are modules that were not found in the repository but there are requirements for them in the dependency tree.
// TODO: Logging to alert user what changes have been made what attempts were made

type moduleInfo struct {
	moduleFilePath            string
	moduleContents            []byte
	requiredReplaceStatements map[string]struct{}
}

type runConfig struct {
	rootPath string
	verbose  bool
	// TODO: callout excluded path should be original module name not replaced module name. aka go.opentelemetry.io not ../replace
	excludedPaths map[string]struct{}
	overwrite     bool
	prune         bool
}

func newModuleInfo() *moduleInfo {
	var mi moduleInfo
	mi.requiredReplaceStatements = make(map[string]struct{})
	return &mi
}

func Crosslink(rc runConfig) {
	var err error
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
		// do not do anything with excluded
		// TODO: Readdress what `excluded` means more concretely. Should crosslink ignore
		// all references to excluded module or only for replacing and pruning?
		// If an exluded module is named should that stop crosslink from making edits to that go.mod file?
		//if _, exists := rc.excludedPaths[modName]; exists {
		//	continue
		//}

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
			panic(fmt.Sprintf("error writing go.mod files: %v", err))
		}
	}
}

func insertReplace(module *moduleInfo, rc runConfig) error {
	// modfile type that we will work with then write to the mod file in the end
	mfParsed, err := modfile.Parse("go.mod", module.moduleContents, nil)
	if err != nil {
		return err
	}

	for reqModule := range module.requiredReplaceStatements {
		// skip excluded
		if _, exists := rc.excludedPaths[reqModule]; exists {
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

		// see if replace statement already exists for module. Verify if it's the same. If it does not exist then add it.
		// AddReplace should handle all of these conditions in terms of add and/or verifying
		// https://cs.opensource.google/go/go/+/master:src/cmd/vendor/golang.org/x/mod/modfile/rule.go;l=1296?q=addReplace
		if _, exists := containsReplace(mfParsed.Replace, reqModule); exists {
			if rc.overwrite {
				// TODO: log overwriting
				mfParsed.AddReplace(reqModule, "", localPath, "")
			} else {
				// TODO: log cannot replace
			}
		} else {
			// does not contain a replace statement. Insert it
			// TODO:log insert
			mfParsed.AddReplace(reqModule, "", localPath, "")
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

func pruneReplace(rootModulePath string, module *moduleInfo, rc runConfig) error {
	mfParsed, err := modfile.Parse("go.mod", module.moduleContents, nil)
	if err != nil {
		return err
	}

	// check to see if its intra dependency and no longer presenent
	for _, rep := range mfParsed.Replace {
		// skip excluded
		if _, exists := rc.excludedPaths[rep.Old.Path]; exists {
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
			mfParsed.DropReplace(rep.Old.Path, rep.Old.Version)
		}
	}
	module.moduleContents, err = mfParsed.Format()
	if err != nil {
		return err
	}

	return nil
}

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
