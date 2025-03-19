package handlers

import (
	"html/template"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"

	"github.com/daniel-vuky/golang-my-wiki-v2/pkg/models"
	"github.com/gin-gonic/gin"
	"github.com/russross/blackfriday/v2"
)

const dataDir = "data"

// LoadTemplates creates a multitemplate renderer for Gin
func LoadTemplates(router *gin.Engine) {
	// Create a function map for the templates
	funcMap := template.FuncMap{
		"add": func(a, b int) int {
			return a + b
		},
	}

	// Set the HTML renderer with custom delimiters and functions
	router.SetFuncMap(funcMap)
	router.Delims("{{", "}}")
	router.LoadHTMLGlob("templates/*.html")
}

// HomeHandler displays the wiki home page
func HomeHandler(c *gin.Context) {
	pages, err := models.GetAllPages()
	if err != nil {
		c.String(http.StatusInternalServerError, "Error loading pages")
		return
	}

	user := c.MustGet("user")
	sideMenu := models.GetMenuItems("")

	c.HTML(http.StatusOK, "home.html", gin.H{
		"Pages":    pages,
		"User":     user,
		"SideMenu": sideMenu,
	})
}

// ViewHandler displays a wiki page
func ViewHandler(c *gin.Context) {
	title := c.Param("title")
	if title == "" {
		c.Redirect(http.StatusFound, "/")
		return
	}

	filename := filepath.Join(dataDir, title+".txt")
	body, err := ioutil.ReadFile(filename)
	if err != nil {
		if os.IsNotExist(err) {
			c.Redirect(http.StatusFound, "/edit/"+title)
			return
		}
		c.String(http.StatusInternalServerError, "Error reading page")
		return
	}

	// Convert markdown to HTML
	html := blackfriday.Run(body)

	user := c.MustGet("user")
	sideMenu := models.GetMenuItems(title)

	c.HTML(http.StatusOK, "view.html", gin.H{
		"Title":    title,
		"Content":  template.HTML(string(html)),
		"User":     user,
		"SideMenu": sideMenu,
	})
}

// EditHandler displays the edit form for a wiki page
func EditHandler(c *gin.Context) {
	title := c.Param("title")
	isNewPage := title == "NewPage"

	var body []byte
	var err error

	if !isNewPage {
		filename := filepath.Join(dataDir, title+".txt")
		body, err = ioutil.ReadFile(filename)
		if err != nil && !os.IsNotExist(err) {
			c.String(http.StatusInternalServerError, "Error reading page")
			return
		}
	}

	user := c.MustGet("user")
	sideMenu := models.GetMenuItems(title)

	c.HTML(http.StatusOK, "edit.html", gin.H{
		"Title":     title,
		"Body":      string(body),
		"User":      user,
		"SideMenu":  sideMenu,
		"IsNewPage": isNewPage,
	})
}

// SaveHandler saves a wiki page
func SaveHandler(c *gin.Context) {
	var oldTitle string
	var newTitle string

	if c.Param("title") == "new" {
		// For new pages, get title from form
		newTitle = c.PostForm("title")
		if newTitle == "" {
			c.String(http.StatusBadRequest, "Title is required")
			return
		}
	} else {
		// For existing pages
		oldTitle = c.Param("title")
		newTitle = c.PostForm("title")

		if newTitle == "" {
			c.String(http.StatusBadRequest, "Title is required")
			return
		}

		// If title hasn't changed, just update the content
		if oldTitle == newTitle {
			body := c.PostForm("body")
			filename := filepath.Join(dataDir, oldTitle+".txt")
			err := ioutil.WriteFile(filename, []byte(body), 0644)
			if err != nil {
				c.String(http.StatusInternalServerError, "Error saving page")
				return
			}
			c.Redirect(http.StatusFound, "/view/"+oldTitle)
			return
		}
	}

	// Check if the new title already exists (except for the current file)
	if oldTitle != newTitle {
		if _, err := os.Stat(filepath.Join(dataDir, newTitle+".txt")); err == nil {
			c.String(http.StatusBadRequest, "A page with this title already exists")
			return
		}
	}

	// Save the content with the new title
	body := c.PostForm("body")
	newFilename := filepath.Join(dataDir, newTitle+".txt")
	err := ioutil.WriteFile(newFilename, []byte(body), 0644)
	if err != nil {
		c.String(http.StatusInternalServerError, "Error saving page")
		return
	}

	// If this was a rename, delete the old file
	if oldTitle != "" && oldTitle != newTitle {
		oldFilename := filepath.Join(dataDir, oldTitle+".txt")
		err = os.Remove(oldFilename)
		if err != nil && !os.IsNotExist(err) {
			// Log the error but don't fail the request
			// The new file was saved successfully
			gin.DefaultErrorWriter.Write([]byte("Error deleting old file: " + err.Error() + "\n"))
		}
	}

	c.Redirect(http.StatusFound, "/view/"+newTitle)
}

// DeleteHandler deletes a wiki page
func DeleteHandler(c *gin.Context) {
	title := c.Param("title")
	if title == "" {
		c.String(http.StatusBadRequest, "No title provided")
		return
	}

	filename := filepath.Join(dataDir, title+".txt")
	err := os.Remove(filename)
	if err != nil {
		if os.IsNotExist(err) {
			c.String(http.StatusNotFound, "Page not found")
			return
		}
		c.String(http.StatusInternalServerError, "Error deleting page")
		return
	}

	c.Redirect(http.StatusFound, "/")
}
