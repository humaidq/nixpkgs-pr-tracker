// SPDX-FileCopyrightText: 2024 Humaid Alqasimi <https://huma.id>
// SPDX-License-Identifier: AGPL-3.0-or-later WITH GPL-3.0-linking-exception
package main

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/charmbracelet/log"
	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
)

// Commit map contains commits and which branches it exists in
var commitMap = make(map[string][]string)
var cacheBuilt = false
var lock sync.Mutex

var releaseRegex = regexp.MustCompile(`-(\d\d).(\d\d)`)

const oldestYear = 23

const REPO_PATH = "nixpkgs"

func setupCache() error {
	// Fetch fresh clone of nixpkgs if it doesn't exist
	if _, err := os.Stat(REPO_PATH); os.IsNotExist(err) {
		log.Infof("'%s' not found, cloning a fresh copy. This may take a while...\n", REPO_PATH)
		err := cloneNixpkgs()
		if err != nil {
			return fmt.Errorf("failed to clone nixpkgs: %w", err)
		}
		err = updateCommitMap(false)
		if err != nil {
			return fmt.Errorf("failed to update commit map: %w", err)
		}
	} else {
		err := updateCommitMap(true)
		if err != nil {
			return fmt.Errorf("failed to update nixpkgs & commit map: %w", err)
		}
	}
	cacheBuilt = true

	// Cache update loop
	for {
		time.Sleep(15 * time.Minute)
		log.Info("scheduler: Updating commit map")
		err := updateCommitMap(true)
		if err != nil {
			log.Error("scheduler: Failed to update commit map", "error", err)
		}
	}
}

func cloneNixpkgs() error {
	cmd := exec.Command("git", "clone", "https://github.com/nixos/nixpkgs.git", REPO_PATH)
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("failed to clone nixpkgs: %w", err)
	}
	return nil
}

func updateNixpkgs() error {
	cmd := exec.Command("git", "fetch", "--all")
	cmd.Dir = REPO_PATH
	err := cmd.Run()

	if err != nil {
		return fmt.Errorf("failed to update nixpkgs: %w", err)
	}
	return nil
}

func getAllBranchNames() (branches []string, err error) {
	repo, err := git.PlainOpen(REPO_PATH)
	if err != nil {
		return []string{}, fmt.Errorf("getAllBranchNames: failed to open repo: %w", err)
	}
	remote, err := repo.Remote("origin")
	if err != nil {
		return []string{}, fmt.Errorf("getAllBranchNames: failed to load origin remote: %w", err)
	}
	refs, err := remote.List(&git.ListOptions{})
	if err != nil {
		return []string{}, fmt.Errorf("getAllBranchNames: failed to list remote refs: %w", err)
	}

	refPrefix := "refs/heads"

	for _, ref := range refs {
		refName := ref.Name().String()

		if !strings.HasPrefix(refName, refPrefix) {
			continue
		}

		branchName := refName[len(refPrefix)+1:]
		if validBranchToCache(branchName) {
			branches = append(branches, branchName)
		}
	}

	return branches, nil
}

func mapWorker(id int, jobs <-chan string) {
	for branchName := range jobs {
		repo, err := git.PlainOpen(REPO_PATH)
		if err != nil {
			return
		}
		log.Info("Building map", "branch", branchName, "worker", id)
		ref, err := repo.Reference(plumbing.NewRemoteReferenceName("origin", branchName), true)
		if err != nil {
			log.Error("Couldn't get reference for branch", "branch", branchName, "error", err, "worker", id)
			return
		}

		// Check if the commit hash exists in the branch history
		cIter, err := repo.Log(&git.LogOptions{From: ref.Hash()})
		if err != nil {
			log.Error("Couldn't get log for branch", "branch", branchName, "error", err, "worker", id)
			return
		}
		for {
			c, err := cIter.Next()
			if err != nil {
				break
			}
			lock.Lock()
			if !slices.Contains(commitMap[c.Hash.String()], branchName) {
				commitMap[c.Hash.String()] = append(commitMap[c.Hash.String()], branchName)
			}
			lock.Unlock()
		}
		log.Info("Completed mapping", "branch", branchName, "worker", id)
	}
}

func updateCommitMap(updateRepo bool) error {
	if updateRepo {
		err := updateNixpkgs()
		if err != nil {
			return fmt.Errorf("updateCommitMap: %w", err)
		}
	}
	branches, err := getAllBranchNames()
	if err != nil {
		return fmt.Errorf("updateCommitMap: %w", err)
	}
	log.Info("Starting to build commit map...")
	jobs := make(chan string, 3)
	for w := 1; w <= 3; w++ {
		go mapWorker(w, jobs)
	}

	for _, branchName := range branches {
		jobs <- branchName
	}

	return nil
}
