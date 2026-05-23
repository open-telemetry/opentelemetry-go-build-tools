// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0
// Package findhelper finds dependents of a module.
package findhelper

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"go.opentelemetry.io/build-tools/grater/internal/module"
	"go.opentelemetry.io/build-tools/grater/internal/workspace"
)

func fetchDependents(mod module.Module) ([]module.Module, error) {
	url := "https://pkg.go.dev/" + mod.ModulePath + "?tab=importedby"
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, err
	}

	dependents := []module.Module{}
	doc.Find(".ImportedBy-details a").Each(func(i int, s *goquery.Selection) {
		pkg := strings.TrimSpace(s.Text())
		if pkg != "" {
			dependents = append(dependents, *module.NewModule(pkg, mod.ModuleVersion))
		}
	})

	return dependents, nil
}

func FindDependents(ws *workspace.Workspace, data string) error {
	modulePath, moduleVersion, found := strings.Cut(data, "@")
	if !found {
		return fmt.Errorf("FindDependents needs a remote path in the form module@version")
	}

	mod := *module.NewModule(modulePath, moduleVersion)

	dependents, err := fetchDependents(mod)
	if err != nil {
		return err
	}

	ws.AddDependents(dependents)
	return ws.WriteDependents()
}