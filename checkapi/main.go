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

	"github.com/goccy/go-json"
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
			metadata, found, err3 := internal.ReadMetadata(base)
			if err3 != nil {
				return err3
			}
			if !found {
				return nil
			}
			if err = walkFolder(cfg, base, metadata); err != nil {
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

func walkFolder(cfg internal.Config, folder string, metadata internal.Metadata) error {
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
	componentType := metadata.Status.Class
	isFactoryComponent := componentType == "connector" || componentType == "exporter" || componentType == "extension" || componentType == "processor" || componentType == "receiver"

	if len(cfg.AllowedFunctions) > 0 {

		functionsPresent := map[string]struct{}{}
	OUTER:
		for _, fnDesc := range cfg.AllowedFunctions {
			if !slices.Contains(fnDesc.Classes, metadata.Status.Class) {
				continue
			}
			// any function
			if fnDesc.Name == "*" {
				functionsPresent[""] = struct{}{}
				break OUTER
			}
			// no functions at all.
			if fnDesc.Name == "" {
				functionsPresent[""] = struct{}{}
				fnNames = make([]string, 0, len(result.Functions))
				for i, fn := range result.Functions {
					if !fn.Internal {
						fnNames[i] = fn.Name
					}
				}
				if len(fnNames) > 0 {
					errs = append(errs, fmt.Errorf("[%s] no functions must be exported under this module, found %q", folder, strings.Join(fnNames, ",")))
				}
				break OUTER
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

		if len(functionsPresent) == 0 && isFactoryComponent {
			errs = append(errs, fmt.Errorf("[%s] no function matching configuration found", folder))
		}
	}

	if cfg.UnkeyedLiteral.Enabled {
		for _, s := range result.Structs {
			if err = checkStructDisallowUnkeyedLiteral(cfg, s, folder); err != nil {
				errs = append(errs, err)
			}
		}
	}

	if (!cfg.JSONSchema.CheckPresent && !cfg.JSONSchema.CheckValid && !cfg.ComponentAPI && !cfg.ComponentAPIStrict) || !isFactoryComponent {
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
		segments := strings.Split(s.Name, ".")
		name := segments[len(segments)-1]
		if ast.IsExported(name) {
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

	if metadata.Config == nil {
		if cfg.JSONSchema.CheckPresent {
			errs = append(errs, err)
		}
	} else {
		configSchemaBytes, err := json.Marshal(metadata.Config)
		if err != nil {
			errs = append(errs, err)
		} else {
			configSchema, err := jsonschema.NewCompiler().Compile(configSchemaBytes)
			if err != nil {
				errs = append(errs, err)
			} else {
				var structDerivedSchema *jsonschema.Schema
				if structDerivedSchema, err = internal.DeriveSchema(*cfgStruct, result.Structs, cfg.JSONSchema.TypeMappings); err != nil {
					errs = append(errs, err)
				} else if err := internal.CompareJSONSchema(folder, configSchema, structDerivedSchema); err != nil {
					errs = append(errs, err)
					configSchemaBytes, _ := structDerivedSchema.MarshalJSON()
					rawJSON := map[string]any{}
					_ = json.Unmarshal(configSchemaBytes, &rawJSON)
					configSchemaYAML, _ := yaml.Marshal(rawJSON)
					errs = append(errs, fmt.Errorf("[%s] new JSON schema: %s", folder, string(configSchemaYAML)))
				}
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
	if s.Internal {
		return nil
	}
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
	return fmt.Errorf("[%s] struct %q does not prevent unkeyed literal initialization", folder, s.Name)
}
