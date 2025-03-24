package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/daniel-vuky/golang-my-wiki-v2/pkg/storage/types"
	"github.com/go-redis/redis/v8"
)

const (
	// Key prefixes for different types of cached data
	pageCachePrefix  = "page:"
	pageListCacheKey = "pages:all"
	foldersCacheKey  = "folders:all"
)

// RedisCache implements a Redis-based caching layer
type RedisCache struct {
	client           *redis.Client
	ctx              context.Context
	enabled          bool
	expirationPeriod time.Duration
}

// NewRedisCache creates a new Redis cache instance
func NewRedisCache(addr string, expirationSeconds int) (*RedisCache, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	ctx := context.Background()

	// Create expiration duration from seconds
	expiration := time.Duration(expirationSeconds) * time.Second

	// Test connection
	_, err := client.Ping(ctx).Result()
	if err != nil {
		log.Printf("Redis connection failed: %v", err)
		// Return a disabled cache rather than failing
		return &RedisCache{
			client:           client,
			ctx:              ctx,
			enabled:          false,
			expirationPeriod: expiration,
		}, nil
	}

	log.Println("Redis cache initialized successfully")
	log.Printf("Cache expiration set to %v", expiration)
	return &RedisCache{
		client:           client,
		ctx:              ctx,
		enabled:          true,
		expirationPeriod: expiration,
	}, nil
}

// Close closes the Redis connection
func (c *RedisCache) Close() error {
	if c.enabled {
		return c.client.Close()
	}
	return nil
}

// SetPage caches a page
func (c *RedisCache) SetPage(page *types.Page) error {
	if !c.enabled {
		return nil
	}

	data, err := json.Marshal(page)
	if err != nil {
		return fmt.Errorf("failed to marshal page: %v", err)
	}

	key := pageCachePrefix + page.Title
	if err := c.client.Set(c.ctx, key, data, c.expirationPeriod).Err(); err != nil {
		return fmt.Errorf("failed to cache page: %v", err)
	}

	log.Printf("Cached page: %s", page.Title)
	return nil
}

// GetPage retrieves a page from cache
func (c *RedisCache) GetPage(title string) (*types.Page, bool, error) {
	if !c.enabled {
		return nil, false, nil
	}

	key := pageCachePrefix + title
	data, err := c.client.Get(c.ctx, key).Bytes()
	if err == redis.Nil {
		// Cache miss
		return nil, false, nil
	}
	if err != nil {
		return nil, false, fmt.Errorf("failed to get page from cache: %v", err)
	}

	var page types.Page
	if err := json.Unmarshal(data, &page); err != nil {
		return nil, false, fmt.Errorf("failed to unmarshal page: %v", err)
	}

	log.Printf("Cache hit for page: %s", title)
	return &page, true, nil
}

// DeletePage removes a page from cache
func (c *RedisCache) DeletePage(title string) error {
	if !c.enabled {
		return nil
	}

	key := pageCachePrefix + title
	if err := c.client.Del(c.ctx, key).Err(); err != nil {
		return fmt.Errorf("failed to delete page from cache: %v", err)
	}

	// Also invalidate the page list cache since it's now stale
	c.client.Del(c.ctx, pageListCacheKey)

	log.Printf("Deleted page from cache: %s", title)
	return nil
}

// SetPageList caches the list of all pages
func (c *RedisCache) SetPageList(pages []types.Page) error {
	if !c.enabled {
		return nil
	}

	data, err := json.Marshal(pages)
	if err != nil {
		return fmt.Errorf("failed to marshal page list: %v", err)
	}

	if err := c.client.Set(c.ctx, pageListCacheKey, data, c.expirationPeriod).Err(); err != nil {
		return fmt.Errorf("failed to cache page list: %v", err)
	}

	log.Printf("Cached page list with %d pages", len(pages))
	return nil
}

// GetPageList retrieves the list of all pages from cache
func (c *RedisCache) GetPageList() ([]types.Page, bool, error) {
	if !c.enabled {
		return nil, false, nil
	}

	data, err := c.client.Get(c.ctx, pageListCacheKey).Bytes()
	if err == redis.Nil {
		// Cache miss
		return nil, false, nil
	}
	if err != nil {
		return nil, false, fmt.Errorf("failed to get page list from cache: %v", err)
	}

	var pages []types.Page
	if err := json.Unmarshal(data, &pages); err != nil {
		return nil, false, fmt.Errorf("failed to unmarshal page list: %v", err)
	}

	log.Printf("Cache hit for page list")
	return pages, true, nil
}

// SetFolderList caches the list of all folders
func (c *RedisCache) SetFolderList(folders []string) error {
	if !c.enabled {
		return nil
	}

	data, err := json.Marshal(folders)
	if err != nil {
		return fmt.Errorf("failed to marshal folder list: %v", err)
	}

	if err := c.client.Set(c.ctx, foldersCacheKey, data, c.expirationPeriod).Err(); err != nil {
		return fmt.Errorf("failed to cache folder list: %v", err)
	}

	log.Printf("Cached folder list with %d folders", len(folders))
	return nil
}

// GetFolderList retrieves the list of all folders from cache
func (c *RedisCache) GetFolderList() ([]string, bool, error) {
	if !c.enabled {
		return nil, false, nil
	}

	data, err := c.client.Get(c.ctx, foldersCacheKey).Bytes()
	if err == redis.Nil {
		// Cache miss
		return nil, false, nil
	}
	if err != nil {
		return nil, false, fmt.Errorf("failed to get folder list from cache: %v", err)
	}

	var folders []string
	if err := json.Unmarshal(data, &folders); err != nil {
		return nil, false, fmt.Errorf("failed to unmarshal folder list: %v", err)
	}

	log.Printf("Cache hit for folder list")
	return folders, true, nil
}

// ClearAll clears all cached data
func (c *RedisCache) ClearAll() error {
	if !c.enabled {
		return nil
	}

	if err := c.client.FlushDB(c.ctx).Err(); err != nil {
		return fmt.Errorf("failed to flush cache: %v", err)
	}

	log.Println("Cleared all cached data")
	return nil
}

// InvalidateCache invalidates all cache entries related to pages and folders
func (c *RedisCache) InvalidateCache() error {
	if !c.enabled {
		return nil
	}

	// Clear the page list cache
	c.client.Del(c.ctx, pageListCacheKey)

	// Clear the folders cache
	c.client.Del(c.ctx, foldersCacheKey)

	// Clear folder pages cache - we need to find and delete all keys with pattern "folder_pages:*"
	pattern := "folder_pages:*"
	var cursor uint64
	for {
		keys, nextCursor, err := c.client.Scan(c.ctx, cursor, pattern, 100).Result()
		if err != nil {
			return fmt.Errorf("failed to scan for folder pages keys: %v", err)
		}

		if len(keys) > 0 {
			if err := c.client.Del(c.ctx, keys...).Err(); err != nil {
				log.Printf("Error deleting folder pages keys: %v", err)
			} else {
				log.Printf("Deleted %d folder pages cache entries", len(keys))
			}
		}

		if nextCursor == 0 {
			break
		}
		cursor = nextCursor
	}

	log.Println("Invalidated cache for page lists and folder lists")
	return nil
}

// SetFolderPages caches pages for a specific folder
func (c *RedisCache) SetFolderPages(folderPath string, pages []types.Page) error {
	if !c.enabled {
		return nil
	}

	data, err := json.Marshal(pages)
	if err != nil {
		return fmt.Errorf("failed to marshal folder pages: %v", err)
	}

	key := "folder_pages:" + folderPath
	if err := c.client.Set(c.ctx, key, data, c.expirationPeriod).Err(); err != nil {
		return fmt.Errorf("failed to cache folder pages: %v", err)
	}

	log.Printf("Cached %d pages for folder: %s", len(pages), folderPath)
	return nil
}

// GetFolderPages retrieves pages for a specific folder from cache
func (c *RedisCache) GetFolderPages(folderPath string) ([]types.Page, bool, error) {
	if !c.enabled {
		return nil, false, nil
	}

	key := "folder_pages:" + folderPath
	data, err := c.client.Get(c.ctx, key).Bytes()
	if err == redis.Nil {
		// Cache miss
		return nil, false, nil
	}
	if err != nil {
		return nil, false, fmt.Errorf("failed to get folder pages from cache: %v", err)
	}

	var pages []types.Page
	if err := json.Unmarshal(data, &pages); err != nil {
		return nil, false, fmt.Errorf("failed to unmarshal folder pages: %v", err)
	}

	log.Printf("Cache hit for folder pages: %s", folderPath)
	return pages, true, nil
}
