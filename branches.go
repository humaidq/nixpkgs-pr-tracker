// SPDX-FileCopyrightText: 2024 Humaid Alqasimi <https://huma.id>
// SPDX-License-Identifier: AGPL-3.0-or-later WITH GPL-3.0-linking-exception
package main

import (
	"context"
	"github.com/google/go-github/v63/github"
	"regexp"
	"slices"
	"strconv"
	"strings"
)

// GetBranchesForCommit returns the branches that contain the given commit
func GetBranchesForCommit(commitHash string) []string {
	lock.Lock()
	value := commitMap[commitHash]
	lock.Unlock()
	return value
}

var nextBranchTable = []struct {
	pattern string
	repl    string
}{
	{`^python-updates$`, "staging"},
	{`^staging$`, "staging-next"},
	{`^staging-next$`, "master"},
	{`^staging-next-([\d.]+)$`, "release-$1"},
	{`^haskell-updates$`, "master"},
	{`^master$`, "nixpkgs-unstable"},
	{`^master$`, "nixos-unstable-small"},
	{`^nixos-(.*)-small$`, "nixos-$1"},
	{`^release-([\d.]+)$`, "nixpkgs-$1-darwin"},
	{`^release-([\d.]+)$`, "nixos-$1-small"},
	{`^staging-((1.|20)\.\d{2})$`, "release-$1"},
	{`^staging-((2[1-9]|[3-90].)\.\d{2})$`, "staging-next-$1"},
}

var branchPrefixes = []string{
	"python-updates",
	"staging",
	"haskell-updates",
	"master",
	"release-",
	"haskell-updates",
	"nixpkgs-",
	"nixos-",
}

type PR struct {
	ID             int
	Title          string
	AuthorUsername string
	MergedStatus   string
	Branches       BranchTree
}

type BranchTree struct {
	BranchName string
	Accepted   bool
	HydraLink  string
	Children   []BranchTree
}

func GetBranchesForPR(prId int) (*PR, error) {
	client := github.NewClient(nil).WithAuthToken(githubToken)
	pr, _, err := client.PullRequests.Get(context.Background(), "NixOS", "nixpkgs", prId)
	if err != nil {
		return &PR{}, err
	}
	// PR target branch
	branchName := *pr.Base.Ref
	tree := buildBranches(branchName, *pr.Head.SHA)

	prResult := PR{
		ID:             prId,
		Title:          *pr.Title,
		AuthorUsername: *pr.User.Login,
		Branches:       tree,
	}

	return &prResult, nil
}

func buildBranches(branchName string, commit string) BranchTree {
	var children []BranchTree
	for _, row := range nextBranchTable {
		re := regexp.MustCompile(row.pattern)
		if re.MatchString(branchName) {
			children = append(children, buildBranches(re.ReplaceAllString(branchName, row.repl), commit))
		}
	}

	// Check whether commit exists in the branch
	lock.Lock()
	contains := slices.Contains(commitMap[commit], branchName)
	lock.Unlock()
	return BranchTree{
		BranchName: branchName,
		Accepted:   contains,
		Children:   children,
	}
}

func validBranchToCache(branchName string) bool {
	for _, pre := range branchPrefixes {
		if strings.HasPrefix(branchName, pre) {
			releaseMatch := releaseRegex.FindStringSubmatch(branchName)
			if len(releaseMatch) > 1 {
				year, err := strconv.Atoi(releaseMatch[1])
				if err == nil && year < oldestYear {
					return false
				}
			}
			return true
		}
	}
	return false
}
