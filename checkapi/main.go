// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

// Checkapi is a tool to check the API of OpenTelemetry Go modules.
package main

import (
	"errors"
	"flag"
	"fmt"
	"go/ast"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strings"
	"unicode"

	"github.com/kaptinlin/jsonschema"

	"go.opentelemetry.io/build-tools/checkapi/internal"

	"gopkg.in/yaml.v3"
)

func main() {
	folder := flag.String("folder", ".", "folder investigated for modules")
	configPath := flag.String("config", "cmd/checkapi/config.yaml", "configuration file")
	flag.Parse()
	if err := run(*folder, *configPath); err != nil {
		log.Fatal(err)
	}
}

func run(folder string, configPath string) error {
	configData, err := os.ReadFile(configPath) // #nosec G304
	if err != nil {
		return err
	}
	var cfg internal.Config
	err = yaml.Unmarshal(configData, &cfg)
	if err != nil {
		return err
	}
	var errs []error
	err = filepath.Walk(folder, func(path string, info fs.FileInfo, _ error) error {
		if info.Name() == "go.mod" {
			base := filepath.Dir(path)
			relativeBase, err2 := filepath.Rel(folder, base)
			if err2 != nil {
				return err2
			}
			// no code paths under internal need to be inspected
			if strings.HasPrefix(relativeBase, "internal") {
				return nil
			}

			for _, a := range cfg.IgnoredPaths {
				if filepath.Join(filepath.SplitList(a)...) == relativeBase {
					fmt.Printf("Ignoring %s per denylist\n", base)
					return nil
				}
			}
			componentType, err3 := internal.ReadComponentType(base)
			if err3 != nil {
				return err3
			}
			if err = walkFolder(cfg, base, componentType); err != nil {
				errs = append(errs, err)
			}
		}
		return nil
	})
	if err != nil {
		return err
	}
	if len(errs) > 0 {
		return errors.Join(errs...)
	}
	return nil
}

func walkFolder(cfg internal.Config, folder string, componentType string) error {
	result, err := internal.Read(folder, cfg.IgnoredFunctions, cfg.ExcludedFiles)
	if err != nil {
		return err
	}

	sort.Slice(result.Structs, func(i int, j int) bool {
		return strings.Compare(result.Structs[i].Name, result.Structs[j].Name) > 0
	})
	sort.Strings(result.Values)
	sort.Slice(result.Functions, func(i int, j int) bool {
		return strings.Compare(result.Functions[i].Name, result.Functions[j].Name) < 0
	})
	fnNames := make([]string, len(result.Functions))
	for i, fn := range result.Functions {
		fnNames[i] = fn.Name
	}
	if len(result.Structs) == 0 && len(result.Values) == 0 && len(result.Functions) == 0 {
		// nothing to validate, return
		return nil
	}

	var errs []error

	if len(cfg.AllowedFunctions) > 0 {

		functionsPresent := map[string]struct{}{}
	OUTER:
		for _, fnDesc := range cfg.AllowedFunctions {
			if !slices.Contains(fnDesc.Classes, componentType) {
				continue
			}
			for _, fn := range result.Functions {
				if fn.Name == fnDesc.Name &&
					slices.Equal(fn.Params, fnDesc.Parameters) &&
					slices.Equal(fn.ReturnTypes, fnDesc.ReturnTypes) {
					functionsPresent[fn.Name] = struct{}{}
					break OUTER
				}
			}
		}

		if len(functionsPresent) == 0 {
			errs = append(errs, fmt.Errorf("[%s] no function matching configuration found", folder))
		}
	}

	if cfg.UnkeyedLiteral.Enabled {
		for _, s := range result.Structs {
			if err := checkStructDisallowUnkeyedLiteral(cfg, s, folder); err != nil {
				errs = append(errs, err)
			}
		}
	}

	if (!cfg.JSONSchema.CheckPresent && !cfg.JSONSchema.CheckValid && !cfg.ComponentAPI && !cfg.ComponentAPIStrict) || (componentType != "connector" && componentType != "exporter" && componentType != "extension" && componentType != "processor" && componentType != "receiver") {
		return errors.Join(errs...)
	}

	if result.ConfigStructName == "" && cfg.ComponentAPIStrict {
		errs = append(errs, fmt.Errorf("[%s] cannot find the createDefaultConfig function", folder))
		return errors.Join(errs...)
	}

	var cfgStruct *internal.APIstruct
	allStructs := make(map[string]struct{}, len(result.Structs))
	structsByName := make(map[string]internal.APIstruct, len(result.Structs))
	for _, s := range result.Structs {
		if ast.IsExported(s.Name) {
			allStructs[s.Name] = struct{}{}
			structsByName[s.Name] = s
		}
		if s.Name == result.ConfigStructName {
			cfgStruct = &s
		}
	}
	if cfgStruct == nil {
		if cfg.ComponentAPIStrict {
			errs = append(errs, fmt.Errorf("[%s] cannot find the config struct", folder))
		}
		return errors.Join(errs...)
	}

	if _, err := os.Stat(filepath.Join(folder, "config.schema.json")); errors.Is(err, os.ErrNotExist) {
		if cfg.JSONSchema.CheckPresent {
			errs = append(errs, err)
		}
	} else {
		jsonSchema, err := internal.ReadJSONSchema(folder)
		if err != nil {
			errs = append(errs, err)
		} else {
			var after *jsonschema.Schema
			if after, err = internal.DeriveSchema(*cfgStruct, result.Structs, cfg.JSONSchema.TypeMappings); err != nil {
				errs = append(errs, err)
			} else if err := internal.CompareJSONSchema(folder, jsonSchema, after); err != nil {
				errs = append(errs, err)
				b, _ := after.MarshalJSON()
				errs = append(errs, fmt.Errorf("[%s] new JSON schema: %s", folder, string(b)))
			}

		}
	}

	if cfg.ComponentAPIStrict || cfg.ComponentAPI {
		delete(allStructs, cfgStruct.Name)
		filterStructs(structsByName, *cfgStruct, allStructs)
		for k, v := range structsByName {
			if v.Internal {
				delete(allStructs, k)
			}
		}
		if len(allStructs) > 0 {
			structNames := make([]string, 0, len(allStructs))
			for k := range allStructs {
				structNames = append(structNames, k)
			}
			errs = append(errs, fmt.Errorf("[%s] these structs are not part of config and cannot be exported: %s", folder, strings.Join(structNames, ",")))
		}
	}

	return errors.Join(errs...)
}

func filterStructs(structMap map[string]internal.APIstruct, current internal.APIstruct, allStructs map[string]struct{}) {
	for _, f := range current.Fields {
		if s, ok := structMap[f.Type]; ok {
			delete(allStructs, s.Name)
			filterStructs(structMap, s, allStructs)
		}
	}
}

func checkStructDisallowUnkeyedLiteral(cfg internal.Config, s internal.APIstruct, folder string) error {
	if !unicode.IsUpper(rune(s.Name[0])) {
		return nil
	}
	if len(s.Fields) > cfg.UnkeyedLiteral.Limit {
		return nil
	}
	if len(s.Fields) == 0 {
		return nil
	}

	for _, f := range s.Fields {
		if len(f.Name) == 0 {
			if !unicode.IsUpper(rune(f.Type[0])) {
				return nil
			}
		} else {
			if !unicode.IsUpper(rune(f.Name[0])) {
				return nil
			}
		}
	}
	return fmt.Errorf("%s struct %q does not prevent unkeyed literal initialization", folder, s.Name)
}
