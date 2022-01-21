package crosslink

import (
	"container/list"
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
	//validate root
	if _, err := os.Stat(filepath.Join(rootPath, "go.mod")); err != nil {
		panic("Invalid root directory, could not locate go.mod file")
	}
	executeCrosslink(rootPath)

}

type Set struct {
	list map[string]struct{}
}

func (s *Set) Has(v string) bool {
	_, ok := s.list[v]
	return ok
}

func (s *Set) Add(v string) {
	s.list[v] = struct{}{}
}

func (s *Set) Remove(v string) {
	delete(s.list, v)
}

func (s *Set) Clear() {
	s.list = make(map[string]struct{})
}

func (s *Set) Size() int {
	return len(s.list)
}

func NewSet() *Set {
	s := &Set{}
	s.list = make(map[string]struct{})
	return s
}

type moduleInfo struct {
	moduleFilePath string
	moduleContents []byte
	//could possibly add some type of caching for transitive requirements
}

func executeCrosslink(rootPath string) error {
	// module map is a key,value store of module name -> moduleInfo struct
	moduleMap := make(map[string]moduleInfo)

	goModFunc := func(filePath string, info fs.FileInfo, err error) error {
		if err != nil {
			fmt.Printf("Warning: file could not be read during filepath.Walk: %v", err)
			return nil
		}

		if filepath.Base(filePath) == "go.mod" {
			//read file
			modFile, err := ioutil.ReadFile(filePath)
			if err != nil {
				return err
			}

			moduleMap[modfile.ModulePath(modFile)] = moduleInfo{moduleFilePath: filePath, moduleContents: modFile}
		}
		return nil
	}

	err := filepath.Walk(rootPath, goModFunc)
	if err != nil {
		fmt.Printf("error walking root directory: %v", err)
	}

	// identify what the root module is
	rootModPath := filepath.Join(rootPath, "go.mod")
	rootModFile, err := os.ReadFile(rootModPath)
	if err != nil {
		return err
	}
	rootModule := modfile.ModulePath(rootModFile)

	for moduleName, modInfo := range moduleMap {
		// reqStack contains a list of module paths that are required to have local replace statements
		// reqStack should only contain intra-repository modules
		reqStack := list.New()
		// set is type map[string]struct{} and has standard set capabilties.
		alreadyInsertedRepSet := NewSet()

		// modfile type that we will work with then write to the mod file in the end
		mfParsed, err := modfile.Parse("go.mod", modInfo.moduleContents, nil)
		if err != nil {
			// fmt.Printf("Error parsing go.mod file: %v", err)
			return err
		}
		// populate initial list of requirements
		for _, req := range mfParsed.Require {
			if strings.Contains(req.Mod.Path, rootModule) {
				reqStack.PushBack(req.Mod.Path)
				alreadyInsertedRepSet.Add(req.Mod.Path)
			}
		}

		// iterate through stack adding replace directives and transitive requirements as needed
		// if the replace directive already exists for the module path then ensure that it is pointing to the write location
		for reqStack.Len() > 0 {
			reqModule := reqStack.Front().Value.(string)
			reqStack.Remove(reqStack.Front())

			if err != nil {
				return err
			}

			localPath, err := filepath.Rel(moduleName, reqModule)
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
			// now prune any leftover intra-repository replace statements that are not required
			// I'm making an assumption here that we will always be leaving the version number blank.
			// Is this a valid assumption or do I need to handle this?
			// This may be a valid assumption for the current state otel but is this going to limit the
			// tools use for outside of OTEL or even future states of OTEL?
			mfParsed.AddReplace(reqModule, "", localPath, "")

			// now find all transitive dependencies for the current required module. Only add to stack if they
			// have not already been added.
			if value, ok := moduleMap[reqModule]; ok {
				m, err := modfile.Parse("go.mod", value.moduleContents, nil)
				if err != nil {
					return err
				}
				for _, transReq := range m.Require {
					if strings.Contains(transReq.Mod.Path, rootModule) && !alreadyInsertedRepSet.Has(transReq.Mod.Path) {
						reqStack.PushBack(transReq.Mod.Path)
						alreadyInsertedRepSet.Add(transReq.Mod.Path)
					}
				}
			}

		}
		// more versioning assumptions
		for _, rep := range mfParsed.Replace {
			// check to see if its intra dependency and its not in our inserted set
			if strings.Contains(rep.Old.Path, rootModule) && !alreadyInsertedRepSet.Has(rep.Old.Path) {
				mfParsed.DropReplace(rep.Old.Path, rep.Old.Version)
			}
		}

		mfParsed.Cleanup()

		//  now overwrite the existing gomod file
		gomodFile, err := mfParsed.Format()
		if err != nil {
			return err
		}
		//write our updated go.mod file
		err = os.WriteFile(modInfo.moduleFilePath, gomodFile, 0700)
		if err != nil {
			return err
		}

	}

	return nil
}
