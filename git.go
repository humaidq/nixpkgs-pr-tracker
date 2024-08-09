// SPDX-FileCopyrightText: 2024 Humaid Alqasimi <https://huma.id>
// SPDX-License-Identifier: AGPL-3.0-or-later WITH GPL-3.0-linking-exception
package main

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/charmbracelet/log"
	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
)

const REPO_PATH = "nixpkgs"
const REPO_URL = "https://github.com/nixos/nixpkgs.git"
const OLDEST_YEAR = 00
const CACHE_FILE = "cache.bin"

type Indexer struct {
	Cache    *Cache
	workerWg sync.WaitGroup
}

var releaseRegex = regexp.MustCompile(`-(\d\d).(\d\d)`)

func NewIndexer() *Indexer {
	cache := NewCache()
	return &Indexer{
		Cache: cache,
	}
}

func (i *Indexer) LoadCache() error {
	e := i.Cache.LoadFromFile(CACHE_FILE)
	if e != nil {
		log.Error("Failed to load cache from file", "error", e)
	}

	if _, err := os.Stat(REPO_PATH); os.IsNotExist(err) {
		log.Infof("'%s' not found, cloning a fresh copy. This may take a while...\n", REPO_PATH)
		err := cloneNixpkgs()
		if err != nil {
			return fmt.Errorf("failed to clone nixpkgs: %w", err)
		}
		err = i.updateCommitMap(false)
		if err != nil {
			return fmt.Errorf("failed to update commit map: %w", err)
		}
	} else {
		err := i.updateCommitMap(true)
		if err != nil {
			return fmt.Errorf("failed to update nixpkgs & commit map: %w", err)
		}
	}
	i.Cache.Built = true
	log.Info("Cache built, saving to file...")
	err := i.Cache.SaveToFile(CACHE_FILE)
	if err != nil {
		log.Error("Failed to save cache to file", "error", err)
	}
	log.Info("Cache saved to file")

	// Cache update loop
	for {
		time.Sleep(15 * time.Minute)
		log.Info("scheduler: Updating commit map")
		err = i.updateCommitMap(true)
		if err != nil {
			log.Error("scheduler: Failed to update commit map", "error", err)
		}
		err = i.Cache.SaveToFile("cache.bin")
		if err != nil {
			log.Error("Failed to save cache to file", "error", err)
		}
		log.Info("Cache saved to file")
	}
}

func cloneNixpkgs() error {
	cmd := exec.Command("git", "clone", REPO_URL, REPO_PATH)
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("failed to clone nixpkgs: %w", err)
	}
	return nil
}

func updateNixpkgs() error {
	// Make sure it uses https instead of ssh
	cmd := exec.Command("git", "remote", "set-url", "origin", REPO_URL)
	cmd.Dir = REPO_PATH
	err := cmd.Run()

	// Fetch all branches
	// TODO don't give up
	cmd = exec.Command("git", "fetch", "--all")
	cmd.Dir = REPO_PATH
	err = cmd.Run()

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

func (i *Indexer) mapWorker(id int, jobs <-chan string) {
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
		last := i.Cache.GetLastBranchHead(branchName)
		for {
			c, err := cIter.Next()
			if err != nil {
				break
			}
			// We shouldn't re-index commits we've already indexed
			if c.Hash.String() == last {
				break
			}
			i.Cache.AddCommitToBranch(c.Hash.String(), branchName)
		}

		i.Cache.SetLastBranchHead(branchName, ref.Hash().String())

		log.Info("Completed mapping", "branch", branchName, "worker", id)
		i.workerWg.Done()
	}
}

func (i *Indexer) updateCommitMap(updateRepo bool) error {
	log.Info("Updating nixpkgs")
	if updateRepo {
		err := updateNixpkgs()
		if err != nil {
			return fmt.Errorf("updateCommitMap: %w", err)
		}
	}
	log.Info("Getting list of branches")
	branches, err := getAllBranchNames()
	if err != nil {
		return fmt.Errorf("updateCommitMap: %w", err)
	}
	log.Info("Starting to build commit map...")
	jobs := make(chan string, 3)
	for w := 1; w <= 3; w++ {
		go i.mapWorker(w, jobs)
	}

	for _, branchName := range branches {
		i.workerWg.Add(1)
		jobs <- branchName
	}

	i.workerWg.Wait()
	close(jobs)

	return nil
}
