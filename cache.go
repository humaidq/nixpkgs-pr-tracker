package main

import (
	"context"
	"encoding/gob"
	"fmt"
	"os"
	"slices"
	"sync"
	"time"
)

// GHPullRequest represents the metadata of a GitHub pull request. It only
// includes relevant fields, and is used in the cache.
type GHPullRequest struct {
	ID             int
	Title          string
	AuthorUsername string
	MergeBranch    string
	CommitHash     string
	CachedAt       time.Time
}

// Cache represents the cache of the application. It includes the commit map,
// the last cached commits of the branches, and the pull request cache.
// It includes methods that are concurrency-safe.
type Cache struct {
	Version         int
	CommitMap       map[string][]string
	mapLock         sync.Mutex
	Built           bool
	LastBranchHeads map[string]string
	PRCache         map[int]GHPullRequest
	prCacheLock     sync.Mutex
}

func (c *Cache) SetLastBranchHead(branch, hash string) {
	c.mapLock.Lock()
	defer c.mapLock.Unlock()
	c.LastBranchHeads[branch] = hash
}

func (c *Cache) GetLastBranchHead(branch string) string {
	c.mapLock.Lock()
	defer c.mapLock.Unlock()
	return c.LastBranchHeads[branch]
}

func NewCache() *Cache {
	return &Cache{
		Version:         1,
		CommitMap:       make(map[string][]string),
		LastBranchHeads: make(map[string]string),
		PRCache:         make(map[int]GHPullRequest),
	}
}

func (c *Cache) GetBranchesForCommit(hash string) []string {
	c.mapLock.Lock()
	defer c.mapLock.Unlock()
	return c.CommitMap[hash]
}

func (c *Cache) CommitExistsInBranch(hash, branch string) bool {
	c.mapLock.Lock()
	defer c.mapLock.Unlock()
	return slices.Contains(c.CommitMap[hash], branch)
}

func (c *Cache) AddCommitToBranch(hash, branch string) {
	c.mapLock.Lock()
	defer c.mapLock.Unlock()

	if !slices.Contains(c.CommitMap[hash], branch) {
		c.CommitMap[hash] = append(c.CommitMap[hash], branch)
	}
}

func (c *Cache) getGHPR(prId int) (GHPullRequest, error) {
	c.prCacheLock.Lock()
	if pr, ok := c.PRCache[prId]; ok {
		c.prCacheLock.Unlock()
		return pr, nil
	}
	// Don't let the request block accessing the cache
	c.prCacheLock.Unlock()

	pr, _, err := client.PullRequests.Get(context.Background(), "NixOS", "nixpkgs", prId)
	if err != nil {
		return GHPullRequest{}, err
	}
	ghpr := GHPullRequest{
		ID:             prId,
		Title:          *pr.Title,
		AuthorUsername: *pr.User.Login,
		MergeBranch:    *pr.Base.Ref,
		CommitHash:     *pr.Head.SHA,
		CachedAt:       time.Now(),
	}

	c.prCacheLock.Lock()
	c.PRCache[prId] = ghpr
	c.prCacheLock.Unlock()
	return ghpr, nil
}

func (c *Cache) SaveToFile(filename string) error {
	c.mapLock.Lock()
	defer c.mapLock.Unlock()
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("error creating file: %v", err)
	}
	defer file.Close()

	encoder := gob.NewEncoder(file)
	err = encoder.Encode(c)
	if err != nil {
		return fmt.Errorf("error encoding data: %v", err)
	}
	return nil
}

// LoadFromFile loads the Cache struct from a file using gob encoding
func (c *Cache) LoadFromFile(filename string) error {
	c.mapLock.Lock()
	defer c.mapLock.Unlock()
	file, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("error opening file: %v", err)
	}
	defer file.Close()

	// Create a new gob decoder and read the Cache from the file
	decoder := gob.NewDecoder(file)
	err = decoder.Decode(c)
	if err != nil {
		return fmt.Errorf("error decoding data: %v", err)
	}
	return nil
}
