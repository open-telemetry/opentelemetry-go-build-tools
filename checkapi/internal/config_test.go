// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package internal

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_FunctionToString(t *testing.T) {
	fn := Function{
		Name:        "fn",
		Receiver:    "foo",
		ReturnTypes: []string{"string", "bool"},
		Params:      []string{"string", "int64"},
		TypeParams:  []string{"T"},
	}
	require.Equal(t, "foo.fn[T](string,int64) string,bool", fn.String())
}

func Test_ApiStructToString(t *testing.T) {
	fn := APIstruct{
		Name:   "MyStruct",
		Fields: []string{"Foo string", "Bar bool"},
	}
	require.Equal(t, "MyStruct(Foo string,Bar bool)", fn.String())
}

func Test_InterfaceToString(t *testing.T) {
	fn := Interface{
		Name: "Interf",
		Methods: []Function{
			{
				Name:        "fn",
				Receiver:    "foo",
				ReturnTypes: []string{"string", "bool"},
				Params:      []string{"string", "int64"},
				TypeParams:  []string{"T"},
			},
		},
	}
	require.Equal(t, "Interf{foo.fn[T](string,int64) string,bool}", fn.String())
}
