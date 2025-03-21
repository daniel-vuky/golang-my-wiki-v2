package handlers

import (
	"fmt"
	"log"
	"net/http"
	"sort"
	"strings"

	"github.com/daniel-vuky/golang-my-wiki-v2/pkg/config"
	"github.com/daniel-vuky/golang-my-wiki-v2/pkg/models"
	"github.com/daniel-vuky/golang-my-wiki-v2/pkg/storage"
	"github.com/daniel-vuky/golang-my-wiki-v2/pkg/storage/types"
	"github.com/gin-gonic/gin"
)

var store types.Storage

// InitHandlers initializes the handlers with the storage instance
func InitHandlers(s types.Storage) {
	store = s
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

	// Get folder parameter (if viewing from a folder)
	folderPath := c.Query("folder")

	// If folder is specified, prepend it to the title path
	var fullPath string
	if folderPath != "" {
		fullPath = folderPath + "/" + title
		log.Printf("Viewing page in folder: %s", fullPath)
	} else {
		fullPath = title
	}

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

	// Get folder parameter (if creating from a folder)
	folderPath := c.Query("folder")
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
	if title == "" || title == "new" {
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
	log.Printf("Attempting to edit page: %s", title)

	// If folder is specified, prepend it to the title path
	var fullPath string
	if folderPath != "" {
		fullPath = folderPath + "/" + title
		log.Printf("Editing page in folder: %s", fullPath)
	} else {
		fullPath = title
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

	title := c.PostForm("title")
	log.Printf("Title from form: %s", title)

	content := c.PostForm("content")
	log.Printf("Content from form: %s", content)

	// Get the original title (before edit)
	originalTitle := c.PostForm("original_title")
	log.Printf("Original title: %s", originalTitle)

	// Get the folder path if present
	folderPath := c.PostForm("folder_path")
	log.Printf("Folder path from form: %s", folderPath)

	if title == "" || content == "" {
		log.Println("Error: Title or content is empty")
		c.HTML(http.StatusBadRequest, "error.html", gin.H{
			"error": "Title and content are required",
		})
		return
	}

	// Determine the full path based on folder
	var filePath string
	if folderPath != "" {
		// If folder path is provided, create the page in that folder
		filePath = folderPath + "/" + title + ".txt"
		log.Printf("Creating page in folder: %s", filePath)
	} else {
		// Otherwise, create at root level
		filePath = title + ".txt"
		log.Printf("Creating page at root: %s", filePath)
	}

	// Create a page object with the new title and path
	page := &types.Page{
		Title:   title,
		Path:    filePath,
		Content: content,
		Body:    []byte(content),
	}

	// If we're editing an existing page and the title has changed
	if originalTitle != "" && originalTitle != title {
		log.Printf("Title changed from %s to %s, creating new file and deleting old one", originalTitle, title)

		// Create new file with new title
		if err := store.CreatePage(page); err != nil {
			log.Printf("Error creating new page: %v", err)
			c.HTML(http.StatusInternalServerError, "error.html", gin.H{
				"error": fmt.Sprintf("Failed to create new page: %v", err),
			})
			return
		}
		log.Printf("Successfully created new page: %s", title)

		// Delete old file with old title
		if err := store.DeletePage(originalTitle); err != nil {
			log.Printf("Warning: Failed to delete old page %s: %v", originalTitle, err)
			// Continue even if delete fails
		} else {
			log.Printf("Successfully deleted old page: %s", originalTitle)
		}
	} else {
		// Normal save without title change
		// Check if the page exists
		log.Printf("No title change detected, checking if page exists: %s", title)
		existingPage, err := store.GetPage(title)

		if err != nil {
			// Page doesn't exist, create it
			log.Printf("Page %s does not exist, creating new page", title)

			if err := store.CreatePage(page); err != nil {
				log.Printf("Error creating page: %v", err)
				c.HTML(http.StatusInternalServerError, "error.html", gin.H{
					"error": fmt.Sprintf("Failed to create page: %v", err),
				})
				return
			}
			log.Printf("Successfully created new page: %s", title)
		} else {
			// Page exists, update it
			log.Printf("Page %s exists, updating content", title)
			page.LastModified = existingPage.LastModified

			if err := store.UpdatePage(page); err != nil {
				log.Printf("Error updating page: %v", err)
				c.HTML(http.StatusInternalServerError, "error.html", gin.H{
					"error": fmt.Sprintf("Failed to update page: %v", err),
				})
				return
			}
			log.Printf("Successfully updated page: %s", title)
		}
	}

	log.Println("=== SaveHandler END ===")
	c.Redirect(http.StatusFound, "/view/"+title)
}

// DeleteHandler handles deleting a page
func DeleteHandler(c *gin.Context) {
	title := c.Param("title")
	log.Printf("=== DeleteHandler START: %s ===", title)

	// No need to add .txt here, the storage will handle it
	log.Printf("Attempting to delete page: %s", title)

	if err := store.DeletePage(title); err != nil {
		log.Printf("Error deleting page: %v", err)
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{
			"error": fmt.Sprintf("Failed to delete page: %v", err),
		})
		return
	}

	log.Printf("Successfully deleted page: %s", title)
	log.Printf("=== DeleteHandler END: %s ===", title)
	c.Redirect(http.StatusFound, "/")
}

// CategoryCreateHandler handles creating a new category
func CategoryCreateHandler(c *gin.Context) {
	log.Println("=== CategoryCreateHandler START ===")

	// Parse request body
	var requestBody struct {
		Name   string `json:"name"`
		Parent string `json:"parent"`
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
	parentFolder := requestBody.Parent

	log.Printf("Received category name: %s", categoryName)
	log.Printf("Parent folder: %s", parentFolder)

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

	log.Printf("Successfully created category: %s", fullPath)
	log.Println("=== CategoryCreateHandler END ===")

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": fmt.Sprintf("Category '%s' created successfully", categoryName),
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

	log.Printf("No subfolder children found for '%s'", path)
	return false
}

// Check if a folder might have notes - this will be used in GetFolderChildrenHandler
func hasNotes(path string, store types.Storage) bool {
	notes, err := store.GetPagesInFolder(path)
	if err != nil {
		log.Printf("Error checking for notes in folder '%s': %v", path, err)
		return false
	}
	hasNotes := len(notes) > 0
	if hasNotes {
		log.Printf("Found %d notes in folder '%s'", len(notes), path)
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
			children = append(children, FolderTreeItem{
				Name:        getNameFromPath(folder),
				Path:        folder,
				HasChildren: hasChildren(folders, folder),
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
				IsExpanded:  false, // Always collapsed by default, even for current folder
				Children:    []FolderTreeItem{},
				IsNote:      false, // This is a folder
			}

			rootItems = append(rootItems, item)
		}
	}

	// Sort root items alphabetically
	sort.Slice(rootItems, func(i, j int) bool {
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
