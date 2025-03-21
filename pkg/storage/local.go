package storage

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/daniel-vuky/golang-my-wiki-v2/pkg/config"
	"github.com/daniel-vuky/golang-my-wiki-v2/pkg/storage/types"
)

// LocalStorage implements the Storage interface using the local filesystem
type LocalStorage struct {
	baseDir string
}

// NewLocalStorage creates a new local storage instance
func NewLocalStorage(config *config.Config) (*LocalStorage, error) {
	return &LocalStorage{
		baseDir: config.Server.DataDir,
	}, nil
}

// ListPages retrieves all pages from the local filesystem
func (l *LocalStorage) ListPages() ([]types.Page, error) {
	var pages []types.Page
	err := filepath.Walk(l.baseDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(info.Name(), ".txt") {
			relPath, err := filepath.Rel(l.baseDir, path)
			if err != nil {
				return err
			}
			page, err := l.GetPage(relPath)
			if err != nil {
				return err
			}
			pages = append(pages, *page)
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list pages: %v", err)
	}
	return pages, nil
}

// GetPage retrieves a specific page from the local filesystem
func (l *LocalStorage) GetPage(path string) (*types.Page, error) {
	fullPath := filepath.Join(l.baseDir, path)
	content, err := ioutil.ReadFile(fullPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read page: %v", err)
	}

	title := strings.TrimSuffix(filepath.Base(path), ".txt")
	return &types.Page{
		Title:   title,
		Path:    path,
		Body:    content,
		Content: string(content),
	}, nil
}

// CreatePage creates a new page in the local filesystem
func (l *LocalStorage) CreatePage(page *types.Page) error {
	fullPath := filepath.Join(l.baseDir, page.Path)
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %v", err)
	}

	if err := ioutil.WriteFile(fullPath, page.Body, 0644); err != nil {
		return fmt.Errorf("failed to write page: %v", err)
	}

	return nil
}

// UpdatePage updates an existing page in the local filesystem
func (l *LocalStorage) UpdatePage(page *types.Page) error {
	return l.CreatePage(page)
}

// DeletePage deletes a page from the local filesystem
func (l *LocalStorage) DeletePage(path string) error {
	fullPath := filepath.Join(l.baseDir, path)
	if err := os.Remove(fullPath); err != nil {
		return fmt.Errorf("failed to delete page: %v", err)
	}
	return nil
}

// ListFolders retrieves all folders from the local filesystem
func (l *LocalStorage) ListFolders() ([]string, error) {
	var folders []string
	err := filepath.Walk(l.baseDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() && path != l.baseDir {
			relPath, err := filepath.Rel(l.baseDir, path)
			if err != nil {
				return err
			}
			folders = append(folders, relPath)
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list folders: %v", err)
	}
	return folders, nil
}

// CreateFolder creates a new folder in the local filesystem
func (l *LocalStorage) CreateFolder(path string) error {
	fullPath := filepath.Join(l.baseDir, path)
	if err := os.MkdirAll(fullPath, 0755); err != nil {
		return fmt.Errorf("failed to create folder: %v", err)
	}
	return nil
}

// DeleteFolder deletes a folder from the local filesystem
func (l *LocalStorage) DeleteFolder(path string) error {
	fullPath := filepath.Join(l.baseDir, path)
	if err := os.RemoveAll(fullPath); err != nil {
		return fmt.Errorf("failed to delete folder: %v", err)
	}
	return nil
}
