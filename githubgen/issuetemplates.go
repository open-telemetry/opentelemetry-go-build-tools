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

type issueTemplatesGenerator struct {
	trimSuffixes []string
}

// folderToSlug removes redundant suffixes from a path.
//
// A path like receiver/myvendorreceiver will be trimmed to receiver/myvendor.
//
// All parts of the path except for the first level will be trimmed.
func (itg *issueTemplatesGenerator) folderToSlug(folder string) string {
	path := strings.Split(folder, "/")
	exists := false
	for _, suffix := range itg.trimSuffixes {
		if strings.Contains(path[0], suffix) {
			exists = true
			break
		}
	}

	if exists {
		for _, suffix := range itg.trimSuffixes {
			path[1] = strings.TrimSuffix(path[1], suffix)
			path[len(path)-1] = strings.TrimSuffix(path[len(path)-1], suffix)
		}
	}

	return strings.Join(path, "/")
}

// Generate takes all files in the .github/ISSUE_TEMPLATE folder, looks for a magic start
// and end string and fills in the list of found components in-between.
func (itg *issueTemplatesGenerator) Generate(data datatype.GithubData) error {
	var componentSlugs []string

	for _, folder := range data.Folders {
		componentSlugs = append(componentSlugs, itg.folderToSlug(strings.TrimPrefix(folder, data.RootFolder+"/")))
	}
	sort.Strings(componentSlugs)

	replacement := []byte(startComponentList + "\n      - " + strings.Join(componentSlugs, "\n      - ") + "\n      " + endComponentList)
	issuesFolder := filepath.Join(data.RootFolder, ".github", "ISSUE_TEMPLATE")
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
