// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0
package main

import (
	"bytes"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"go.opentelemetry.io/build-tools/githubgen/datatype"
)

func folderToShortName(folder string) string {
	if folder == "internal/coreinternal" {
		return "internal/core"
	}

	suffixes := []string{"receiver", "exporter", "extension", "processor", "connector"}

	path := strings.Split(folder, "/")
	if strings.Contains(path[0], suffixes[0]) ||
		strings.Contains(path[0], suffixes[1]) ||
		strings.Contains(path[0], suffixes[2]) ||
		strings.Contains(path[0], suffixes[3]) ||
		strings.Contains(path[0], suffixes[4]) {
		for _, suffix := range suffixes {
			path[1] = strings.TrimSuffix(path[1], suffix)
			path[len(path)-1] = strings.TrimSuffix(path[len(path)-1], suffix)
		}
	}

	return strings.Join(path, "/")
}

type issueTemplatesGenerator struct{}

func (itg *issueTemplatesGenerator) Generate(data datatype.GithubData) error {
	var componentSlugs []string

	for _, folder := range data.Folders {
		componentSlugs = append(componentSlugs, folderToShortName(folder))
	}
	sort.Strings(componentSlugs)

	replacement := []byte(startComponentList + "\n      - " + strings.Join(componentSlugs, "\n      - ") + "\n      " + endComponentList)
	issuesFolder := filepath.Join(".github", "ISSUE_TEMPLATE")
	entries, err := os.ReadDir(issuesFolder)
	if err != nil {
		return err
	}
	for _, e := range entries {
		templateContents, err := os.ReadFile(filepath.Join(issuesFolder, e.Name())) // nolint: gosec
		if err != nil {
			return err
		}
		matchOldContent := regexp.MustCompile("(?s)" + startComponentList + ".*" + endComponentList)
		oldContent := matchOldContent.FindSubmatch(templateContents)
		if len(oldContent) > 0 {
			templateContents = bytes.ReplaceAll(templateContents, oldContent[0], replacement)
			err = os.WriteFile(filepath.Join(issuesFolder, e.Name()), templateContents, 0o600)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
