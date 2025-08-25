// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package internal

import (
	"errors"
	"fmt"
	"os"
	"path"
	"slices"
	"strings"

	"github.com/kaptinlin/jsonschema"
)

// ReadJSONSchema returns a JSON schema from the config.schema.json file under the folder
func ReadJSONSchema(folder string) (*jsonschema.Schema, error) {
	schemaFile := path.Join(folder, "config.schema.json")
	b, err := os.ReadFile(schemaFile) //nolint: gosec
	if err != nil {
		return nil, err
	}
	compiler := jsonschema.NewCompiler()
	return compiler.Compile(b)
}

// DeriveSchema interprets the config struct to return a valid JSON schema
func DeriveSchema(cfgStruct APIstruct, structs []APIstruct, typeMappings map[string]string) (*jsonschema.Schema, error) {
	return createObject(parseObject(cfgStruct, structs, typeMappings)), nil
}

func matchingStruct(structRef string, structs []APIstruct) APIstruct {
	var selected APIstruct
	parts := strings.Split(structRef, ".")
	shortStructRef := parts[len(parts)-1]

	for _, s := range structs {
		if s.Name == shortStructRef {
			selected = s
			break
		}
	}
	return selected
}

func parseObject(s APIstruct, structs []APIstruct, typeMappings map[string]string) []any {
	var fields []any
	for _, f := range s.Fields {
		if f.Tag == "" {
			continue
		}
		tagValue := readTag(f.Tag)
		if tagValue == ",squash" {
			if found, ok := typeMappings[f.Type]; ok {
				fields = append(fields, jsonschema.Ref(found))
			} else {
				selected := matchingStruct(f.Type, structs)
				fields = append(fields, parseObject(selected, structs, typeMappings)...)
			}
		} else {
			if found, ok := typeMappings[f.Type]; ok {
				fields = append(fields, jsonschema.Prop(tagValue, jsonschema.Ref(found)))
			} else {
				fields = append(fields, jsonschema.Prop(tagValue, fieldTypeToJSONType(f.Type, structs, typeMappings)))
			}
		}
	}
	return fields
}

func fieldTypeToJSONType(fieldType string, structs []APIstruct, typeMappings map[string]string) *jsonschema.Schema {
	switch fieldType {
	case "string":
		return jsonschema.String()
	case "bool":
		return jsonschema.Boolean()
	case "int":
		return jsonschema.Integer()
	case "[]string":
		return jsonschema.Array(jsonschema.Items(jsonschema.String()))
	default:
		selected := matchingStruct(fieldType, structs)
		return createObject(parseObject(selected, structs, typeMappings))
	}
}

func createObject(fields []any) *jsonschema.Schema {
	obj := jsonschema.Object(fields...)
	var refs []*jsonschema.Schema
	for _, f := range fields {
		if field, ok := f.(*jsonschema.Schema); ok {
			refs = append(refs, field)
		}
	}
	obj.AllOf = refs
	return obj
}

func readTag(tag string) string {
	cut, _ := strings.CutSuffix(tag, "\"`")
	cut, _ = strings.CutPrefix(cut, "`mapstructure:\"")
	cut, _ = strings.CutSuffix(cut, ",omitempty")
	return cut
}

// CompareJSONSchema compares the presence of fields and their types.
func CompareJSONSchema(folder string, before *jsonschema.Schema, after *jsonschema.Schema) error {
	if before.Properties == nil && after.Properties == nil {
		return nil
	}
	if before.Properties == nil || after.Properties == nil || len(*before.Properties) != len(*after.Properties) {
		return fmt.Errorf("[%s] number of fields differ", folder)
	}
	return compareProperties(folder, before, after)
}

func compareProperties(folder string, before *jsonschema.Schema, after *jsonschema.Schema) error {
	var errs []error
	if before.Properties != nil {
		for name, bs := range *before.Properties {
			as, ok := (*after.Properties)[name]
			if !ok {
				errs = append(errs, fmt.Errorf("[%s] field %q is missing", folder, name))
			} else {
				if !slices.Equal(bs.Type, as.Type) {
					errs = append(errs, fmt.Errorf("[%s] field %q type changed", folder, name))
				}
				if bs.Ref != as.Ref {
					errs = append(errs, fmt.Errorf("[%s] field %q ref changed", folder, name))
				}
				if bs.Properties != nil {
					for subName, subBs := range *bs.Properties {
						subAs, ok := (*as.Properties)[subName]
						if !ok {
							errs = append(errs, fmt.Errorf("[%s] roperty %q is missing", folder, subName))
						} else {
							errs = append(errs, compareProperties(folder, subBs, subAs))
						}
					}
				}
			}
		}
	}
	if len(before.AllOf) != len(after.AllOf) {
		errs = append(errs, fmt.Errorf("[%s] references length do not match %d %d", folder, len(before.AllOf), len(after.AllOf)))
	}
	for i, b := range before.AllOf {
		a := after.AllOf[i]
		if a.Ref != b.Ref {
			errs = append(errs, fmt.Errorf("[%s] references do not match %q %q", folder, a.Ref, b.Ref))
		}
	}
	return errors.Join(errs...)
}
