// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0
package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/google/go-github/v66/github"

	"go.opentelemetry.io/build-tools/githubgen/datatype"
)

type codeownersGenerator struct {
	skipGithub bool
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

	codeowners := fmt.Sprintf(codeownersHeader, data.RepoName, data.DefaultCodeOwner)
	deprecatedList := deprecatedListHeader
	unmaintainedList := unmaintainedListHeader

	unmaintainedCodeowners := unmaintainedHeader
	currentFirstSegment := ""

LOOP:
	for _, key := range data.Folders {
		m := data.Components[key]
		for stability := range m.Status.Stability {
			if stability == unmaintainedStatus {
				unmaintainedList += key + "/\n"
				unmaintainedCodeowners += fmt.Sprintf("%s/%s %s \n", key, strings.Repeat(" ", data.MaxLength-len(key)), data.DefaultCodeOwner)
				continue LOOP
			}
			if stability == "deprecated" && (m.Status.Codeowners == nil || len(m.Status.Codeowners.Active) == 0) {
				deprecatedList += key + "/\n"
			}
		}

		if m.Status.Codeowners != nil {
			parts := strings.Split(key, string(os.PathSeparator))
			firstSegment := parts[0]
			if firstSegment != currentFirstSegment {
				currentFirstSegment = firstSegment
				codeowners += "\n"
			}
			owners := ""
			for _, owner := range m.Status.Codeowners.Active {
				owners += " "
				owners += "@" + owner
			}
			codeowners += fmt.Sprintf("%s/%s %s%s\n", key, strings.Repeat(" ", data.MaxLength-len(key)), data.DefaultCodeOwner, owners)
		}
	}

	codeowners += fmt.Sprintf(distributionCodeownersHeader, data.RepoName)
	longestName := 0
	for _, dist := range data.Distributions {
		if longestName < len(dist.Name) {
			longestName = len(dist.Name)
		}
	}

	for _, dist := range data.Distributions {
		var maintainers []string
		for _, m := range dist.Maintainers {
			maintainers = append(maintainers, fmt.Sprintf("@%s", m))
		}
		codeowners += fmt.Sprintf("reports/distributions/%s.yaml%s %s %s\n", dist.Name, strings.Repeat(" ", longestName-len(dist.Name)), data.DefaultCodeOwner, strings.Join(maintainers, " "))
	}

	err = os.WriteFile(filepath.Join(".github", "CODEOWNERS"), []byte(codeowners+unmaintainedCodeowners), 0o600)
	if err != nil {
		return err
	}
	err = os.WriteFile(filepath.Join(".github", "ALLOWLIST"), []byte(fmt.Sprintf(allowlistHeader, data.RepoName)+deprecatedList+unmaintainedList), 0o600)
	if err != nil {
		return err
	}
	return nil
}

func (cg *codeownersGenerator) verifyCodeOwnerOrgMembership(allowlistData []byte, data datatype.GithubData) error {
	allowlistLines := strings.Split(string(allowlistData), "\n")

	allowlist := make(map[string]struct{}, len(allowlistLines))
	unusedAllowlist := make(map[string]struct{}, len(allowlistLines))

	for _, line := range allowlistLines {
		if line == "" {
			continue
		}
		allowlist[line] = struct{}{}
		unusedAllowlist[line] = struct{}{}
	}

	var missingCodeowners []string
	var duplicateCodeowners []string

	members, err := cg.GetGithubMembers()
	if err != nil {
		return err
	}

	// sort codeowners
	for _, codeowner := range data.Codeowners {
		_, ownerPresentInMembers := members[codeowner]

		if !ownerPresentInMembers {
			_, ownerInAllowlist := allowlist[codeowner]
			delete(unusedAllowlist, codeowner)
			ownerInAllowlist = ownerInAllowlist || strings.HasPrefix(codeowner, "open-telemetry/")
			if !ownerInAllowlist {
				missingCodeowners = append(missingCodeowners, codeowner)
			}
		} else if _, exists := allowlist[codeowner]; exists {
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
		var unused []string
		for k := range unusedAllowlist {
			unused = append(unused, k)
		}
		sort.Strings(unused)
		return fmt.Errorf("unused members in allowlist: %s", strings.Join(unused, ", "))
	}
	return err
}

func (cg *codeownersGenerator) GetGithubMembers() (map[string]struct{}, error) {
	if cg.skipGithub {
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
		users, resp, err := client.Organizations.ListMembers(context.Background(), "open-telemetry",
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
