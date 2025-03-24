package types

// Page represents a wiki page
type Page struct {
	Title        string
	Path         string
	Body         []byte
	Content      string
	Preview      string
	LastModified string
}

// Storage defines the interface for different storage backends
type Storage interface {
	// Page operations
	ListPages() ([]Page, error)
	GetPage(path string) (*Page, error)
	CreatePage(page *Page) error
	UpdatePage(page *Page) error
	DeletePage(path string) error
	GetPagesInFolder(folderPath string) ([]Page, error)

	// Folder operations
	ListFolders() ([]string, error)
	CreateFolder(path string) error
	DeleteFolder(path string) error
}
