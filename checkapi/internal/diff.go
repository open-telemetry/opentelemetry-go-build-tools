// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package internal

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"slices"
	"strings"
)

// Diff is the set of changes only in the first API, the common API, and the set of changes in the second API.
type Diff struct {
	Left  API `json:"left"`
	Equal API `json:"equal"`
	Right API `json:"right"`
}

func separate[E any](left []E, right []E, comp func(E, E) int) ([]E, []E, []E) {
	i := 0
	j := 0
	var l []E
	var e []E
	var r []E
	for i < len(left) || j < len(right) {
		if i >= len(left) {
			r = append(r, right[j])
			j++
			continue
		}
		if j >= len(right) {
			l = append(l, left[i])
			i++
			continue
		}
		nextl := left[i]
		nextr := right[j]
		result := comp(nextl, nextr)
		switch result {
		case 0:
			e = append(e, nextl)
			i++
			j++
		case -1:
			l = append(l, nextl)
			i++
		case 1:
			r = append(r, nextr)
			j++
		}
	}
	return l, e, r
}

// ReadAPI reads an API from a file.
func ReadAPI(path string) (API, error) {
	var api API
	b, err := os.ReadFile(path) // #nosec G304
	if os.IsNotExist(err) {
		return api, nil
	}
	err = json.Unmarshal(b, &api)
	return api, err
}

// WriteAPI writes an API to a file.
func WriteAPI(a API, path string) error {
	b, err := json.MarshalIndent(a, "", "  ")
	if err != nil {
		return err
	}
	err = os.WriteFile(path, b, 0600)
	return err
}

// Compare compares two APIs and computes a Diff object.
func Compare(left API, right API) Diff {
	l := API{}
	e := API{}
	r := API{}

	l.Values, e.Values, r.Values = separate(left.Values, right.Values, strings.Compare)
	fnComp := func(function Function, function2 Function) int {
		if name := strings.Compare(function.Name, function2.Name); name != 0 {
			return name
		}
		if receiver := strings.Compare(function.Receiver, function2.Receiver); receiver != 0 {
			return receiver
		}
		if returnTypes := slices.Compare(function.ReturnTypes, function2.ReturnTypes); returnTypes != 0 {
			return returnTypes
		}
		if paramTypes := slices.Compare(function.TypeParams, function2.TypeParams); paramTypes != 0 {
			return paramTypes
		}
		if params := slices.Compare(function.Params, function2.Params); params != 0 {
			return params
		}
		return 0
	}

	l.Functions, e.Functions, r.Functions = separate(left.Functions, right.Functions, fnComp)
	l.Structs, e.Structs, r.Structs = separate(left.Structs, right.Structs, func(a APIstruct, b APIstruct) int {
		if name := strings.Compare(a.Name, b.Name); name != 0 {
			return name
		}
		if fields := slices.Compare(a.Fields, b.Fields); fields != 0 {
			return fields
		}
		return 0
	})
	l.Interfaces, e.Interfaces, r.Interfaces = separate(left.Interfaces, right.Interfaces, func(a Interface, b Interface) int {
		if name := strings.Compare(a.Name, b.Name); name != 0 {
			return name
		}
		if methods := slices.CompareFunc(a.Methods, b.Methods, fnComp); methods != 0 {
			return methods
		}
		return 0
	})

	return Diff{
		Left:  l,
		Equal: e,
		Right: r,
	}
}

func (d Diff) identical() bool {
	return len(d.Left.Functions) == 0 && len(d.Left.Interfaces) == 0 && len(d.Left.Values) == 0 && len(d.Left.Structs) == 0 &&
		len(d.Right.Functions) == 0 && len(d.Right.Interfaces) == 0 && len(d.Right.Values) == 0 && len(d.Right.Structs) == 0
}

func (d Diff) Error(errorOnAddition, errorOnRemoval bool) error {
	if d.identical() {
		return nil
	}
	msg := ""

	if errorOnRemoval {
		for _, f := range d.Left.Functions {
			msg += fmt.Sprintf("- Missing function %s\n", f)
		}
		for _, f := range d.Left.Structs {
			msg += fmt.Sprintf("- Missing struct %s\n", f)
		}
		for _, f := range d.Left.Interfaces {
			msg += fmt.Sprintf("- Missing interface %s\n", f)
		}
		for _, f := range d.Left.Values {
			msg += fmt.Sprintf("- Missing values %s\n", f)
		}
	}

	if errorOnAddition {
		for _, f := range d.Right.Functions {
			msg += fmt.Sprintf("- New function %s\n", f)
		}
		for _, f := range d.Right.Structs {
			msg += fmt.Sprintf("- New struct %s\n", f)
		}
		for _, f := range d.Right.Interfaces {
			msg += fmt.Sprintf("- New interface %s\n", f)
		}
		for _, f := range d.Right.Values {
			msg += fmt.Sprintf("- New values %s\n", f)
		}
	}

	return errors.New(msg)
}
