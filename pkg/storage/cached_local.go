package storage

import (
	"log"
	"strings"

	"github.com/daniel-vuky/golang-my-wiki-v2/pkg/cache"
	"github.com/daniel-vuky/golang-my-wiki-v2/pkg/config"
	"github.com/daniel-vuky/golang-my-wiki-v2/pkg/storage/types"
)

// CachedLocalStorage wraps LocalStorage with a Redis cache
type CachedLocalStorage struct {
	local *LocalStorage
	cache *cache.RedisCache
}

// NewCachedLocalStorage creates a new local storage instance with Redis caching
func NewCachedLocalStorage(config *config.Config) (*CachedLocalStorage, error) {
	// Initialize local storage
	local, err := NewLocalStorage(config)
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

	return &CachedLocalStorage{
		local: local,
		cache: redisCache,
	}, nil
}

// ListPages retrieves all pages, using cache when available
func (cl *CachedLocalStorage) ListPages() ([]types.Page, error) {
	// Try to get from cache first
	pages, found, err := cl.cache.GetPageList()
	if err != nil {
		log.Printf("Cache error when listing pages: %v", err)
	}

	if found {
		log.Printf("Using cached page list with %d pages", len(pages))
		return pages, nil
	}

	// Cache miss, get from local storage
	log.Printf("Cache miss for page list, fetching from local storage")
	pages, err = cl.local.ListPages()
	if err != nil {
		return nil, err
	}

	// Store in cache for next time
	if err := cl.cache.SetPageList(pages); err != nil {
		log.Printf("Failed to cache page list: %v", err)
	}

	return pages, nil
}

// GetPage retrieves a specific page, using cache when available
func (cl *CachedLocalStorage) GetPage(path string) (*types.Page, error) {
	// Extract title from path
	title := strings.TrimSuffix(path, ".txt")

	// Try to get from cache first
	page, found, err := cl.cache.GetPage(title)
	if err != nil {
		log.Printf("Cache error when getting page %s: %v", title, err)
	}

	if found {
		log.Printf("Using cached page: %s", title)
		return page, nil
	}

	// Cache miss, get from local storage
	log.Printf("Cache miss for page %s, fetching from local storage", title)
	page, err = cl.local.GetPage(path)
	if err != nil {
		return nil, err
	}

	// Store in cache for next time
	if err := cl.cache.SetPage(page); err != nil {
		log.Printf("Failed to cache page %s: %v", title, err)
	}

	return page, nil
}

// CreatePage creates a new page and invalidates relevant caches
func (cl *CachedLocalStorage) CreatePage(page *types.Page) error {
	// Create page in local storage
	if err := cl.local.CreatePage(page); err != nil {
		return err
	}

	// Update cache
	if err := cl.cache.SetPage(page); err != nil {
		log.Printf("Failed to cache page %s: %v", page.Title, err)
	}

	// Invalidate the page list cache since we added a new page
	if err := cl.cache.InvalidateCache(); err != nil {
		log.Printf("Failed to invalidate cache: %v", err)
	}

	return nil
}

// UpdatePage updates an existing page and updates the cache
func (cl *CachedLocalStorage) UpdatePage(page *types.Page) error {
	// Update page in local storage
	if err := cl.local.UpdatePage(page); err != nil {
		return err
	}

	// Update cache
	if err := cl.cache.SetPage(page); err != nil {
		log.Printf("Failed to update page %s in cache: %v", page.Title, err)
	}

	return nil
}

// DeletePage deletes a page and invalidates relevant caches
func (cl *CachedLocalStorage) DeletePage(path string) error {
	// Delete page in local storage
	if err := cl.local.DeletePage(path); err != nil {
		return err
	}

	// Remove from cache
	title := strings.TrimSuffix(path, ".txt")
	if err := cl.cache.DeletePage(title); err != nil {
		log.Printf("Failed to delete page %s from cache: %v", title, err)
	}

	// Invalidate the page list cache since we removed a page
	if err := cl.cache.InvalidateCache(); err != nil {
		log.Printf("Failed to invalidate cache: %v", err)
	}

	return nil
}

// ListFolders retrieves all folders, using cache when available
func (cl *CachedLocalStorage) ListFolders() ([]string, error) {
	// Try to get from cache first
	folders, found, err := cl.cache.GetFolderList()
	if err != nil {
		log.Printf("Cache error when listing folders: %v", err)
	}

	if found {
		log.Printf("Using cached folder list with %d folders", len(folders))
		return folders, nil
	}

	// Cache miss, get from local storage
	log.Printf("Cache miss for folder list, fetching from local storage")
	folders, err = cl.local.ListFolders()
	if err != nil {
		return nil, err
	}

	// Store in cache for next time
	if err := cl.cache.SetFolderList(folders); err != nil {
		log.Printf("Failed to cache folder list: %v", err)
	}

	return folders, nil
}

// CreateFolder creates a new folder and invalidates folder cache
func (cl *CachedLocalStorage) CreateFolder(path string) error {
	// Create folder in local storage
	if err := cl.local.CreateFolder(path); err != nil {
		return err
	}

	// Invalidate the folder list cache
	if err := cl.cache.InvalidateCache(); err != nil {
		log.Printf("Failed to invalidate cache: %v", err)
	}

	return nil
}

// DeleteFolder deletes a folder and invalidates folder cache
func (cl *CachedLocalStorage) DeleteFolder(path string) error {
	// Delete folder in local storage
	if err := cl.local.DeleteFolder(path); err != nil {
		return err
	}

	// Invalidate the folder list cache
	if err := cl.cache.InvalidateCache(); err != nil {
		log.Printf("Failed to invalidate cache: %v", err)
	}

	return nil
}

// GetPagesInFolder retrieves all pages from a specific folder, using cache when available
func (cl *CachedLocalStorage) GetPagesInFolder(folderPath string) ([]types.Page, error) {
	// Try to get from cache first
	pages, found, err := cl.cache.GetFolderPages(folderPath)
	if err != nil {
		log.Printf("Cache error when getting pages in folder %s: %v", folderPath, err)
	}

	if found {
		log.Printf("Using cached pages for folder %s: %d pages", folderPath, len(pages))
		return pages, nil
	}

	// Cache miss, get from local storage
	log.Printf("Cache miss for folder %s pages, fetching from local storage", folderPath)
	pages, err = cl.local.GetPagesInFolder(folderPath)
	if err != nil {
		return nil, err
	}

	// Store in cache for next time
	if err := cl.cache.SetFolderPages(folderPath, pages); err != nil {
		log.Printf("Failed to cache pages for folder %s: %v", folderPath, err)
	}

	return pages, nil
}

// InvalidateCache invalidates all relevant caches
func (cl *CachedLocalStorage) InvalidateCache() error {
	return cl.cache.InvalidateCache()
}
