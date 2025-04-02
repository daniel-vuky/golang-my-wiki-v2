package storage

import (
	"fmt"
	"log"
	"strings"

	"github.com/daniel-vuky/golang-my-wiki-v2/pkg/config"
	"github.com/daniel-vuky/golang-my-wiki-v2/pkg/storage/types"
)

// CombinedStorage implements both local and GitHub storage
type CombinedStorage struct {
	local  types.Storage
	github types.Storage
}

// NewCombinedStorage creates a new combined storage instance
func NewCombinedStorage(cfg *config.Config) (*CombinedStorage, error) {
	local, err := NewLocalStorage(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create local storage: %v", err)
	}

	github, err := NewGitHubStorage(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create GitHub storage: %v", err)
	}

	return &CombinedStorage{
		local:  local,
		github: github,
	}, nil
}

// Sync synchronizes data between local and GitHub storage
func (s *CombinedStorage) Sync() error {
	// First pull from GitHub to get latest changes
	if err := s.pullFromGitHub(); err != nil {
		return fmt.Errorf("failed to pull from GitHub: %v", err)
	}

	// Then push local changes to GitHub
	if err := s.pushToGitHub(); err != nil {
		return fmt.Errorf("failed to push to GitHub: %v", err)
	}

	return nil
}

// pullFromGitHub pulls changes from GitHub to local storage
func (s *CombinedStorage) pullFromGitHub() error {
	// Get all pages from GitHub
	pages, err := s.github.ListPages()
	if err != nil {
		return err
	}

	// For each page, get its content and save locally
	for _, page := range pages {
		pagePtr := &page
		if err := s.local.CreatePage(pagePtr); err != nil {
			return err
		}
	}

	return nil
}

// pushToGitHub pushes local changes to GitHub
func (s *CombinedStorage) pushToGitHub() error {
	// Get all pages from local storage
	pages, err := s.local.ListPages()
	if err != nil {
		return err
	}

	// For each page, get its content and save to GitHub
	for _, page := range pages {
		pagePtr := &page
		if err := s.github.CreatePage(pagePtr); err != nil {
			return err
		}
	}

	return nil
}

// GetPage retrieves a page from local storage
func (s *CombinedStorage) GetPage(path string) (*types.Page, error) {
	return s.local.GetPage(path)
}

// CreatePage creates a page in both local and GitHub storage
func (s *CombinedStorage) CreatePage(page *types.Page) error {
	// Create in local storage first
	if err := s.local.CreatePage(page); err != nil {
		return err
	}

	// Then create in GitHub
	return s.github.CreatePage(page)
}

// UpdatePage updates a page in both local and GitHub storage
func (s *CombinedStorage) UpdatePage(page *types.Page) error {
	// Get the old page using the old path (without .txt extension)
	oldPath := strings.TrimSuffix(page.Path, ".txt")
	oldPage, err := s.GetPage(oldPath)
	if err != nil {
		// If old page doesn't exist, treat as create
		return s.CreatePage(page)
	}

	// If the path has changed (title changed), delete the old file first
	if oldPage.Path != page.Path {
		log.Printf("Title changed from %s to %s, deleting old file", oldPage.Path, page.Path)

		// Delete old file from GitHub
		if err := s.github.DeletePage(oldPage.Path); err != nil {
			return fmt.Errorf("failed to delete old page from GitHub: %v", err)
		}

		// Delete old file from local storage
		if err := s.local.DeletePage(oldPage.Path); err != nil {
			log.Printf("Warning: Failed to delete old page from local storage: %v", err)
		}

		// Create new file in GitHub
		if err := s.github.CreatePage(page); err != nil {
			return fmt.Errorf("failed to create new page in GitHub: %v", err)
		}

		// Create new file in local storage
		if err := s.local.CreatePage(page); err != nil {
			log.Printf("Warning: Failed to create new page in local storage: %v", err)
		}

		return nil
	}

	// If path hasn't changed, just update the content
	if err := s.github.UpdatePage(page); err != nil {
		return err
	}

	if err := s.local.UpdatePage(page); err != nil {
		log.Printf("Warning: Failed to update page in local storage: %v", err)
	}

	return nil
}

// DeletePage deletes a page from both local and GitHub storage
func (s *CombinedStorage) DeletePage(path string) error {
	// Delete from local storage first
	if err := s.local.DeletePage(path); err != nil {
		return err
	}

	// Then delete from GitHub
	return s.github.DeletePage(path)
}

// ListPages lists all pages from local storage
func (s *CombinedStorage) ListPages() ([]types.Page, error) {
	return s.local.ListPages()
}

// GetPagesInFolder gets all pages in a folder from local storage
func (s *CombinedStorage) GetPagesInFolder(folderPath string) ([]types.Page, error) {
	return s.local.GetPagesInFolder(folderPath)
}

// CreateFolder creates a folder in both local and GitHub storage
func (s *CombinedStorage) CreateFolder(path string) error {
	// Create in local storage first
	if err := s.local.CreateFolder(path); err != nil {
		return err
	}

	// Then create in GitHub
	return s.github.CreateFolder(path)
}

// DeleteFolder deletes a folder from both local and GitHub storage
func (s *CombinedStorage) DeleteFolder(path string) error {
	// Delete from local storage first
	if err := s.local.DeleteFolder(path); err != nil {
		return err
	}

	// Then delete from GitHub
	return s.github.DeleteFolder(path)
}

// ListFolders lists all folders from local storage
func (s *CombinedStorage) ListFolders() ([]string, error) {
	return s.local.ListFolders()
}
