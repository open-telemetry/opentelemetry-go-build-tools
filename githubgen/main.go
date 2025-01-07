// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"slices"

	"go.opentelemetry.io/collector/confmap"
	"go.opentelemetry.io/collector/confmap/provider/fileprovider"
	"gopkg.in/yaml.v3"

	"go.opentelemetry.io/build-tools/githubgen/datatype"
)

const unmaintainedStatus = "unmaintained"

// Generates files specific to GitHub according to status datatype.Metadata:
// .github/CODEOWNERS
// .github/ALLOWLIST
// .github/ISSUE_TEMPLATES/*.yaml (list of components)
// reports/distributions/*
func main() {
	folder := flag.String("folder", ".", "folder investigated for codeowners")
	allowlistFilePath := flag.String("allowlist", "cmd/githubgen/allowlist.txt", "path to a file containing an allowlist of members outside the OpenTelemetry organization")
	skipGithubCheck := flag.Bool("skipgithub", false, "skip checking GitHub membership check for CODEOWNERS datatype.Generator")
	repoName := flag.String("repo-name", "", "name of the repository (e.g. \"OpenTelemetry Collector Contrib\")")
	defaultCodeOwner := flag.String("default-codeowners", "", "GitHub user or team name that will be used as default codeowner")

	flag.Parse()
	var generators []datatype.Generator
	for _, arg := range flag.Args() {
		switch arg {
		case "issue-templates":
			generators = append(generators, &issueTemplatesGenerator{})
		case "codeowners":
			generators = append(generators, &codeownersGenerator{skipGithub: *skipGithubCheck})
		case "distributions":
			generators = append(generators, &distributionsGenerator{})
		default:
			panic(fmt.Sprintf("Unknown datatype.Generator: %s", arg))
		}
	}
	if len(generators) == 0 {
		generators = []datatype.Generator{&issueTemplatesGenerator{}, &codeownersGenerator{skipGithub: *skipGithubCheck}}
	}

	distributions, err := getDistributions(*folder)
	if err != nil {
		log.Fatal(err)
	}

	if err = run(*folder, *allowlistFilePath, *repoName, generators, distributions, *defaultCodeOwner); err != nil {
		log.Fatal(err)
	}
}

func loadMetadata(filePath string) (datatype.Metadata, error) {
	cp, err := fileprovider.NewFactory().Create(confmap.ProviderSettings{}).Retrieve(context.Background(), "file:"+filePath, nil)
	if err != nil {
		return datatype.Metadata{}, err
	}

	conf, err := cp.AsConf()
	if err != nil {
		return datatype.Metadata{}, err
	}

	md := datatype.Metadata{}
	if err := conf.Unmarshal(&md, confmap.WithIgnoreUnused()); err != nil {
		return md, err
	}

	return md, nil
}

func run(folder string, allowlistFilePath string, repoName string, generators []datatype.Generator, distros []datatype.DistributionData, defaultCodeOwner string) error {
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
			if m.Status.Codeowners == nil {
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

	data := datatype.GithubData{
		Folders:           foldersList,
		Codeowners:        allCodeowners,
		AllowlistFilePath: allowlistFilePath,
		MaxLength:         maxLength,
		Components:        components,
		Distributions:     distros,
		RepoName:          repoName,
		DefaultCodeOwner:  defaultCodeOwner,
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
