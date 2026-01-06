// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0
package main

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"sort"
	"strings"

	"github.com/google/go-github/v81/github"

	"go.opentelemetry.io/build-tools/githubgen/datatype"
)

type codeownersGenerator struct {
	skipGithub       bool
	getGitHubMembers func(skipGithub bool, githubOrg string) (map[string]struct{}, error)
	getFile          func(fileName string) ([]byte, error)
	setFile          func(fileName string, data []byte) error
}

func (cg *codeownersGenerator) Generate(data datatype.GithubData) error {
	allowlistData, err := cg.getFile(data.AllowlistFilePath)
	if err != nil {
		return err
	}

	err = cg.verifyCodeOwnerOrgMembership(allowlistData, data)
	if err != nil {
		return err
	}

	var ownerComponents, allowListUnmaintainedComponents, unmaintainedCodeowners, distributions, allowListDeprecatedComponents []string

LOOP:
	for _, folder := range data.Folders {
		m := data.Components[folder]
		// check if component is unmaintained or deprecated
		for stability := range m.Status.Stability {
			if stability == unmaintainedStatus {
				allowListUnmaintainedComponents = append(allowListUnmaintainedComponents, folder)
				unmaintainedCodeowners = append(unmaintainedCodeowners, fmt.Sprintf("%s/%s %s", folder, strings.Repeat(" ", data.MaxLength-len(folder)), data.DefaultCodeOwner))
				continue LOOP
			}
			if stability == deprecatedStatus && (m.Status.Codeowners == nil || len(m.Status.Codeowners.Active) == 0) {
				allowListDeprecatedComponents = append(allowListDeprecatedComponents, folder+"/\n")
			}
		}

		// check and handle active codeowners
		if m.Status.Codeowners != nil {
			owners := ""
			for _, owner := range m.Status.Codeowners.Active {
				owners += " "
				owners += formatGithubUser(owner)
			}
			ownerComponents = append(ownerComponents, fmt.Sprintf("%s/%s %s%s", strings.TrimPrefix(folder, data.RootFolder+"/"), strings.Repeat(" ", data.MaxLength-len(folder)), data.DefaultCodeOwner, owners))
		}
	}

	longestName := cg.longestNameSpaces(data)

	for _, dist := range data.Distributions {
		var maintainers []string
		for _, m := range dist.Maintainers {
			maintainers = append(maintainers, formatGithubUser(m))
		}

		distribution := fmt.Sprintf("reports/distributions/%s.yaml%s %s", dist.Name, strings.Repeat(" ", longestName-len(dist.Name)), data.DefaultCodeOwner)
		if len(maintainers) > 0 {
			distribution += fmt.Sprintf(" %s", strings.Join(maintainers, " "))
		}

		distributions = append(distributions, distribution)
	}

	// CODEOWNERS file
	codeownersFile := filepath.Join(data.RootFolder, ".github", "CODEOWNERS")
	templateContents, err := cg.getFile(codeownersFile)
	if err != nil {
		return err
	}

	templateContents = injectContent(startComponentList, endComponentList, templateContents, ownerComponents)
	templateContents = injectContent(startDistributionList, endDistributionList, templateContents, distributions)
	templateContents = injectContent(startUnmaintainedList, endUnmaintainedList, templateContents, unmaintainedCodeowners)

	err = cg.setFile(codeownersFile, templateContents)
	if err != nil {
		return err
	}

	// ALLOWLIST file
	allowListFile := filepath.Join(data.RootFolder, ".github", "ALLOWLIST")
	allowListContents, err := cg.getFile(allowListFile)
	if err != nil {
		return err
	}

	allowListContents = injectContent(startUnmaintainedList, endUnmaintainedList, allowListContents, allowListUnmaintainedComponents)
	allowListContents = injectContent(startDeprecatedList, endDeprecatedList, allowListContents, allowListDeprecatedComponents)

	err = cg.setFile(allowListFile, allowListContents)
	if err != nil {
		return err
	}
	return nil
}

func injectContent(startMagicString, endMagicString string, templateContents []byte, replaceContent []string) []byte {
	matchOldContent := regexp.MustCompile("(?s)" + startMagicString + ".*" + endMagicString)
	oldContent := matchOldContent.FindSubmatch(templateContents)
	if len(oldContent) > 0 {
		replacement := []byte(startMagicString + "\n\n" + strings.Join(replaceContent, "\n") + "\n\n" + endMagicString)
		templateContents = bytes.ReplaceAll(templateContents, oldContent[0], replacement)
	}
	return templateContents
}

func formatGithubUser(user string) string {
	if !strings.HasPrefix(user, "@") {
		return "@" + user
	}
	return user
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

func getGithubMembers(skipGithub bool, githubOrg string) (map[string]struct{}, error) {
	if skipGithub {
		// don't try to get organization members if no token is expected
		return map[string]struct{}{}, nil
	}
	githubToken := os.Getenv("GITHUB_TOKEN")
	if githubToken == "" {
		return nil, fmt.Errorf("set the environment variable `GITHUB_TOKEN` to a PAT token to authenticate")
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

func getFile(fileName string) ([]byte, error) {
	return os.ReadFile(fileName) // nolint: gosec
}

func setFile(fileName string, data []byte) error {
	return os.WriteFile(fileName, data, 0o600) // nolint: gosec
}
