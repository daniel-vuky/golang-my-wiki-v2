package storage

import (
	"fmt"

	"github.com/daniel-vuky/golang-my-wiki-v2/pkg/config"
	"github.com/daniel-vuky/golang-my-wiki-v2/pkg/storage/types"
)

// Storage defines the interface for different storage backends
type Storage interface {
	// Page operations
	ListPages() ([]types.Page, error)
	GetPage(path string) (*types.Page, error)
	CreatePage(page *types.Page) error
	UpdatePage(page *types.Page) error
	DeletePage(path string) error

	// Folder operations
	ListFolders() ([]string, error)
	CreateFolder(path string) error
	DeleteFolder(path string) error
}

// NewStorage creates a new storage instance based on the configuration
func NewStorage(config *config.Config) (types.Storage, error) {
	switch config.StorageMode {
	case "github":
		// If Redis cache is enabled, use cached GitHub storage
		if config.Redis.Enabled {
			return NewCachedGitHubStorage(config)
		}
		// Otherwise use regular GitHub storage
		return NewGitHubStorage(config)
	case "local":
		// If Redis cache is enabled, use cached local storage
		if config.Redis.Enabled {
			return NewCachedLocalStorage(config)
		}
		// Otherwise use regular local storage
		return NewLocalStorage(config)
	default:
		return nil, fmt.Errorf("unsupported storage mode: %s", config.StorageMode)
	}
}
