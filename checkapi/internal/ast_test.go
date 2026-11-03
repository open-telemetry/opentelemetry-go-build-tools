// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package internal

import (
	"go/ast"
	"go/parser"
	"go/token"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAST(t *testing.T) {
	tests := []struct {
		name     string
		expr     ast.Expr
		expected string
	}{
		{
			"nil",
			nil,
			"",
		}, {
			"star",
			&ast.StarExpr{},
			"*",
		}, {
			"[]string",
			&ast.ArrayType{
				Elt: ast.NewIdent("string"),
			},
			"[]string",
		}, {
			"map",
			&ast.MapType{
				Key:   ast.NewIdent("string"),
				Value: ast.NewIdent("string"),
			},
			"map[string]string",
		}, {
			"map with struct value",
			&ast.MapType{
				Key: ast.NewIdent("string"),
				Value: &ast.StructType{
					Fields: &ast.FieldList{List: []*ast.Field{
						{
							Type: ast.NewIdent("string"),
						},
					}},
				},
			},
			"map[string]{string}",
		},
		{
			"func",
			&ast.FuncType{Params: &ast.FieldList{List: []*ast.Field{}}},
			"func()",
		},
		{
			"func with params and return type",
			&ast.FuncType{
				Params: &ast.FieldList{
					List: []*ast.Field{
						{
							Names: []*ast.Ident{
								ast.NewIdent("foo"),
							},
							Type: ast.NewIdent("bool"),
						},
					},
				}, Results: &ast.FieldList{List: []*ast.Field{
					{
						Names: []*ast.Ident{
							ast.NewIdent("foo"),
						},
						Type: ast.NewIdent("int"),
					},
				}},
			},
			"func(bool) int",
		},
		{
			"1",
			&ast.BasicLit{Value: "1"},
			"1",
		},
		{
			"...",
			&ast.Ellipsis{},
			"...",
		},
		{
			"index",
			&ast.IndexExpr{},
			"[]",
		},
		{
			"index complete",
			&ast.IndexExpr{
				X:     ast.NewIdent("foo"),
				Index: &ast.BasicLit{Value: "1"},
			},
			"foo[1]",
		},

		{
			"selector",
			&ast.SelectorExpr{
				X:   ast.NewIdent("foo"),
				Sel: ast.NewIdent("bar"),
			},
			"foo.bar",
		},
		{
			"interface",
			&ast.InterfaceType{Methods: &ast.FieldList{List: []*ast.Field{
				{
					Type: &ast.FuncType{Params: &ast.FieldList{List: []*ast.Field{}}},
				},
			}}},
			"{func func()}",
		},
		{
			"chan",
			&ast.ChanType{
				Value: ast.NewIdent("string"),
			},
			"chan(string)",
		},
		{
			"generics",
			&ast.FuncType{
				TypeParams: &ast.FieldList{List: []*ast.Field{
					{
						Names: []*ast.Ident{
							ast.NewIdent("T"),
						},
						Type: ast.NewIdent("T ~string"),
					},
				}},
				Params: &ast.FieldList{List: []*ast.Field{
					{
						Names: []*ast.Ident{
							ast.NewIdent("foo"),
						},
						Type: ast.NewIdent("T"),
					},
				},
				},
				Results: nil,
			},
			"func[T ~string](T)",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.expected, ExprToString(test.expr))
		})
	}
}

func TestCollectQualifiedLiterals(t *testing.T) {
	src := `package p

type Config struct{}

func createDefaultConfig() any {
	local := Config{}
	_ = local
	return &Config{
		A: configgrpc.ClientConfig{},
		B: configtls.ClientConfig{},
		C: []configgrpc.ServerConfig{{}},
	}
}
`
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "code.go", src, 0)
	require.NoError(t, err)
	fn := f.Decls[len(f.Decls)-1].(*ast.FuncDecl)

	literals := collectQualifiedLiterals(fset, fn)
	got := make([]string, 0, len(literals))
	for _, l := range literals {
		got = append(got, l.Type)
	}
	assert.Equal(t, []string{"configgrpc.ClientConfig", "configtls.ClientConfig"}, got)

	assert.Equal(t, "code.go", literals[0].File)
	assert.Equal(t, 9, literals[0].Line)
}
