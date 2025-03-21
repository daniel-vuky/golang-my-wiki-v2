package models

import (
	"log"
	"strings"
	"time"

	"github.com/daniel-vuky/golang-my-wiki-v2/pkg/storage/types"
)

// Page represents a wiki page
type Page struct {
	types.Page
	LastModified time.Time
}

// Save saves the page to storage
func (p *Page) Save(store types.Storage) error {
	return store.CreatePage(&p.Page)
}

// LoadPage loads a page from storage
func LoadPage(path string, store types.Storage) (*Page, error) {
	p, err := store.GetPage(path)
	if err != nil {
		return nil, err
	}

	return &Page{
		Page: *p,
	}, nil
}

// MenuItem represents an item in the side menu
type MenuItem struct {
	Title  string
	URL    string
	Active bool
}

// GetMenuItems returns the side menu items
func GetMenuItems(currentPage string, store types.Storage) []MenuItem {
	// Get all pages
	pages, err := GetAllPages(store)
	if err != nil {
		return []MenuItem{}
	}

	// Normalize the current page - strip any .txt extension
	currentPage = strings.TrimSuffix(currentPage, ".txt")

	// Debug the current page to help troubleshoot
	log.Printf("GetMenuItems: Current page is '%s'", currentPage)

	// Create menu items from existing pages
	var items []MenuItem
	for _, page := range pages {
		isActive := currentPage == page.Title
		if isActive {
			log.Printf("Found active page: %s", page.Title)
		}

		items = append(items, MenuItem{
			Title:  page.Title,
			URL:    "/view/" + page.Title,
			Active: isActive,
		})
	}

	return items
}

// GetPreview returns a preview of the page content
func (p *Page) GetPreview() string {
	// Convert markdown to plain text for preview
	text := string(p.Body)
	// Remove markdown formatting
	text = strings.ReplaceAll(text, "#", "")
	text = strings.ReplaceAll(text, "*", "")
	text = strings.ReplaceAll(text, "_", "")
	text = strings.ReplaceAll(text, "`", "")
	// Get first 150 characters
	if len(text) > 150 {
		return text[:150] + "..."
	}
	return text
}

// GetLastModified returns the last modified time of the page
func (p *Page) GetLastModified() string {
	// For GitHub storage, we don't have access to file modification time
	// So we'll return a placeholder
	return "Unknown"
}

// GetAllPages retrieves all pages from the storage
func GetAllPages(store types.Storage) ([]Page, error) {
	pages, err := store.ListPages()
	if err != nil {
		return nil, err
	}

	var result []Page
	for _, p := range pages {
		result = append(result, Page{
			Page: p,
		})
	}
	return result, nil
}

// NewPage creates a new page with the given title and content
func NewPage(title, content string) *Page {
	return &Page{
		Page: types.Page{
			Title:   title,
			Path:    title + ".txt",
			Body:    []byte(content),
			Content: content,
		},
		LastModified: time.Now(),
	}
}
