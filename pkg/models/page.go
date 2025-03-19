package models

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Page represents a wiki page
type Page struct {
	Title        string
	Body         []byte
	Preview      string
	LastModified string
}

// Save saves the page to a file
func (p *Page) Save() error {
	// Create data directory if it doesn't exist
	dataDir := filepath.Join("data")
	if _, err := os.Stat(dataDir); os.IsNotExist(err) {
		os.Mkdir(dataDir, 0755)
	}

	filename := filepath.Join(dataDir, p.Title+".txt")
	return os.WriteFile(filename, p.Body, 0600)
}

// LoadPage loads a page from a file
func LoadPage(title string) (*Page, error) {
	filename := filepath.Join("data", title+".txt")
	body, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return &Page{Title: title, Body: body}, nil
}

// MenuItem represents an item in the side menu
type MenuItem struct {
	Title  string
	URL    string
	Active bool
}

// GetMenuItems returns the side menu items
func GetMenuItems(currentPage string) []MenuItem {
	// Get all pages
	pages, err := GetAllPages()
	if err != nil {
		return []MenuItem{}
	}

	// Create menu items from existing pages
	var items []MenuItem
	for _, page := range pages {
		items = append(items, MenuItem{
			Title:  page.Title,
			URL:    "/view/" + page.Title,
			Active: currentPage == page.Title,
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
	filename := filepath.Join("data", p.Title+".txt")
	info, err := os.Stat(filename)
	if err != nil {
		return "Unknown"
	}
	return info.ModTime().Format("2006-01-02 15:04:05")
}

// GetAllPages returns a list of all wiki pages
func GetAllPages() ([]Page, error) {
	var pages []Page
	files, err := os.ReadDir("data")
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		if filepath.Ext(file.Name()) == ".txt" {
			title := strings.TrimSuffix(file.Name(), ".txt")
			page, err := LoadPage(title)
			if err != nil {
				continue // Skip files that can't be read
			}
			page.Preview = page.GetPreview()
			page.LastModified = page.GetLastModified()
			pages = append(pages, *page)
		}
	}

	// Sort pages by title
	sort.Slice(pages, func(i, j int) bool {
		return pages[i].Title < pages[j].Title
	})

	return pages, nil
}
