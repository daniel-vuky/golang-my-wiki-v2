package storage

import (
	"log"
	"strings"

	"github.com/daniel-vuky/golang-my-wiki-v2/pkg/cache"
	"github.com/daniel-vuky/golang-my-wiki-v2/pkg/config"
	"github.com/daniel-vuky/golang-my-wiki-v2/pkg/storage/types"
	"github.com/google/go-github/v45/github"
)

// CachedGitHubStorage wraps GitHubStorage with a Redis cache
type CachedGitHubStorage struct {
	github *GitHubStorage
	cache  *cache.RedisCache
}

// NewCachedGitHubStorage creates a new GitHub storage instance with Redis caching
func NewCachedGitHubStorage(config *config.Config) (*CachedGitHubStorage, error) {
	// Initialize GitHub storage
	github, err := NewGitHubStorage(config)
	if err != nil {
		return nil, err
	}

	// Initialize Redis cache
	redisAddr := config.Redis.Address
	if redisAddr == "" {
		redisAddr = "localhost:6379" // Default Redis address
	}

	// Use expiration from config
	expirationSeconds := config.Redis.ExpirationSeconds

	redisCache, err := cache.NewRedisCache(redisAddr, expirationSeconds)
	if err != nil {
		log.Printf("Warning: Redis cache initialization failed: %v. Continuing without cache.", err)
		// Continue without cache
	}

	return &CachedGitHubStorage{
		github: github,
		cache:  redisCache,
	}, nil
}

// ListPages retrieves all pages, using cache when available
func (cg *CachedGitHubStorage) ListPages() ([]types.Page, error) {
	// Try to get from cache first
	pages, found, err := cg.cache.GetPageList()
	if err != nil {
		log.Printf("Cache error when listing pages: %v", err)
	}

	if found {
		log.Printf("Using cached page list with %d pages", len(pages))
		return pages, nil
	}

	// Cache miss, get from GitHub
	log.Printf("Cache miss for page list, fetching from GitHub")
	pages, err = cg.github.ListPages()
	if err != nil {
		return nil, err
	}

	// Store in cache for next time
	if err := cg.cache.SetPageList(pages); err != nil {
		log.Printf("Failed to cache page list: %v", err)
	}

	return pages, nil
}

// GetPage retrieves a specific page, using cache when available
func (cg *CachedGitHubStorage) GetPage(path string) (*types.Page, error) {
	// Extract title from path
	title := strings.TrimSuffix(path, ".txt")

	// Try to get from cache first
	page, found, err := cg.cache.GetPage(title)
	if err != nil {
		log.Printf("Cache error when getting page %s: %v", title, err)
	}

	if found {
		log.Printf("Using cached page: %s", title)
		return page, nil
	}

	// Cache miss, get from GitHub
	log.Printf("Cache miss for page %s, fetching from GitHub", title)
	page, err = cg.github.GetPage(path)
	if err != nil {
		return nil, err
	}

	// Store in cache for next time
	if err := cg.cache.SetPage(page); err != nil {
		log.Printf("Failed to cache page %s: %v", title, err)
	}

	return page, nil
}

// CreatePage creates a new page and invalidates relevant caches
func (cg *CachedGitHubStorage) CreatePage(page *types.Page) error {
	// Check if page exists in GitHub first by trying to get its contents
	// This is similar to how the UpdatePage method in github.go checks for existence

	// Make sure the path has .txt extension
	if !strings.HasSuffix(page.Path, ".txt") {
		page.Path = page.Path + ".txt"
	}

	log.Printf("Checking if file %s already exists in GitHub before attempting to create", page.Path)

	// Check if the page exists in GitHub
	// We can't use cg.GetPage here as it could return a cached version that's inconsistent
	// Instead, directly check GitHub
	_, _, resp, err := cg.github.client.Repositories.GetContents(
		cg.github.ctx,
		cg.github.owner,
		cg.github.repository,
		page.Path,
		&github.RepositoryContentGetOptions{Ref: cg.github.branch},
	)

	if err != nil {
		if resp != nil && resp.StatusCode == 404 {
			// 404 means file doesn't exist, which is what we want for a create operation
			log.Printf("File %s does not exist in GitHub (404), proceeding with create", page.Path)
		} else {
			// Some other error occurred when checking file existence
			log.Printf("Error checking if file %s exists: %v", page.Path, err)
			if resp != nil {
				log.Printf("Response status: %s", resp.Status)
			}
		}
	} else {
		// File exists, use update instead to get the SHA
		log.Printf("File %s already exists in GitHub, using UpdatePage instead", page.Path)
		return cg.UpdatePage(page)
	}

	// Cache check complete, proceed with actual GitHub API call

	// Create page in GitHub
	if err := cg.github.CreatePage(page); err != nil {
		return err
	}

	// Update cache
	if err := cg.cache.SetPage(page); err != nil {
		log.Printf("Failed to cache page %s: %v", page.Title, err)
	}

	// Invalidate the page list cache since we added a new page
	if err := cg.cache.InvalidateCache(); err != nil {
		log.Printf("Failed to invalidate cache: %v", err)
	}

	return nil
}

// UpdatePage updates an existing page and updates the cache
func (cg *CachedGitHubStorage) UpdatePage(page *types.Page) error {
	// Update page in GitHub
	if err := cg.github.UpdatePage(page); err != nil {
		return err
	}

	// Update cache
	if err := cg.cache.SetPage(page); err != nil {
		log.Printf("Failed to update page %s in cache: %v", page.Title, err)
	}

	return nil
}

// DeletePage deletes a page and invalidates relevant caches
func (cg *CachedGitHubStorage) DeletePage(path string) error {
	// Delete page in GitHub
	if err := cg.github.DeletePage(path); err != nil {
		return err
	}

	// Remove from cache
	title := strings.TrimSuffix(path, ".txt")
	if err := cg.cache.DeletePage(title); err != nil {
		log.Printf("Failed to delete page %s from cache: %v", title, err)
	}

	// Invalidate the page list cache since we removed a page
	if err := cg.cache.InvalidateCache(); err != nil {
		log.Printf("Failed to invalidate cache: %v", err)
	}

	return nil
}

// ListFolders retrieves all folders, using cache when available
func (cg *CachedGitHubStorage) ListFolders() ([]string, error) {
	// Try to get from cache first
	folders, found, err := cg.cache.GetFolderList()
	if err != nil {
		log.Printf("Cache error when listing folders: %v", err)
	}

	if found {
		log.Printf("Using cached folder list with %d folders", len(folders))
		return folders, nil
	}

	// Cache miss, get from GitHub
	log.Printf("Cache miss for folder list, fetching from GitHub")
	folders, err = cg.github.ListFolders()
	if err != nil {
		return nil, err
	}

	// Store in cache for next time
	if err := cg.cache.SetFolderList(folders); err != nil {
		log.Printf("Failed to cache folder list: %v", err)
	}

	return folders, nil
}

// CreateFolder creates a new folder and invalidates folder cache
func (cg *CachedGitHubStorage) CreateFolder(path string) error {
	// Create folder in GitHub
	if err := cg.github.CreateFolder(path); err != nil {
		return err
	}

	// Invalidate the folder list cache
	if err := cg.cache.InvalidateCache(); err != nil {
		log.Printf("Failed to invalidate cache: %v", err)
	}

	return nil
}

// DeleteFolder deletes a folder and invalidates folder cache
func (cg *CachedGitHubStorage) DeleteFolder(path string) error {
	// Delete folder in GitHub
	if err := cg.github.DeleteFolder(path); err != nil {
		return err
	}

	// Invalidate the folder list cache
	if err := cg.cache.InvalidateCache(); err != nil {
		log.Printf("Failed to invalidate cache: %v", err)
	}

	return nil
}

// GetPagesInFolder retrieves all pages from a specific folder, using cache when available
func (cg *CachedGitHubStorage) GetPagesInFolder(folderPath string) ([]types.Page, error) {
	// Try to get from cache first
	pages, found, err := cg.cache.GetFolderPages(folderPath)
	if err != nil {
		log.Printf("Cache error when getting pages in folder %s: %v", folderPath, err)
	}

	if found {
		log.Printf("Using cached pages for folder %s: %d pages", folderPath, len(pages))
		return pages, nil
	}

	// Cache miss, get from GitHub
	log.Printf("Cache miss for folder %s pages, fetching from GitHub", folderPath)
	pages, err = cg.github.GetPagesInFolder(folderPath)
	if err != nil {
		return nil, err
	}

	// Store in cache for next time
	if err := cg.cache.SetFolderPages(folderPath, pages); err != nil {
		log.Printf("Failed to cache pages for folder %s: %v", folderPath, err)
	}

	return pages, nil
}
