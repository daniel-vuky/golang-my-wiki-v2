package handlers

import (
	"fmt"
	"log"
	"net/http"

	"github.com/daniel-vuky/golang-my-wiki-v2/pkg/models"
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

	pages, err := models.GetAllPages(store)
	if err != nil {
		log.Printf("Error getting pages: %v", err)
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{
			"error": fmt.Sprintf("Failed to get pages: %v", err),
		})
		return
	}

	log.Printf("Found %d pages", len(pages))
	for _, page := range pages {
		log.Printf("Page: %s", page.Title)
	}

	user := c.MustGet("user")
	log.Printf("User: %+v", user)

	c.HTML(http.StatusOK, "home.html", gin.H{
		"Pages":    pages,
		"User":     user,
		"SideMenu": models.GetMenuItems("", store),
	})

	log.Println("=== HomeHandler END ===")
}

// ViewHandler handles viewing a page
func ViewHandler(c *gin.Context) {
	title := c.Param("title")
	log.Printf("=== ViewHandler START: %s ===", title)

	page, err := store.GetPage(title)
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

	c.HTML(http.StatusOK, "view.html", gin.H{
		"Title":    page.Title,
		"Content":  page.Content,
		"SideMenu": models.GetMenuItems(title, store),
		"User":     c.MustGet("user"),
	})
	log.Printf("=== ViewHandler END: %s ===", title)
}

// EditHandler handles editing a page
func EditHandler(c *gin.Context) {
	title := c.Param("title")
	log.Printf("=== EditHandler START: %s ===", title)

	// Handle new page creation
	if title == "" || title == "new" {
		log.Printf("Creating new page form")
		c.HTML(http.StatusOK, "edit.html", gin.H{
			"Title":     "",
			"Content":   "",
			"IsNewPage": true,
			"SideMenu":  models.GetMenuItems("", store),
			"User":      c.MustGet("user"),
		})
		log.Printf("=== EditHandler END (new page) ===")
		return
	}

	// Editing an existing page - GetPage will handle adding .txt
	log.Printf("Attempting to edit page: %s", title)
	page, err := store.GetPage(title)
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
		"Title":     page.Title,
		"Content":   page.Content,
		"IsNewPage": false,
		"SideMenu":  models.GetMenuItems(page.Title, store),
		"User":      c.MustGet("user"),
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

	if title == "" || content == "" {
		log.Println("Error: Title or content is empty")
		c.HTML(http.StatusBadRequest, "error.html", gin.H{
			"error": "Title and content are required",
		})
		return
	}

	// Create a page object with the new title
	page := &types.Page{
		Title:   title,
		Path:    title + ".txt",
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
