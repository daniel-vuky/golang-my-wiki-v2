package storage

import (
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
	GetPagesInFolder(folderPath string) ([]types.Page, error)

	// Folder operations
	ListFolders() ([]string, error)
	CreateFolder(path string) error
	DeleteFolder(path string) error

	// Sync operations
	Sync() error
}

// NewStorage creates a new storage instance using combined local and GitHub storage
func NewStorage(config *config.Config) (types.Storage, error) {
	// Always use combined storage that supports both local and GitHub
	return NewCombinedStorage(config)
}
