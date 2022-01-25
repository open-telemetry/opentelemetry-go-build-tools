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

	graph, err := buildDepedencyGraph(rootPath)
	if err != nil {
		// unsure if we should return the errors up and out or panic here
		panic(fmt.Sprintf("failed to build dependency graph: %v", err))
	}

	for _, moduleInfo := range graph {
		err = insertReplace(rootPath, &moduleInfo)
		if err != nil {
			panic(fmt.Sprintf("failed to insert replace statements: %v", err))
		}

		err = pruneReplace(rootPath, &moduleInfo)

		if err != nil {
			panic(fmt.Sprintf("error pruning replace statements: %v", err))
		}

		err = writeModules(rootPath, moduleInfo)
		if err != nil {
			panic(fmt.Sprintf("error writing go.mod files: %v", err))
		}

	}

}

type moduleInfo struct {
	moduleFilePath string
	moduleContents []byte
	// should probably be a set for easy access
	requiredReplaceStatements map[string]struct{}
}

func newModuleInfo() *moduleInfo {
	var mi moduleInfo
	mi.requiredReplaceStatements = make(map[string]struct{})
	return &mi
}

func buildDepedencyGraph(rootPath string) (map[string]moduleInfo, error) {
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
	err := filepath.Walk(rootPath, goModFunc)
	if err != nil {
		fmt.Printf("error walking root directory: %v", err)
	}

	// identify and read the root module
	rootModPath := filepath.Join(rootPath, "go.mod")
	rootModFile, err := os.ReadFile(rootModPath)
	if err != nil {
		return nil, err
	}
	rootModule := modfile.ModulePath(rootModFile)

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
		// populate initial list of requirements
		for _, req := range mfParsed.Require {
			// store all modules requirements for use when pruning
			// do not add current module
			if strings.Contains(req.Mod.Path, rootModule) && req.Mod.Path != mfParsed.Module.Mod.Path {
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
					if _, ok := alreadyInsertedRepSet[transReq.Mod.Path]; transReq.Mod.Path != mfParsed.Module.Mod.Path &&
						strings.Contains(transReq.Mod.Path, rootModule) && !ok {
						reqStack = append(reqStack, transReq.Mod.Path)
						alreadyInsertedRepSet[transReq.Mod.Path] = struct{}{}
					}
				}
			}

		}
	}
	return moduleMap, nil
}

func insertReplace(rootPath string, module *moduleInfo) error {

	// modfile type that we will work with then write to the mod file in the end
	mfParsed, err := modfile.Parse("go.mod", module.moduleContents, nil)
	if err != nil {
		return err
	}

	for reqModule, _ := range module.requiredReplaceStatements {
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

func pruneReplace(rootPath string, module *moduleInfo) error {

	mfParsed, err := modfile.Parse("go.mod", module.moduleContents, nil)
	if err != nil {
		return err
	}

	rootModPath := filepath.Join(rootPath, "go.mod")
	rootModFile, err := os.ReadFile(rootModPath)
	if err != nil {
		return err
	}
	rootModule := modfile.ModulePath(rootModFile)

	// prune
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
		if _, ok := module.requiredReplaceStatements[rep.Old.Path]; strings.Contains(rep.Old.Path, rootModule) && !ok {
			mfParsed.DropReplace(rep.Old.Path, rep.Old.Version)
		}
	}
	module.moduleContents, err = mfParsed.Format()
	if err != nil {
		return err
	}

	return nil

}

func writeModules(rootPath string, module moduleInfo) error {

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
