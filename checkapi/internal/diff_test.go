// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package internal

import (
	"encoding/json"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDiff(t *testing.T) {
	a, err := Read(filepath.Join("testpkg", "receiver", "validreceiver"), nil, nil)
	require.NoError(t, err)
	b := a
	bPreConst := b
	bPreConst.Values = append([]string{"foobar"}, bPreConst.Values...)
	bAddConst := b
	bAddConst.Values = append(bAddConst.Values, "foobar")
	bAddFn := b
	bAddFn.Functions = append(bAddFn.Functions, Function{
		Name:        "MyFn",
		Receiver:    "foo",
		ReturnTypes: []string{"string", "error"},
		Params:      []string{"string", "bool"},
		TypeParams:  []string{"~string"},
	})
	bBeforeFn := b
	bBeforeFn.Functions = append([]Function{{
		Name:        "Aaaa",
		Receiver:    "foo",
		ReturnTypes: []string{"string", "error"},
		Params:      []string{"string", "bool"},
	}}, bBeforeFn.Functions...)
	bAddStruct := b
	bAddStruct.Structs = append(bAddStruct.Structs, APIstruct{Name: "MyStruct", Fields: []string{"foo", "bar"}})

	tests := []struct {
		name         string
		a            API
		b            API
		expectedDiff Diff
	}{
		{
			name: "same",
			a:    a,
			b:    b,
			expectedDiff: Diff{
				Left:  API{},
				Equal: a,
				Right: API{},
			},
		},
		{
			name: "one more constant",
			a:    a,
			b:    bAddConst,
			expectedDiff: Diff{
				Left:  API{},
				Equal: a,
				Right: API{
					Values: []string{"foobar"},
				},
			},
		},
		{
			name: "prepended constant",
			a: API{
				Values: []string{"foobar"},
			},
			b: API{
				Values: []string{"aaa", "foobar"},
			},
			expectedDiff: Diff{
				Left: API{},
				Equal: API{
					Values: []string{"foobar"},
				},
				Right: API{
					Values: []string{"aaa"},
				},
			},
		},
		{
			name: "one more function",
			a:    a,
			b:    bAddFn,
			expectedDiff: Diff{
				Left:  API{},
				Equal: a,
				Right: API{
					Functions: []Function{
						{
							Name:        "MyFn",
							Receiver:    "foo",
							ReturnTypes: []string{"string", "error"},
							Params:      []string{"string", "bool"},
							TypeParams:  []string{"~string"},
						},
					},
				},
			},
		},
		{
			name: "one more struct",
			a:    a,
			b:    bAddStruct,
			expectedDiff: Diff{
				Left:  API{},
				Equal: a,
				Right: API{
					Structs: []APIstruct{
						{
							Name:   "MyStruct",
							Fields: []string{"foo", "bar"},
						},
					},
				},
			},
		},
		{
			name: "combined changes",
			a:    bAddFn,
			b:    bAddStruct,
			expectedDiff: Diff{
				Left: API{
					Functions: []Function{
						{
							Name:        "MyFn",
							Receiver:    "foo",
							ReturnTypes: []string{"string", "error"},
							Params:      []string{"string", "bool"},
							TypeParams:  []string{"~string"},
						},
					},
				},
				Equal: a,
				Right: API{
					Structs: []APIstruct{
						{
							Name:   "MyStruct",
							Fields: []string{"foo", "bar"},
						},
					},
				},
			},
		},
		{
			name: "fn added to right",
			a:    a,
			b:    bBeforeFn,
			expectedDiff: Diff{
				Left:  API{},
				Equal: a,
				Right: API{
					Functions: []Function{
						{
							Name:        "Aaaa",
							Receiver:    "foo",
							ReturnTypes: []string{"string", "error"},
							Params:      []string{"string", "bool"},
						},
					},
				},
			},
		},
		{
			name: "struct field added",
			a: API{
				Structs: []APIstruct{{Name: "MyStruct", Fields: []string{"foo", "bar"}}},
			},
			b: API{
				Structs: []APIstruct{{Name: "MyStruct", Fields: []string{"foo", "bar", "foobar"}}},
			},
			expectedDiff: Diff{
				Left: API{
					Structs: []APIstruct{{Name: "MyStruct", Fields: []string{"foo", "bar"}}},
				},
				Equal: API{},
				Right: API{
					Structs: []APIstruct{{Name: "MyStruct", Fields: []string{"foo", "bar", "foobar"}}},
				},
			},
		},
		{
			name: "add a parameter to a function",
			a: API{
				Functions: []Function{{Name: "MyFn", Params: []string{"foo", "bar"}}},
			},
			b: API{
				Functions: []Function{{Name: "MyFn", Params: []string{"foo", "bar", "foobar"}}},
			},
			expectedDiff: Diff{
				Left: API{
					Functions: []Function{{Name: "MyFn", Params: []string{"foo", "bar"}}},
				},
				Equal: API{},
				Right: API{
					Functions: []Function{{Name: "MyFn", Params: []string{"foo", "bar", "foobar"}}},
				},
			},
		},
		{
			name: "add a type parameter to a function",
			a: API{
				Functions: []Function{{Name: "MyFn", Params: []string{"foo", "bar"}, TypeParams: []string{"string"}}},
			},
			b: API{
				Functions: []Function{{Name: "MyFn", Params: []string{"foo", "bar"}, TypeParams: []string{"string", "bool"}}},
			},
			expectedDiff: Diff{
				Left: API{
					Functions: []Function{{Name: "MyFn", Params: []string{"foo", "bar"}, TypeParams: []string{"string"}}},
				},
				Equal: API{},
				Right: API{
					Functions: []Function{{Name: "MyFn", Params: []string{"foo", "bar"}, TypeParams: []string{"string", "bool"}}},
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(tt *testing.T) {
			d := Compare(test.a, test.b)
			assert.Equal(tt, test.expectedDiff, d)
			expectedJSON, _ := json.MarshalIndent(test.expectedDiff, "", "  ")
			newJSON, _ := json.MarshalIndent(d, "", "  ")
			assert.Equal(tt, string(expectedJSON), string(newJSON))
		})
	}
}
