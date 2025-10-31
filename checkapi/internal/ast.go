// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

// Package internal provides internal utilities for the checkapi package.
package internal

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// ExprToString converts an AST expression to a string representation.
func ExprToString(expr ast.Expr) string {
	switch e := expr.(type) {
	case *ast.MapType:
		return fmt.Sprintf("map[%s]%s", ExprToString(e.Key), ExprToString(e.Value))
	case *ast.ArrayType:
		return fmt.Sprintf("[%s]%s", ExprToString(e.Len), ExprToString(e.Elt))
	case *ast.StructType:
		var fields []string
		for _, f := range e.Fields.List {
			fields = append(fields, ExprToString(f.Type))
		}
		return fmt.Sprintf("{%s}", strings.Join(fields, ","))
	case *ast.InterfaceType:
		var methods []string
		for _, f := range e.Methods.List {
			methods = append(methods, "func "+ExprToString(f.Type))
		}
		return fmt.Sprintf("{%s}", strings.Join(methods, ","))
	case *ast.ChanType:
		return fmt.Sprintf("chan(%s)", ExprToString(e.Value))
	case *ast.FuncType:
		var results []string
		if e.Results != nil {
			for _, r := range e.Results.List {
				results = append(results, ExprToString(r.Type))
			}
		}
		var params []string
		if e.Params != nil {
			for _, r := range e.Params.List {
				params = append(params, ExprToString(r.Type))
			}
		}
		var typeParams []string
		if e.TypeParams != nil {
			for _, r := range e.TypeParams.List {
				typeParams = append(typeParams, ExprToString(r.Type))
			}
		}
		generics := ""
		if len(typeParams) > 0 {
			generics = fmt.Sprintf("[%s]", strings.Join(typeParams, ","))
		}
		if len(results) == 0 {
			return fmt.Sprintf("func%s(%s)", generics, strings.Join(params, ","))
		}
		return fmt.Sprintf("func%s(%s) %s", generics, strings.Join(params, ","), strings.Join(results, ","))
	case *ast.SelectorExpr:
		return fmt.Sprintf("%s.%s", ExprToString(e.X), e.Sel.Name)
	case *ast.Ident:
		return e.Name
	case nil:
		return ""
	case *ast.StarExpr:
		return fmt.Sprintf("*%s", ExprToString(e.X))
	case *ast.Ellipsis:
		return fmt.Sprintf("%s...", ExprToString(e.Elt))
	case *ast.IndexExpr:
		return fmt.Sprintf("%s[%s]", ExprToString(e.X), ExprToString(e.Index))
	case *ast.BasicLit:
		return e.Value
	case *ast.IndexListExpr:
		var exprs []string
		for _, e := range e.Indices {
			exprs = append(exprs, ExprToString(e))
		}
		return strings.Join(exprs, ",")
	case *ast.UnaryExpr:
		return fmt.Sprintf("%s%s", e.Op.String(), ExprToString(e.X))
	case *ast.BinaryExpr:
		return fmt.Sprintf("%s%s", e.Op.String(), ExprToString(e.X))
	case *ast.ParenExpr:
		return fmt.Sprintf("(%s)", ExprToString(e.X))
	default:
		panic(fmt.Sprintf("Unsupported expr type: %#v", expr))
	}
}

// Read reads the Go files in the specified folder and returns an API object.
func Read(folder string, ignoredFunctions []string, excludedFiles []string) (API, error) {
	result := &API{}
	set := token.NewFileSet()

	readErr := filepath.Walk(folder, func(path string, info fs.FileInfo, _ error) error {
		if info.IsDir() {

			isInternal := false
			for _, s := range strings.Split(path, string(os.PathSeparator)) {
				if s == "internal" {
					isInternal = true
				}
			}
			goModPresent := false
			if _, err := os.Stat(filepath.Join(path, "go.mod")); err == nil {
				goModPresent = true
			}
			// if a subfolder has its own go.mod, do not read the folder.
			if path != folder && goModPresent {
				return nil
			}
			packs, err := parser.ParseDir(set, path, nil, 0)
			if err != nil {
				return err
			}

			for _, pack := range packs {
				if err := readPackage(pack, ignoredFunctions, excludedFiles, result, isInternal, path == folder); err != nil {
					return err
				}
			}
		}
		return nil
	})

	return *result, readErr
}

func readPackage(pack *ast.Package, ignoredFunctions []string, excludedFiles []string, result *API, internal bool, root bool) error { // nolint:staticcheck // SA1019
FILE:
	for path, f := range pack.Files {
		for _, exclusionPattern := range excludedFiles {
			ok, err2 := filepath.Match(exclusionPattern, filepath.Base(path))
			if err2 != nil {
				return err2
			}
			if ok {
				continue FILE
			}
		}
		packageName := pack.Name
		if root {
			packageName = ""
		}
		readFile(ignoredFunctions, f, result, internal, packageName)
	}
	return nil
}

func interpretFieldType(f *ast.Field, expr ast.Expr) []APIstructField {
	var fieldNames []APIstructField
	fieldType := expr
	switch t := fieldType.(type) {
	case *ast.StarExpr:
		fieldType = t.X
	case *ast.ArrayType:
		fieldType = t.Elt
		if tt, ok := fieldType.(*ast.StarExpr); ok {
			fieldType = tt.X
		}
	case *ast.MapType:
		fieldType = t.Value
		if tt, ok := fieldType.(*ast.StarExpr); ok {
			fieldType = tt.X
		}
		fieldNames = interpretFieldType(f, t.Key)
	case *ast.Ident:
		// nothing to do
	case *ast.ChanType:
		fieldType = t.Value
	case *ast.StructType:
		// nothing to do
	case *ast.IndexExpr:
		fieldType = t.X
		fieldNames = interpretFieldType(f, t.Index)
	}
	tag := ""
	if f.Tag != nil {
		tag = f.Tag.Value
	}
	fieldNames = append(fieldNames, APIstructField{Name: f.Names[0].Name, Type: ExprToString(fieldType), Tag: tag})
	return fieldNames
}

func readFile(ignoredFunctions []string, f *ast.File, result *API, internal bool, packageName string) {
	for _, d := range f.Decls {
		if str, isStr := d.(*ast.GenDecl); isStr {
			for _, s := range str.Specs {
				if values, ok := s.(*ast.ValueSpec); ok {
					for _, v := range values.Names {
						if v.IsExported() {
							result.Values = append(result.Values, v.Name)
						}
					}
				}
				if t, ok := s.(*ast.TypeSpec); ok {
					if structType, ok := t.Type.(*ast.StructType); ok {
						var fieldNames []APIstructField
						if structType.Fields != nil {
							fieldNames = make([]APIstructField, 0, len(structType.Fields.List))
							for _, f := range structType.Fields.List {
								if len(f.Names) > 0 {
									fieldNames = append(fieldNames, interpretFieldType(f, f.Names[0].Obj.Decl.(*ast.Field).Type)...)
								} else {
									// Embedded struct
									fieldType := f.Type
									switch t := f.Type.(type) {
									case *ast.StarExpr:
										fieldType = t.X
									case *ast.ArrayType:
										fieldType = t.Elt
										if tt, ok := fieldType.(*ast.StarExpr); ok {
											fieldType = tt.X
										}
									}
									tag := ""
									if f.Tag != nil {
										tag = f.Tag.Value
									}
									fieldNames = append(fieldNames, APIstructField{Name: "", Type: ExprToString(fieldType), Tag: tag})
								}
							}
						}
						name := t.Name.String()
						if packageName != "" {
							name = fmt.Sprintf("%s.%s", packageName, t.Name)
						}
						result.Structs = append(result.Structs, APIstruct{
							Name:     name,
							Fields:   fieldNames,
							Internal: internal,
						})
					}
				}
			}
		}
		if fn, isFn := d.(*ast.FuncDecl); isFn {
			exported := false
			receiver := ""
			if fn.Recv.NumFields() == 0 && !isFunctionIgnored(ignoredFunctions, fn.Name.String()) {
				exported = true
			}
			if fn.Recv.NumFields() > 0 {
				for _, t := range fn.Recv.List {
					for _, n := range t.Names {
						exported = exported || n.IsExported()
						if n.IsExported() {
							receiver = n.Name
						}
					}
				}
			}
			if exported {
				var returnTypes []string
				if fn.Type.Results.NumFields() > 0 {
					for _, r := range fn.Type.Results.List {
						returnTypes = append(returnTypes, ExprToString(r.Type))
					}
				}
				var params []string
				if fn.Type.Params.NumFields() > 0 {
					for _, r := range fn.Type.Params.List {
						params = append(params, ExprToString(r.Type))
					}
				}
				var typeParams []string
				if fn.Type.TypeParams.NumFields() > 0 {
					for _, r := range fn.Type.TypeParams.List {
						typeParams = append(typeParams, ExprToString(r.Type))
					}
				}

				apiFn := Function{
					Name:        fn.Name.Name,
					Receiver:    receiver,
					Params:      params,
					ReturnTypes: returnTypes,
					TypeParams:  typeParams,
					Internal:    internal,
				}
				if !fn.Name.IsExported() && len(apiFn.ReturnTypes) == 1 && apiFn.ReturnTypes[0] == "component.Config" {
					result.ConfigStructName = extractFunctionReturnType(fn)
				} else if fn.Name.IsExported() {
					result.Functions = append(result.Functions, apiFn)
				}

			}
		}
	}
}

func extractFunctionReturnType(fn *ast.FuncDecl) string {
	ret, ok := fn.Body.List[len(fn.Body.List)-1].(*ast.ReturnStmt)
	if !ok || len(ret.Results) != 1 {
		return ""
	}
	switch r := ret.Results[0].(type) {
	case *ast.UnaryExpr:
		switch x := r.X.(type) {
		case *ast.Ident:
			return x.Name
		case *ast.CompositeLit:
			switch subt := x.Type.(type) {
			case *ast.Ident:
				return subt.Name
			case *ast.SelectorExpr:
				return subt.X.(*ast.Ident).Name
			}
		}
	case *ast.Ident:
		return r.Name
	case *ast.CallExpr:
		switch x := r.Fun.(type) {
		case *ast.SelectorExpr:
			if x.Sel.Obj != nil {
				return extractFunctionReturnType(x.Sel.Obj.Decl.(*ast.FuncDecl))
			}
		case *ast.Ident:
			if x.Obj != nil {
				return extractFunctionReturnType(x.Obj.Decl.(*ast.FuncDecl))
			}
		default:
			panic(fmt.Sprintf("[%s] Unsupported function reference %T", fn.Name.Name, x))
		}
	default:
		panic(fmt.Sprintf("[%s] Unsupported type %T", fn.Name.Name, r))
	}
	return ""
}

func isFunctionIgnored(ignoredFunctions []string, fnName string) bool {
	for _, v := range ignoredFunctions {
		reg := regexp.MustCompile(v)
		if reg.MatchString(fnName) {
			return true
		}
	}
	return false
}
