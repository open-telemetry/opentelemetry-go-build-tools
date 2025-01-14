// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"flag"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"gopkg.in/yaml.v3"

	"go.opentelemetry.io/build-tools/githubgen/datatype"
)

// Generates files specific to GitHub according to status datatype.Metadata:
// .github/CODEOWNERS
// .github/ALLOWLIST
// .github/ISSUE_TEMPLATES/*.yaml (list of components)
// reports/distributions/*
func main() {
	folder := flag.String("folder", ".", "folder investigated for codeowners")
	allowlistFilePath := flag.String("allowlist", "cmd/githubgen/allowlist.txt", "path to a file containing an allowlist of members outside the defined Github organization")
	skipGithubCheck := flag.Bool("skipgithub", false, "skip checking if codeowners are part of the GitHub organization")
	defaultCodeOwner := flag.String("default-codeowner", "@open-telemetry/collector-contrib-approvers", "GitHub user or team name that will be used as default codeowner")
	trimSuffixes := flag.String("trim-component-suffixes", "receiver, exporter, extension, processor, connector, internal", "Define a comma-separated list of suffixes that should be trimmed from paths during generation of issue templates")
	githubOrgSlug := flag.String("github-org", "open-telemetry", "GitHub organization name to check if codeowners are org members")

	flag.Parse()
	var generators []datatype.Generator

	for _, arg := range flag.Args() {
		switch arg {
		case "issue-templates":
			generators = append(generators, newIssueTemplatesGenerator(trimSuffixes))
		case "codeowners":
			generators = append(generators, newCodeownersGenerator(skipGithubCheck))
		case "distributions":
			generators = append(generators, newDistributionsGenerator())
		default:
			panic(fmt.Sprintf("Unknown datatype.Generator: %s", arg))
		}
	}
	if len(generators) == 0 {
		generators = []datatype.Generator{newIssueTemplatesGenerator(trimSuffixes), newCodeownersGenerator(skipGithubCheck)}
	}

	distributions, err := getDistributions(*folder)
	if err != nil {
		log.Fatal(err)
	}

	if err = run(*folder, *allowlistFilePath, generators, distributions, *defaultCodeOwner, *githubOrgSlug); err != nil {
		log.Fatal(err)
	}
}

func loadMetadata(filePath string) (datatype.Metadata, error) {
	yamlFile, err := os.ReadFile(filePath)
	if err != nil {
		return datatype.Metadata{}, err
	}

	md := datatype.Metadata{}
	if err = yaml.Unmarshal(yamlFile, &md); err != nil {
		return md, err
	}

	return md, nil
}

func run(folder string, allowlistFilePath string, generators []datatype.Generator, distros []datatype.DistributionData, defaultCodeOwner string, githubOrg string) error {
	components := map[string]datatype.Metadata{}
	var foldersList []string
	maxLength := 0
	var allCodeowners []string
	err := filepath.Walk(folder, func(path string, info fs.FileInfo, _ error) error {
		if info.Name() == "metadata.yaml" {
			m, err := loadMetadata(path)
			if err != nil {
				return err
			}
			if m.Status == nil {
				return nil
			}
			currentFolder := filepath.Dir(path)

			components[currentFolder] = m
			foldersList = append(foldersList, currentFolder)

			for stability := range m.Status.Stability {
				if stability == unmaintainedStatus {
					// do not account for unmaintained status to change the max length of the component line.
					return nil
				}
			}
			if m.Status.Codeowners == nil && defaultCodeOwner != "" {
				log.Printf("component %q has no codeowners section, using default codeowner: %s\n", currentFolder, defaultCodeOwner)
				defaultOwners := &datatype.Codeowners{
					Active: []string{},
				}
				m.Status.Codeowners = defaultOwners
			} else if m.Status.Codeowners == nil && defaultCodeOwner == "" {
				return fmt.Errorf("component %q has no codeowners section", currentFolder)
			}

			allCodeowners = append(allCodeowners, m.Status.Codeowners.Active...)
			if len(currentFolder) > maxLength {
				maxLength = len(currentFolder)
			}
		}
		return nil
	})
	if err != nil {
		return err
	}

	slices.Sort(foldersList)
	slices.Sort(allCodeowners)
	allCodeowners = slices.Compact(allCodeowners)

	if !strings.HasPrefix(defaultCodeOwner, "@") {
		defaultCodeOwner = "@" + defaultCodeOwner
	}

	data := datatype.GithubData{
		RootFolder:        folder,
		Folders:           foldersList,
		Codeowners:        allCodeowners,
		AllowlistFilePath: allowlistFilePath,
		MaxLength:         maxLength,
		Components:        components,
		Distributions:     distros,
		DefaultCodeOwner:  defaultCodeOwner,
		GitHubOrg:         githubOrg,
	}

	for _, g := range generators {
		if err = g.Generate(data); err != nil {
			return err
		}
	}
	return nil
}

func getDistributions(folder string) ([]datatype.DistributionData, error) {
	var distributions []datatype.DistributionData
	dd, err := os.ReadFile(filepath.Join(folder, "distributions.yaml")) // nolint: gosec
	if err != nil {
		return nil, err
	}
	err = yaml.Unmarshal(dd, &distributions)
	if err != nil {
		return nil, err
	}
	return distributions, nil
}

func newIssueTemplatesGenerator(trimSuffixes *string) *issueTemplatesGenerator {
	suffixSlice := strings.Split(*trimSuffixes, ", ")
	return &issueTemplatesGenerator{
		trimSuffixes: suffixSlice,
	}
}

func newCodeownersGenerator(skipGithubCheck *bool) *codeownersGenerator {
	return &codeownersGenerator{skipGithub: *skipGithubCheck, getGitHubMembers: GetGithubMembers}
}

func newDistributionsGenerator() *distributionsGenerator {
	return &distributionsGenerator{}
}
