// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0
package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strings"

	"github.com/google/go-github/v66/github"

	"go.opentelemetry.io/build-tools/githubgen/datatype"
)

type codeownersGenerator struct {
	skipGithub       bool
	getGitHubMembers func(skipGithub bool, githubOrg string) (map[string]struct{}, error)
}

func (cg *codeownersGenerator) Generate(data datatype.GithubData) error {
	allowlistData, err := os.ReadFile(data.AllowlistFilePath)
	if err != nil {
		return err
	}

	err = cg.verifyCodeOwnerOrgMembership(allowlistData, data)
	if err != nil {
		return err
	}

	codeowners := fmt.Sprintf(codeownersHeader, data.DefaultCodeOwner)
	deprecatedList := deprecatedListHeader
	unmaintainedList := unmaintainedListHeader

	unmaintainedCodeowners := unmaintainedHeader
	currentFirstSegment := ""

LOOP:
	for _, folder := range data.Folders {
		m := data.Components[folder]
		for stability := range m.Status.Stability {
			if stability == unmaintainedStatus {
				unmaintainedList += folder + "/\n"
				unmaintainedCodeowners += fmt.Sprintf("%s/%s %s \n", folder, strings.Repeat(" ", data.MaxLength-len(folder)), data.DefaultCodeOwner)
				continue LOOP
			}
			if stability == "deprecated" && (m.Status.Codeowners == nil || len(m.Status.Codeowners.Active) == 0) {
				deprecatedList += folder + "/\n"
			}
		}

		if m.Status.Codeowners != nil {
			parts := strings.Split(folder, string(os.PathSeparator))
			firstSegment := parts[0]
			if firstSegment != currentFirstSegment {
				currentFirstSegment = firstSegment
				codeowners += "\n"
			}
			owners := ""
			for _, owner := range m.Status.Codeowners.Active {
				owners += " "
				if !strings.HasPrefix(owner, "@") {
					owners += "@" + owner
				}
			}
			codeowners += fmt.Sprintf("%s/%s %s%s\n", strings.TrimPrefix(folder, data.RootFolder+"/"), strings.Repeat(" ", data.MaxLength-len(folder)), data.DefaultCodeOwner, owners)
		}
	}

	codeowners += distributionCodeownersHeader
	longestName := cg.longestNameSpaces(data)

	for _, dist := range data.Distributions {
		var maintainers []string
		for _, m := range dist.Maintainers {
			maintainers = append(maintainers, fmt.Sprintf("@%s", m))
		}
		codeowners += fmt.Sprintf("reports/distributions/%s.yaml%s %s %s\n", dist.Name, strings.Repeat(" ", longestName-len(dist.Name)), data.DefaultCodeOwner, strings.Join(maintainers, " "))
	}

	err = os.WriteFile(filepath.Join(data.RootFolder, ".github", "CODEOWNERS"), []byte(codeowners+unmaintainedCodeowners), 0o600)
	if err != nil {
		return err
	}
	err = os.WriteFile(filepath.Join(data.RootFolder, ".github", "ALLOWLIST"), []byte(allowlistHeader+deprecatedList+unmaintainedList), 0o600)
	if err != nil {
		return err
	}
	return nil
}

func (cg *codeownersGenerator) longestNameSpaces(data datatype.GithubData) int {
	longestName := 0
	for _, dist := range data.Distributions {
		if longestName < len(dist.Name) {
			longestName = len(dist.Name)
		}
	}
	return longestName
}

// verifyCodeOwnerOrgMembership verifies that all codeOwners are members of the defined GitHub organization
//
// If a codeOwner is not part of the GitHub org, that user will be looked for in the allowlist.
//
// The method returns an error if:
// - there are code owners that are not org members and not in the allowlist (only if skipGithub is set to false)
// - there are redundant entries in the allowlist
// - there are entries in the allowlist that are unused
func (cg *codeownersGenerator) verifyCodeOwnerOrgMembership(allowlistData []byte, data datatype.GithubData) error {
	allowlist := strings.Split(string(allowlistData), "\n")
	allowlist = slices.DeleteFunc(allowlist, func(s string) bool {
		return s == ""
	})
	unusedAllowlist := append([]string{}, allowlist...)

	var missingCodeowners []string
	var duplicateCodeowners []string

	members, err := cg.getGitHubMembers(cg.skipGithub, data.GitHubOrg)
	if err != nil {
		return err
	}

	// sort codeowners
	for _, codeowner := range data.Codeowners {
		_, ownerPresentInMembers := members[codeowner]

		if !ownerPresentInMembers {
			ownerInAllowlist := slices.Contains(allowlist, codeowner)
			unusedAllowlist = slices.DeleteFunc(unusedAllowlist, func(s string) bool {
				return s == codeowner
			})

			ownerInAllowlist = ownerInAllowlist || strings.HasPrefix(codeowner, data.GitHubOrg+"/")

			if !ownerInAllowlist {
				missingCodeowners = append(missingCodeowners, codeowner)
			}
		} else if slices.Contains(allowlist, codeowner) {
			duplicateCodeowners = append(duplicateCodeowners, codeowner)
		}
	}

	// error cases
	if len(missingCodeowners) > 0 && !cg.skipGithub {
		sort.Strings(missingCodeowners)
		return fmt.Errorf("codeowners are not members: %s", strings.Join(missingCodeowners, ", "))
	}
	if len(duplicateCodeowners) > 0 {
		sort.Strings(duplicateCodeowners)
		return fmt.Errorf("codeowners members duplicate in allowlist: %s", strings.Join(duplicateCodeowners, ", "))
	}
	if len(unusedAllowlist) > 0 {
		unused := append([]string{}, unusedAllowlist...)
		sort.Strings(unused)
		return fmt.Errorf("unused members in allowlist: %s", strings.Join(unused, ", "))
	}
	return err
}

func GetGithubMembers(skipGithub bool, githubOrg string) (map[string]struct{}, error) {
	if skipGithub {
		// don't try to get organization members if no token is expected
		return map[string]struct{}{}, nil
	}
	githubToken := os.Getenv("GITHUB_TOKEN")
	if githubToken == "" {
		return nil, fmt.Errorf("Set the environment variable `GITHUB_TOKEN` to a PAT token to authenticate")
	}
	client := github.NewClient(nil).WithAuthToken(githubToken)
	var allUsers []*github.User
	pageIndex := 0
	for {
		users, resp, err := client.Organizations.ListMembers(context.Background(), githubOrg,
			&github.ListMembersOptions{
				PublicOnly: false,
				ListOptions: github.ListOptions{
					PerPage: 50,
					Page:    pageIndex,
				},
			},
		)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()
		if len(users) == 0 {
			break
		}
		allUsers = append(allUsers, users...)
		pageIndex++
	}

	usernames := make(map[string]struct{}, len(allUsers))
	for _, u := range allUsers {
		usernames[*u.Login] = struct{}{}
	}
	return usernames, nil
}
