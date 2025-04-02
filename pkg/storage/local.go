package storage

import (
	"fmt"
	"io/ioutil"
	"log"
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
	// Ensure path ends with .txt
	if !strings.HasSuffix(path, ".txt") {
		path = path + ".txt"
	}

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
	if page == nil {
		return fmt.Errorf("page cannot be nil")
	}
	if page.Path == "" {
		return fmt.Errorf("page path cannot be empty")
	}

	// Ensure path ends with .txt
	if !strings.HasSuffix(page.Path, ".txt") {
		page.Path = page.Path + ".txt"
	}

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
	// Ensure path ends with .txt
	if !strings.HasSuffix(path, ".txt") {
		path = path + ".txt"
	}

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

// GetPagesInFolder retrieves all pages from a specific folder
func (l *LocalStorage) GetPagesInFolder(folderPath string) ([]types.Page, error) {
	log.Printf("=== GetPagesInFolder START: %s ===", folderPath)
	var pages []types.Page
	fullPath := filepath.Join(l.baseDir, folderPath)
	log.Printf("Looking for pages in full path: %s", fullPath)

	// Read all files in the directory
	files, err := ioutil.ReadDir(fullPath)
	if err != nil {
		log.Printf("Error reading directory: %v", err)
		return nil, fmt.Errorf("failed to read folder: %v", err)
	}

	log.Printf("Found %d files in directory", len(files))
	for _, file := range files {
		log.Printf("Checking file: %s (isDir: %v)", file.Name(), file.IsDir())
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".txt") {
			log.Printf("Found .txt file: %s", file.Name())
			relativePath := filepath.Join(folderPath, file.Name())
			page, err := l.GetPage(relativePath)
			if err != nil {
				log.Printf("Error reading page %s: %v, skipping", file.Name(), err)
				continue // Skip files that can't be read
			}

			// Add a preview of the content
			contentStr := page.Content
			if len(contentStr) > 150 {
				page.Preview = contentStr[:150] + "..."
			} else {
				page.Preview = contentStr
			}

			pages = append(pages, *page)
			log.Printf("Added page: %s", page.Title)
		} else {
			log.Printf("Skipping file: %s (not a .txt file or is a directory)", file.Name())
		}
	}

	log.Printf("=== GetPagesInFolder END: %s, found %d pages ===", folderPath, len(pages))
	return pages, nil
}

// Sync is a no-op for local storage since it doesn't need to sync with anything
func (l *LocalStorage) Sync() error {
	return nil
}
