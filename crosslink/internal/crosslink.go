package crosslink

import (
	"fmt"
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	tools "go.opentelemetry.io/build-tools"
	"golang.org/x/mod/modfile"
)

// TODO: Print warning if there are modules that were not found in the repository but there are requirements for them in the dependency tree.
// TODO: Conig for nondestructive vs destructive changes.
// TODO: Logging to alert user what changes have been made what attempts were made

type moduleInfo struct {
	moduleFilePath            string
	moduleContents            []byte
	requiredReplaceStatements map[string]struct{}
}

type runConfig struct {
	rootPath      string
	verbose       bool
	excludedPaths []string
	overwrite     bool
	prune         bool
}

func newModuleInfo() *moduleInfo {
	var mi moduleInfo
	mi.requiredReplaceStatements = make(map[string]struct{})
	return &mi
}

func Crosslink(rootPath string) {
	var err error
	if rootPath == "" {
		rootPath, err = tools.FindRepoRoot()
		if err != nil {
			panic("Could not find repo root directory")
		}
	}

	if _, err := os.Stat(filepath.Join(rootPath, "go.mod")); err != nil {
		panic("Invalid root directory, could not locate go.mod file")
	}

	// identify and read the root module
	rootModPath := filepath.Join(rootPath, "go.mod")
	rootModFile, err := os.ReadFile(rootModPath)
	if err != nil {
		panic(fmt.Sprintf("Could not read go.mod file in root path: %v", err))
	}
	rootModulePath := modfile.ModulePath(rootModFile)

	graph, err := buildDepedencyGraph(rootPath, rootModulePath)
	if err != nil {
		// unsure if we should return the errors up and out or panic here
		panic(fmt.Sprintf("failed to build dependency graph: %v", err))
	}

	for _, moduleInfo := range graph {
		err = insertReplace(&moduleInfo)
		if err != nil {
			panic(fmt.Sprintf("failed to insert replace statements: %v", err))
		}

		err = pruneReplace(rootModulePath, &moduleInfo)

		if err != nil {
			panic(fmt.Sprintf("error pruning replace statements: %v", err))
		}

		err = writeModules(moduleInfo)
		if err != nil {
			panic(fmt.Sprintf("error writing go.mod files: %v", err))
		}
	}
}

func buildDepedencyGraph(rootDir string, rootModulePath string) (map[string]moduleInfo, error) {
	moduleMap := make(map[string]moduleInfo)
	goModFunc := func(filePath string, info fs.FileInfo, err error) error {
		if err != nil {
			fmt.Printf("Warning: file could not be read during filepath.Walk: %v", err)
			return nil
		}

		if filepath.Base(filePath) == "go.mod" {
			modFile, err := ioutil.ReadFile(filePath)
			if err != nil {
				return err
			}

			modInfo := newModuleInfo()
			modInfo.moduleContents = modFile
			modInfo.moduleFilePath = filePath

			moduleMap[modfile.ModulePath(modFile)] = *modInfo
		}
		return nil
	}
	err := filepath.Walk(rootDir, goModFunc)
	if err != nil {
		fmt.Printf("error walking root directory: %v", err)
	}

	for _, modInfo := range moduleMap {
		// reqStack contains a list of module paths that are required to have local replace statements
		// reqStack should only contain intra-repository modules
		reqStack := make([]string, 0)
		alreadyInsertedRepSet := make(map[string]struct{})

		// modfile type that we will work with then write to the mod file in the end
		mfParsed, err := modfile.Parse("go.mod", modInfo.moduleContents, nil)
		if err != nil {
			return nil, err
		}

		// NOTE: when adding to the stack or writing the replace statements I do not verify that the module exists in the local repository path.
		// I believe this check should be done to avoid inserting replace statements to local directories that do not exist.
		// This should maybe be a warning to the user that the replace statement could not be made because the
		// local repository does not exist in the path.
		// TODO: Add test case for this
		// populate initial list of requirements
		// Modules should only be queued for replacement if they meet the following criteria
		// 1. They exist within the set of go.mod files discovered during the filepath walk
		//		- This prevents uneccessary or erroneous replace statements from being added.
		//		- Crosslink will not make an assumption that a module exists even though it falls under the module path.
		// 2. They fall under the module path of the root module
		// 3. They are not the same module that we are currently working with.
		for _, req := range mfParsed.Require {
			if _, existsInPath := moduleMap[req.Mod.Path]; strings.Contains(req.Mod.Path, rootModulePath) &&
				req.Mod.Path != mfParsed.Module.Mod.Path && existsInPath {
				reqStack = append(reqStack, req.Mod.Path)
				alreadyInsertedRepSet[req.Mod.Path] = struct{}{}
			}
		}

		// iterate through stack adding replace directives and transitive requirements as needed
		// if the replace directive already exists for the module path then ensure that it is pointing to the correct location
		for len(reqStack) > 0 {
			var reqModule string
			reqModule, reqStack = reqStack[len(reqStack)-1], reqStack[:len(reqStack)-1]
			modInfo.requiredReplaceStatements[reqModule] = struct{}{}

			// now find all transitive dependencies for the current required module. Only add to stack if they
			// have not already been added and they are not the current module we are working in.
			if value, ok := moduleMap[reqModule]; ok {
				m, err := modfile.Parse("go.mod", value.moduleContents, nil)
				if err != nil {
					return nil, err
				}
				for _, transReq := range m.Require {
					_, existsInPath := moduleMap[transReq.Mod.Path]
					_, alreadyInserted := alreadyInsertedRepSet[transReq.Mod.Path]
					if transReq.Mod.Path != mfParsed.Module.Mod.Path &&
						strings.Contains(transReq.Mod.Path, rootModulePath) &&
						!alreadyInserted && existsInPath {
						reqStack = append(reqStack, transReq.Mod.Path)
						alreadyInsertedRepSet[transReq.Mod.Path] = struct{}{}
					}
				}
			}

		}
	}
	return moduleMap, nil
}

func insertReplace(module *moduleInfo) error {
	// modfile type that we will work with then write to the mod file in the end
	mfParsed, err := modfile.Parse("go.mod", module.moduleContents, nil)
	if err != nil {
		return err
	}

	for reqModule := range module.requiredReplaceStatements {
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
		mfParsed.AddReplace(reqModule, "", localPath, "")
	}
	module.moduleContents, err = mfParsed.Format()
	if err != nil {
		return err
	}

	return nil
}

func pruneReplace(rootModulePath string, module *moduleInfo) error {
	mfParsed, err := modfile.Parse("go.mod", module.moduleContents, nil)
	if err != nil {
		return err
	}

	// check to see if its intra dependency and no longer presenent
	for _, rep := range mfParsed.Replace {
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
