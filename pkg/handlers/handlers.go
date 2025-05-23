package handlers

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"sort"
	"strings"

	"github.com/daniel-vuky/golang-my-wiki-v2/pkg/config"
	"github.com/daniel-vuky/golang-my-wiki-v2/pkg/models"
	"github.com/daniel-vuky/golang-my-wiki-v2/pkg/storage"
	"github.com/daniel-vuky/golang-my-wiki-v2/pkg/storage/types"
	"github.com/gin-gonic/gin"
)

var (
	store types.Storage
)

// InitHandlers initializes the handlers with the given storage
func InitHandlers(s types.Storage) {
	store = s
}

// GetStorage returns the current storage instance
func GetStorage() types.Storage {
	return store
}

// HomeHandler handles the home page
func HomeHandler(c *gin.Context) {
	log.Println("=== HomeHandler START ===")

	// Get all categories (folders)
	allFolders, err := store.ListFolders()
	if err != nil {
		log.Printf("Error getting categories: %v", err)
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{
			"error": fmt.Sprintf("Failed to get categories: %v", err),
		})
		return
	}

	// Filter to only show root-level folders
	var rootFolders []string
	for _, folder := range allFolders {
		// Only include folders that don't have a slash (not children)
		if !strings.Contains(folder, "/") {
			rootFolders = append(rootFolders, folder)
		}
	}

	log.Printf("Found %d root categories", len(rootFolders))
	for _, category := range rootFolders {
		log.Printf("Root category: %s", category)
	}

	user := c.MustGet("user")
	log.Printf("User: %+v", user)

	c.HTML(http.StatusOK, "home.html", gin.H{
		"Categories": rootFolders,
		"User":       user,
		"SideMenu":   models.GetMenuItems("", store),
	})

	log.Println("=== HomeHandler END ===")
}

// ViewHandler handles viewing a page
func ViewHandler(c *gin.Context) {
	title := c.Param("title")
	log.Printf("=== ViewHandler START: %s ===", title)

	// URL decode the title
	decodedTitle, err := url.QueryUnescape(title)
	if err != nil {
		log.Printf("Error decoding title: %v", err)
		c.HTML(http.StatusBadRequest, "error.html", gin.H{
			"error": "Invalid page title",
		})
		return
	}

	// Get folder parameter (if viewing from a folder)
	folderPath := c.Query("folder")
	if folderPath != "" {
		// URL decode the folder path
		decodedFolderPath, err := url.QueryUnescape(folderPath)
		if err != nil {
			log.Printf("Error decoding folder path: %v", err)
			c.HTML(http.StatusBadRequest, "error.html", gin.H{
				"error": "Invalid folder path",
			})
			return
		}
		folderPath = decodedFolderPath
	}

	// If folder is specified, prepend it to the title path
	var fullPath string
	if folderPath != "" {
		fullPath = folderPath + "/" + decodedTitle
		log.Printf("Viewing page in folder: %s", fullPath)
	} else {
		fullPath = decodedTitle
		log.Printf("Viewing page at root: %s", fullPath)
	}

	// Try to get the page
	page, err := store.GetPage(fullPath)
	if err != nil {
		log.Printf("Error getting page: %v", err)
		c.HTML(http.StatusNotFound, "error.html", gin.H{
			"error": "Page not found",
		})
		return
	}

	log.Printf("Page title: %s", page.Title)
	log.Printf("Page content length: %d", len(page.Content))
	log.Printf("Page content preview: %s", page.Content[:min(100, len(page.Content))])

	// Get the folder tree to use in the sidebar
	folderTree, err := GetFolderTree(store, folderPath)
	if err != nil {
		log.Printf("Error building folder tree: %v", err)
		// Continue without folder tree - not critical
	}

	// Create breadcrumbs for the view
	var breadcrumbs []map[string]string
	if folderPath != "" {
		// Add the home breadcrumb
		breadcrumbs = append(breadcrumbs, map[string]string{
			"name": "Home",
			"path": "",
		})

		// Build breadcrumbs for folder path
		parts := strings.Split(folderPath, "/")
		partialPath := ""
		for _, part := range parts {
			if partialPath != "" {
				partialPath += "/"
			}
			partialPath += part
			breadcrumbs = append(breadcrumbs, map[string]string{
				"name": part,
				"path": partialPath,
			})
		}
	}

	c.HTML(http.StatusOK, "view.html", gin.H{
		"Title":       page.Title,
		"Content":     page.Content,
		"FolderTree":  folderTree,
		"FolderPath":  folderPath,
		"CurrentPath": folderPath, // For highlighting the active folder
		"Breadcrumbs": breadcrumbs,
		"User":        c.MustGet("user"),
	})
	log.Printf("=== ViewHandler END: %s ===", title)
}

// EditHandler handles editing a page
func EditHandler(c *gin.Context) {
	title := c.Param("title")
	log.Printf("=== EditHandler START: %s ===", title)

	// URL decode the title
	decodedTitle, err := url.QueryUnescape(title)
	if err != nil {
		log.Printf("Error decoding title: %v", err)
		c.HTML(http.StatusBadRequest, "error.html", gin.H{
			"error": "Invalid page title",
		})
		return
	}

	// Get folder parameter (if creating from a folder)
	folderPath := c.Query("folder")
	if folderPath != "" {
		// URL decode the folder path
		decodedFolderPath, err := url.QueryUnescape(folderPath)
		if err != nil {
			log.Printf("Error decoding folder path: %v", err)
			c.HTML(http.StatusBadRequest, "error.html", gin.H{
				"error": "Invalid folder path",
			})
			return
		}
		folderPath = decodedFolderPath
	}
	log.Printf("Folder path from query: %s", folderPath)

	// Get the folder tree to use in the sidebar
	folderTree, err := GetFolderTree(store, folderPath)
	if err != nil {
		log.Printf("Error building folder tree: %v", err)
		// Continue without folder tree - not critical
	}

	// Create breadcrumbs for the edit page
	var breadcrumbs []map[string]string
	if folderPath != "" {
		// Add the home breadcrumb
		breadcrumbs = append(breadcrumbs, map[string]string{
			"name": "Home",
			"path": "",
		})

		// Build breadcrumbs for folder path
		parts := strings.Split(folderPath, "/")
		partialPath := ""
		for _, part := range parts {
			if partialPath != "" {
				partialPath += "/"
			}
			partialPath += part
			breadcrumbs = append(breadcrumbs, map[string]string{
				"name": part,
				"path": partialPath,
			})
		}
	}

	// Handle new page creation
	if decodedTitle == "" || decodedTitle == "new" {
		log.Printf("Creating new page form")
		c.HTML(http.StatusOK, "edit.html", gin.H{
			"Title":       "",
			"Content":     "",
			"IsNewPage":   true,
			"FolderPath":  folderPath, // Pass the folder path to the template
			"FolderTree":  folderTree,
			"CurrentPath": folderPath,  // For highlighting the active folder
			"Breadcrumbs": breadcrumbs, // Add breadcrumbs
			"User":        c.MustGet("user"),
		})
		log.Printf("=== EditHandler END (new page) ===")
		return
	}

	// Editing an existing page - GetPage will handle adding .txt
	log.Printf("Attempting to edit page: %s", decodedTitle)

	// If folder is specified, prepend it to the title path
	var fullPath string
	if folderPath != "" {
		fullPath = folderPath + "/" + decodedTitle
		log.Printf("Editing page in folder: %s", fullPath)
	} else {
		fullPath = decodedTitle
	}

	page, err := store.GetPage(fullPath)
	if err != nil {
		log.Printf("Error getting page: %v", err)
		c.HTML(http.StatusNotFound, "error.html", gin.H{
			"error": fmt.Sprintf("Page not found: %v", err),
		})
		log.Printf("=== EditHandler END (error) ===")
		return
	}

	log.Printf("Found page, title: %s, content length: %d", page.Title, len(page.Content))
	c.HTML(http.StatusOK, "edit.html", gin.H{
		"Title":       page.Title,
		"Content":     page.Content,
		"IsNewPage":   false,
		"FolderPath":  folderPath, // Pass the folder path to the template
		"FolderTree":  folderTree,
		"CurrentPath": folderPath,  // For highlighting the active folder
		"Breadcrumbs": breadcrumbs, // Add breadcrumbs
		"User":        c.MustGet("user"),
	})
	log.Printf("=== EditHandler END (success) ===")
}

// SaveHandler handles saving a page
func SaveHandler(c *gin.Context) {
	log.Println("=== SaveHandler START ===")

	// Parse JSON request body
	var requestBody struct {
		Title    string `json:"title"`
		Content  string `json:"content"`
		Folder   string `json:"folder"`
		OldTitle string `json:"oldTitle"` // Add oldTitle to track title changes
	}

	if err := c.ShouldBindJSON(&requestBody); err != nil {
		log.Printf("Error binding JSON: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf("Failed to parse request: %v", err),
		})
		return
	}

	title := requestBody.Title
	content := requestBody.Content
	folderPath := requestBody.Folder
	oldTitle := requestBody.OldTitle

	log.Printf("Title from JSON: %s", title)
	log.Printf("Old title from JSON: %s", oldTitle)
	log.Printf("Content from JSON: %s", content)
	log.Printf("Folder from JSON: %s", folderPath)

	if title == "" {
		log.Println("Error: Title is empty")
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Title is required",
		})
		return
	}

	// Determine the full path based on folder
	var filePath string
	if folderPath != "" {
		filePath = folderPath + "/" + title
		log.Printf("Creating/updating page in folder: %s", filePath)
	} else {
		filePath = title
		log.Printf("Creating/updating page at root: %s", filePath)
	}

	// Create a page object with the new title and path
	page := &types.Page{
		Title:   title,
		Path:    filePath,
		Content: content,
		Body:    []byte(content),
	}

	// If we have an old title, try to get the old page first
	if oldTitle != "" {
		var oldFilePath string
		if folderPath != "" {
			oldFilePath = folderPath + "/" + oldTitle
		} else {
			oldFilePath = oldTitle
		}
		log.Printf("Checking for old page at: %s", oldFilePath)

		oldPage, err := store.GetPage(oldFilePath)
		if err == nil {
			// Old page exists, delete it first
			log.Printf("Found old page, deleting: %s", oldFilePath)
			if err := store.DeletePage(oldPage.Path); err != nil {
				log.Printf("Error deleting old page: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": fmt.Sprintf("Failed to delete old page: %v", err),
				})
				return
			}
			log.Printf("Successfully deleted old page: %s", oldFilePath)

			// Now create the new page
			log.Printf("Creating new page: %s", filePath)
			if err := store.CreatePage(page); err != nil {
				log.Printf("Error creating new page: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": fmt.Sprintf("Failed to create new page: %v", err),
				})
				return
			}
			log.Printf("Successfully created new page: %s", filePath)
		} else {
			// Old page doesn't exist, create new one
			log.Printf("Old page not found, creating new page: %s", filePath)
			if err := store.CreatePage(page); err != nil {
				log.Printf("Error creating page: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": fmt.Sprintf("Failed to create page: %v", err),
				})
				return
			}
			log.Printf("Successfully created new page: %s", filePath)
		}
	} else {
		// No old title, check if page exists at new path
		_, err := store.GetPage(filePath)
		if err != nil {
			// Page doesn't exist, create it
			log.Printf("Page doesn't exist, creating new page: %s", filePath)
			if err := store.CreatePage(page); err != nil {
				log.Printf("Error creating page: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": fmt.Sprintf("Failed to create page: %v", err),
				})
				return
			}
			log.Printf("Successfully created new page: %s", filePath)
		} else {
			// Page exists, update it
			log.Printf("Page exists, updating: %s", filePath)
			if err := store.UpdatePage(page); err != nil {
				log.Printf("Error updating page: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": fmt.Sprintf("Failed to update page: %v", err),
				})
				return
			}
			log.Printf("Successfully updated page: %s", filePath)
		}
	}

	log.Println("=== SaveHandler END ===")
	// Include folder path in redirect URL if present
	redirectURL := "/view/" + url.QueryEscape(title)
	if folderPath != "" {
		redirectURL += "?folder=" + url.QueryEscape(folderPath)
	}
	c.JSON(http.StatusOK, gin.H{
		"redirect": redirectURL,
	})
}

// DeleteHandler handles deleting a page
func DeleteHandler(c *gin.Context) {
	title := c.Param("title")
	log.Printf("=== DeleteHandler START: %s ===", title)

	// Get folder parameter (if deleting from a folder)
	folderPath := c.Query("folder")
	if folderPath != "" {
		// URL decode the folder path
		decodedFolderPath, err := url.QueryUnescape(folderPath)
		if err != nil {
			log.Printf("Error decoding folder path: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   "Invalid folder path",
			})
			return
		}
		folderPath = decodedFolderPath
	}

	// If folder is specified, prepend it to the title path
	var fullPath string
	if folderPath != "" {
		fullPath = folderPath + "/" + title
		log.Printf("Deleting page in folder: %s", fullPath)
	} else {
		fullPath = title
		log.Printf("Deleting page at root: %s", fullPath)
	}

	if err := store.DeletePage(fullPath); err != nil {
		log.Printf("Error deleting page: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   fmt.Sprintf("Failed to delete page: %v", err),
		})
		return
	}

	log.Printf("Successfully deleted page: %s", fullPath)
	log.Printf("=== DeleteHandler END: %s ===", title)

	// Return success response with redirect URL
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"redirect": func() string {
			if folderPath != "" {
				return "/category/" + url.QueryEscape(folderPath)
			}
			return "/"
		}(),
	})
}

// CategoryCreateHandler handles creating a new category
func CategoryCreateHandler(c *gin.Context) {
	log.Println("=== CategoryCreateHandler START ===")

	// Parse request body
	var requestBody struct {
		Name       string `json:"name"`
		ParentPath string `json:"parentPath"`
		SHA        string `json:"sha"`
	}

	log.Println("Attempting to bind JSON")
	if err := c.ShouldBindJSON(&requestBody); err != nil {
		log.Printf("Error binding JSON: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf("Failed to parse request: %v", err),
		})
		return
	}

	categoryName := requestBody.Name
	parentFolder := requestBody.ParentPath
	parentFolderSha := requestBody.SHA

	log.Printf("Received category name: %s", categoryName)
	log.Printf("Parent folder: %s", parentFolder)
	log.Printf("Parent folder SHA: %s", parentFolderSha)

	if categoryName == "" {
		log.Println("Error: Category name is empty")
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Category name is required",
		})
		return
	}

	// Determine full path based on parent
	var fullPath string
	if parentFolder != "" {
		// Check if parent folder exists by checking the list of folders
		allFolders, err := store.ListFolders()
		if err != nil {
			log.Printf("Error getting folders: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": fmt.Sprintf("Failed to get folders: %v", err),
			})
			return
		}

		parentExists := false
		for _, folder := range allFolders {
			if folder == parentFolder {
				parentExists = true
				break
			}
		}

		if !parentExists {
			log.Printf("Parent folder does not exist: %s", parentFolder)
			c.JSON(http.StatusBadRequest, gin.H{
				"error": fmt.Sprintf("Parent folder '%s' does not exist", parentFolder),
			})
			return
		}
		fullPath = parentFolder + "/" + categoryName
	} else {
		fullPath = categoryName
	}

	// Get max category level from config
	maxLevel := config.GetMaxCategoryLevel()

	// Check category nesting level
	if fullPath != "" {
		levels := strings.Count(fullPath, "/") + 1
		log.Printf("Category level: %d (max %d)", levels, maxLevel)

		if levels > maxLevel {
			log.Printf("Error: Maximum category nesting level reached (%d)", maxLevel)
			c.JSON(http.StatusBadRequest, gin.H{
				"error": fmt.Sprintf("Maximum category nesting level reached (%d levels max)", maxLevel),
			})
			return
		}
	}

	log.Printf("Creating category at path: %s", fullPath)

	// Create the category folder
	if err := store.CreateFolder(fullPath); err != nil {
		log.Printf("Error creating category: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("Failed to create category: %v", err),
		})
		return
	}

	// Force refresh the folder list by invalidating cache
	if cacheable, ok := store.(interface{ InvalidateCache() error }); ok {
		if err := cacheable.InvalidateCache(); err != nil {
			log.Printf("Warning: Failed to invalidate cache: %v", err)
		}
	}

	log.Printf("Successfully created category: %s", fullPath)
	log.Println("=== CategoryCreateHandler END ===")

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": fmt.Sprintf("Category '%s' created successfully", categoryName),
	})
}

// DeleteFolderHandler handles deleting a folder
func DeleteFolderHandler(c *gin.Context) {
	path := c.Query("path")
	if path == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Folder path is required",
		})
		return
	}

	log.Printf("=== DeleteFolderHandler START: %s ===", path)

	// Get parent folder path for redirect before deleting
	parentPath := getParentPath(path)
	redirectURL := "/"
	if parentPath != "" {
		redirectURL = "/category/" + url.QueryEscape(parentPath)
	}

	// Delete the folder
	if err := store.DeleteFolder(path); err != nil {
		log.Printf("Error deleting folder: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("Failed to delete folder: %v", err),
		})
		return
	}

	// Force refresh the folder list by invalidating cache
	if cacheable, ok := store.(interface{ InvalidateCache() error }); ok {
		if err := cacheable.InvalidateCache(); err != nil {
			log.Printf("Warning: Failed to invalidate cache: %v", err)
		}
	}

	log.Printf("Successfully deleted folder: %s", path)
	log.Printf("=== DeleteFolderHandler END ===")

	// Return success response with redirect URL
	c.JSON(http.StatusOK, gin.H{
		"success":  true,
		"message":  fmt.Sprintf("Folder '%s' deleted successfully", path),
		"redirect": redirectURL,
	})
}

// FolderTreeItem represents an item in the folder tree
type FolderTreeItem struct {
	Name        string
	Path        string
	HasChildren bool
	IsExpanded  bool
	Children    []FolderTreeItem
	IsNote      bool
}

// CategoryHandler handles viewing a category/folder
func CategoryHandler(c *gin.Context) {
	pathParam := c.Param("path")

	// When using *path, the param includes the leading slash,
	// so we need to trim it to get the actual path
	path := strings.TrimPrefix(pathParam, "/")

	log.Printf("=== CategoryHandler START: %s (from param: %s) ===", path, pathParam)

	// Force refresh parameter check
	forceRefresh := c.Query("refresh") == "true"
	if forceRefresh {
		log.Printf("Force refresh requested, invalidating folder cache")
		// Access the storage as CachedGitHubStorage to invalidate cache
		if cacheable, ok := store.(interface{ InvalidateCache() error }); ok {
			if err := cacheable.InvalidateCache(); err != nil {
				log.Printf("Error invalidating cache: %v", err)
			}
		}
	}

	// Get the folder tree using our helper function
	folderTree, err := GetFolderTree(store, path)
	if err != nil {
		log.Printf("Error building folder tree: %v", err)
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{
			"error": fmt.Sprintf("Failed to build folder tree: %v", err),
		})
		return
	}

	// Get the pages in this folder
	notes, err := store.GetPagesInFolder(path)
	if err != nil {
		log.Printf("Error getting pages in folder %s: %v", path, err)
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{
			"error": fmt.Sprintf("Failed to get pages in folder: %v", err),
		})
		return
	}

	log.Printf("Found %d notes in folder %s", len(notes), path)
	for _, note := range notes {
		log.Printf("Note: %s, Path: %s", note.Title, note.Path)
	}

	// Get all folders
	allFolders, err := store.ListFolders()
	if err != nil {
		log.Printf("Error getting folders: %v", err)
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{
			"error": fmt.Sprintf("Failed to get folders: %v", err),
		})
		return
	}

	// Get the subfolders for this folder
	subFolders := getDirectChildren(allFolders, path)

	// Set folder name - use last part of path or "Home" for root
	folderName := getNameFromPath(path)
	if path == "" {
		folderName = "Home"
	}

	// Get breadcrumb trail for this folder
	breadcrumbs := getBreadcrumbs(path)

	// Calculate current nesting level
	currentLevel := 0
	if path != "" {
		currentLevel = strings.Count(path, "/") + 1
	}

	// Get max level from config
	maxLevel := config.GetMaxCategoryLevel()

	// Check if at max nesting level - only if at max level
	isMaxLevel := currentLevel >= maxLevel

	// Get parent folder's SHA if this is a subfolder
	var parentFolderSha string
	if path != "" {
		parentPath := getParentPath(path)
		if parentPath != "" {
			// Use GetPage to check if the parent folder exists and get its SHA
			folderPath := parentPath + "/.folder"
			page, err := store.GetPage(folderPath)
			if err == nil && page != nil {
				// Get the SHA from the page content
				parentFolderSha = string(page.Content)
			}
		}
	}

	log.Printf("Loading folder view for %s with %d subfolders and %d notes (level: %d, max: %v, maxLevel: %d)",
		path, len(subFolders), len(notes), currentLevel, isMaxLevel, maxLevel)

	c.HTML(http.StatusOK, "folder.html", gin.H{
		"FolderName":      folderName,
		"FolderPath":      path,
		"SubFolders":      subFolders,
		"Notes":           notes,
		"FolderTree":      folderTree,
		"CurrentPath":     path,
		"Breadcrumbs":     breadcrumbs,
		"CurrentLevel":    currentLevel,
		"MaxLevel":        maxLevel,
		"MaxLevelReached": isMaxLevel,
		"ParentFolderSha": parentFolderSha,
		"User": gin.H{
			"Name": "Admin",
		},
	})

	log.Printf("=== CategoryHandler END ===")
}

// getBreadcrumbs creates a breadcrumb trail for a folder path
func getBreadcrumbs(path string) []map[string]string {
	if path == "" {
		return []map[string]string{
			{"name": "Home", "path": ""},
		}
	}

	parts := strings.Split(path, "/")
	breadcrumbs := make([]map[string]string, len(parts)+1)

	// Add home as first breadcrumb
	breadcrumbs[0] = map[string]string{
		"name": "Home",
		"path": "",
	}

	// Build up the path parts
	currentPath := ""
	for i, part := range parts {
		if i > 0 {
			currentPath += "/"
		}
		currentPath += part

		breadcrumbs[i+1] = map[string]string{
			"name": part,
			"path": currentPath,
		}
	}

	return breadcrumbs
}

// Helper functions for folder tree
func getNameFromPath(path string) string {
	parts := strings.Split(path, "/")
	return parts[len(parts)-1]
}

func getParentPath(path string) string {
	lastSlashIndex := strings.LastIndex(path, "/")
	if lastSlashIndex == -1 {
		log.Printf("Path %s has no parent (no slash found)", path)
		return ""
	}
	parentPath := path[:lastSlashIndex]
	log.Printf("Parent of %s is %s", path, parentPath)
	return parentPath
}

// Get the store from the context
func getStoreFromContext(c *gin.Context) (types.Storage, error) {
	storeVal, exists := c.Get("store")
	if !exists {
		return nil, fmt.Errorf("store not found in context")
	}
	store, ok := storeVal.(types.Storage)
	if !ok {
		return nil, fmt.Errorf("store is not of type Storage")
	}
	return store, nil
}

// Check if a folder has any children (subfolders or notes)
func hasChildren(folders []string, path string) bool {
	log.Printf("Checking if folder '%s' has children...", path)
	prefix := path + "/"
	log.Printf("Looking for prefix: '%s'", prefix)

	// First check for subfolder children
	for _, folder := range folders {
		if strings.HasPrefix(folder, prefix) {
			log.Printf("Found child folder: '%s' is a child of '%s'", folder, path)
			return true
		}
	}

	// For final level folders (no subfolders), check for notes
	notes, err := store.GetPagesInFolder(path)
	if err != nil {
		log.Printf("Error checking for notes in folder '%s': %v", path, err)
		return false
	}
	if len(notes) > 0 {
		log.Printf("Found %d notes in folder '%s':", len(notes), path)
		for _, note := range notes {
			log.Printf("  - %s", note.Title)
		}
		return true
	}

	log.Printf("No children (folders or notes) found for '%s'", path)
	return false
}

// Check if a folder might have notes - this will be used in GetFolderChildrenHandler
func hasNotes(path string, store types.Storage) bool {
	log.Printf("Checking if folder '%s' has notes...", path)
	notes, err := store.GetPagesInFolder(path)
	if err != nil {
		log.Printf("Error checking for notes in folder '%s': %v", path, err)
		return false
	}
	hasNotes := len(notes) > 0
	if hasNotes {
		log.Printf("Found %d notes in folder '%s':", len(notes), path)
		for _, note := range notes {
			log.Printf("  - %s", note.Title)
		}
	} else {
		log.Printf("No notes found in folder '%s'", path)
	}
	return hasNotes
}

func isParentOf(folder, path string) bool {
	return strings.HasPrefix(path, folder+"/")
}

func getDirectChildren(folders []string, parentPath string) []FolderTreeItem {
	var children []FolderTreeItem
	for _, folder := range folders {
		// Only get direct children (parent/child), not deeper descendants (parent/child/grandchild)
		if getParentPath(folder) == parentPath {
			// Check if folder has subfolder children
			hasSubfolders := hasChildren(folders, folder)

			// Check if folder has note children
			hasNotes := hasNotes(folder, store)

			// A folder has children if it has either subfolders or notes
			hasAnyChildren := hasSubfolders || hasNotes

			log.Printf("Folder '%s' has subfolders: %v, has notes: %v, has any children: %v",
				folder, hasSubfolders, hasNotes, hasAnyChildren)

			children = append(children, FolderTreeItem{
				Name:        getNameFromPath(folder),
				Path:        folder,
				HasChildren: hasAnyChildren,
				IsNote:      false, // This is a folder
			})
		}
	}
	return children
}

func buildFolderSubtree(folders []string, rootFolder string, path string) []FolderTreeItem {
	var children []FolderTreeItem
	for _, folder := range folders {
		if getParentPath(folder) == rootFolder {
			childItem := FolderTreeItem{
				Name:        getNameFromPath(folder),
				Path:        folder,
				HasChildren: hasChildren(folders, folder),
				IsExpanded:  isParentOf(folder, path),
				IsNote:      false, // This is a folder
			}
			if childItem.IsExpanded {
				childItem.Children = buildFolderSubtree(folders, folder, path)
			}
			children = append(children, childItem)
		}
	}
	return children
}

// GetFolderTree retrieves the folder tree structure for the sidebar
func GetFolderTree(store storage.Storage, currentPath string) ([]FolderTreeItem, error) {
	// Get all folders
	allFolders, err := store.ListFolders()
	if err != nil {
		return nil, err
	}

	// Create a map of folders to determine parent-child relationships
	folderPathsMap := make(map[string]bool)
	for _, folder := range allFolders {
		folderPathsMap[folder] = true
	}

	// Get root folders (those without a parent)
	var rootItems []FolderTreeItem

	// First, add root-level notes
	rootNotes, err := store.GetPagesInFolder("")
	if err != nil {
		log.Printf("Error getting root notes: %v", err)
	} else {
		for _, note := range rootNotes {
			noteItem := FolderTreeItem{
				Name:        note.Title,
				Path:        note.Path,
				HasChildren: false,
				IsExpanded:  false,
				Children:    []FolderTreeItem{},
				IsNote:      true,
			}
			rootItems = append(rootItems, noteItem)
		}
	}

	// Then add root folders
	for _, folder := range allFolders {
		parts := strings.Split(folder, "/")
		if len(parts) == 1 {
			// This is a root folder
			name := parts[0]

			// Check if it has children - any folder that starts with this folder + "/"
			hasChildrenVal := hasChildren(allFolders, folder)

			// Create folder tree item - collapsed by default
			item := FolderTreeItem{
				Name:        name,
				Path:        folder,
				HasChildren: hasChildrenVal,
				IsExpanded:  isParentOf(folder, currentPath),
				Children:    []FolderTreeItem{},
				IsNote:      false,
			}

			// If this folder is expanded, get its children
			if item.IsExpanded {
				item.Children = buildFolderSubtree(allFolders, folder, currentPath)
			}

			rootItems = append(rootItems, item)
		}
	}

	// Sort root items alphabetically - folders first, then notes
	sort.Slice(rootItems, func(i, j int) bool {
		// If one is a folder and one is a note, folder comes first
		if rootItems[i].IsNote != rootItems[j].IsNote {
			return !rootItems[i].IsNote
		}
		// Otherwise sort alphabetically
		return strings.ToLower(rootItems[i].Name) < strings.ToLower(rootItems[j].Name)
	})

	return rootItems, nil
}

// GetFolderChildrenHandler returns the children of a specific folder as JSON
func GetFolderChildrenHandler(c *gin.Context) {
	pathParam := c.Param("path")

	// Trim the leading slash from the path parameter
	parentPath := strings.TrimPrefix(pathParam, "/")

	log.Printf("=== GetFolderChildrenHandler START: %s (from param: %s) ===", parentPath, pathParam)

	// Get all folders
	allFolders, err := store.ListFolders()
	if err != nil {
		log.Printf("Error getting folders: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("Failed to get folders: %v", err),
		})
		return
	}

	// Get direct children folders of the specified folder
	var children []FolderTreeItem
	for _, folder := range allFolders {
		// Only include direct children
		if getParentPath(folder) == parentPath {
			// Check if folder has subfolder children
			hasSubfolders := hasChildren(allFolders, folder)

			// Check if folder has note children (if no subfolder children)
			hasAnyChildren := hasSubfolders
			if !hasSubfolders {
				hasAnyChildren = hasNotes(folder, store)
			}

			child := FolderTreeItem{
				Name:        getNameFromPath(folder),
				Path:        folder,
				HasChildren: hasAnyChildren,
				IsExpanded:  false,
				Children:    []FolderTreeItem{}, // Empty children array
				IsNote:      false,              // This is a folder
			}
			children = append(children, child)
		}
	}

	// Get notes in this folder
	notes, err := store.GetPagesInFolder(parentPath)
	if err != nil {
		log.Printf("Error getting notes in folder: %v", err)
		// Continue even if we couldn't get notes
	} else {
		// Add notes as children
		for _, note := range notes {
			// Extract just the filename without .txt extension
			noteName := note.Title

			noteItem := FolderTreeItem{
				Name:        noteName,
				Path:        note.Path,
				HasChildren: false, // Notes never have children
				IsExpanded:  false,
				Children:    []FolderTreeItem{},
				IsNote:      true, // This is a note, not a folder
			}
			children = append(children, noteItem)
		}
	}

	// Sort children alphabetically - put folders first, then notes
	sort.Slice(children, func(i, j int) bool {
		// If one is a folder and one is a note, folder comes first
		if children[i].IsNote != children[j].IsNote {
			return !children[i].IsNote // Folders come first
		}
		// Otherwise sort alphabetically
		return strings.ToLower(children[i].Name) < strings.ToLower(children[j].Name)
	})

	log.Printf("Found %d children (folders and notes) for folder %s", len(children), parentPath)
	log.Printf("=== GetFolderChildrenHandler END ===")

	c.JSON(http.StatusOK, gin.H{
		"children": children,
	})
}

// HandleSync handles the sync operation between local and GitHub storage
func HandleSync(c *gin.Context) {
	if c.Request.Method != http.MethodPost {
		c.JSON(http.StatusMethodNotAllowed, gin.H{"error": "Method not allowed"})
		return
	}

	store := GetStorage()
	if store == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Storage not initialized"})
		return
	}

	// Check if storage is CombinedStorage
	combinedStore, ok := store.(*storage.CombinedStorage)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Storage is not configured for sync"})
		return
	}

	err := combinedStore.Sync()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Sync completed successfully"})
}
