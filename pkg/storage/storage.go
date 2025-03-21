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
		return NewGitHubStorage(config)
	case "local":
		return NewLocalStorage(config)
	default:
		return nil, fmt.Errorf("unsupported storage mode: %s", config.StorageMode)
	}
}
