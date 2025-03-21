package storage

import (
	"context"
	"fmt"
	"log"
	"path/filepath"
	"strings"
	"time"

	"github.com/daniel-vuky/golang-my-wiki-v2/pkg/config"
	"github.com/daniel-vuky/golang-my-wiki-v2/pkg/storage/types"
	"github.com/google/go-github/v45/github"
	"golang.org/x/oauth2"
)

// GitHubStorage implements the Storage interface using GitHub as the backend
type GitHubStorage struct {
	client     *github.Client
	owner      string
	repository string
	branch     string
	ctx        context.Context
}

// NewGitHubStorage creates a new GitHub storage instance
func NewGitHubStorage(config *config.Config) (*GitHubStorage, error) {
	log.Printf("Initializing GitHub storage with owner: %s, repo: %s, branch: %s",
		config.GitHub.Owner, config.GitHub.Repository, config.GitHub.Branch)

	if config.GitHub.Token == "" {
		return nil, fmt.Errorf("GitHub token is required")
	}
	if config.GitHub.Owner == "" {
		return nil, fmt.Errorf("GitHub owner is required")
	}
	if config.GitHub.Repository == "" {
		return nil, fmt.Errorf("GitHub repository is required")
	}
	if config.GitHub.Branch == "" {
		return nil, fmt.Errorf("GitHub branch is required")
	}

	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: config.GitHub.Token},
	)
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)

	// Verify token permissions
	user, _, err := client.Users.Get(ctx, "")
	if err != nil {
		log.Printf("Error verifying token: %v", err)
		return nil, fmt.Errorf("failed to verify GitHub token: %v", err)
	}
	log.Printf("Authenticated as GitHub user: %s", user.GetLogin())

	// Verify repository access
	repo, resp, err := client.Repositories.Get(ctx, config.GitHub.Owner, config.GitHub.Repository)
	if err != nil {
		log.Printf("Error accessing repository: %v", err)
		if resp != nil {
			log.Printf("GitHub API response status: %s", resp.Status)
			log.Printf("GitHub API response body: %s", resp.Body)
		}
		return nil, fmt.Errorf("failed to access repository: %v", err)
	}
	log.Printf("Successfully accessed repository: %s", repo.GetFullName())

	// Verify branch exists
	branch, resp, err := client.Repositories.GetBranch(ctx, config.GitHub.Owner, config.GitHub.Repository, config.GitHub.Branch, false)
	if err != nil {
		log.Printf("Error accessing branch: %v", err)
		if resp != nil {
			log.Printf("GitHub API response status: %s", resp.Status)
			log.Printf("GitHub API response body: %s", resp.Body)
		}
		return nil, fmt.Errorf("failed to access branch: %v", err)
	}
	log.Printf("Successfully accessed branch: %s", branch.GetName())

	log.Printf("Successfully initialized GitHub storage")
	return &GitHubStorage{
		client:     client,
		owner:      config.GitHub.Owner,
		repository: config.GitHub.Repository,
		branch:     config.GitHub.Branch,
		ctx:        ctx,
	}, nil
}

// ListPages retrieves all pages from the GitHub repository
func (g *GitHubStorage) ListPages() ([]types.Page, error) {
	ctx := context.Background()
	opts := &github.RepositoryContentGetOptions{
		Ref: g.branch,
	}

	_, contents, _, err := g.client.Repositories.GetContents(ctx, g.owner, g.repository, "", opts)
	if err != nil {
		return nil, fmt.Errorf("failed to get repository contents: %v", err)
	}

	var pages []types.Page
	for _, content := range contents {
		if content.GetType() == "file" && strings.HasSuffix(content.GetName(), ".txt") {
			page, err := g.GetPage(content.GetPath())
			if err != nil {
				continue // Skip files that can't be read
			}
			pages = append(pages, *page)
		}
	}

	return pages, nil
}

// GetPage retrieves a page from GitHub
func (g *GitHubStorage) GetPage(path string) (*types.Page, error) {
	log.Printf("=== GetPage START: %s ===", path)

	// If path doesn't end with .txt, add it
	if !strings.HasSuffix(path, ".txt") {
		// Check if this is a path with directories
		if strings.Contains(path, "/") {
			// Split the path to get the filename
			parts := strings.Split(path, "/")
			filename := parts[len(parts)-1]

			// If the filename doesn't have .txt, add it
			if !strings.HasSuffix(filename, ".txt") {
				parts[len(parts)-1] = filename + ".txt"
				path = strings.Join(parts, "/")
			}
		} else {
			// Simple filename without directories
			path = path + ".txt"
		}
		log.Printf("Fixed path to include .txt extension: %s", path)
	}

	// Get the file content
	fileContent, _, _, err := g.client.Repositories.GetContents(
		g.ctx,
		g.owner,
		g.repository,
		path,
		&github.RepositoryContentGetOptions{Ref: g.branch},
	)
	if err != nil {
		log.Printf("Error getting content: %v", err)
		return nil, fmt.Errorf("failed to get content: %v", err)
	}

	// Get the file name without extension
	fileName := strings.TrimSuffix(filepath.Base(path), ".txt")
	log.Printf("File name without extension: %s", fileName)

	// Get the content string directly
	contentStr, err := fileContent.GetContent()
	if err != nil {
		log.Printf("Error getting content string: %v", err)
		return nil, fmt.Errorf("failed to get content string: %v", err)
	}
	log.Printf("Content length: %d", len(contentStr))
	log.Printf("Content preview: %s", contentStr[:min(100, len(contentStr))])

	// Create the page object
	page := &types.Page{
		Title:   fileName,
		Path:    path,
		Content: contentStr,
		Body:    []byte(contentStr),
	}

	log.Printf("=== GetPage END: %s ===", path)
	return page, nil
}

// Helper function to get minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// CreatePage creates a new page in GitHub
func (g *GitHubStorage) CreatePage(page *types.Page) error {
	log.Printf("=== CreatePage START: %s ===", page.Path)
	log.Printf("Repository: %s/%s, Branch: %s", g.owner, g.repository, g.branch)

	if page == nil {
		return fmt.Errorf("page cannot be nil")
	}
	if page.Path == "" {
		return fmt.Errorf("page path cannot be empty")
	}
	if len(page.Body) == 0 {
		return fmt.Errorf("page content cannot be empty")
	}

	// Ensure path ends with .txt
	if !strings.HasSuffix(page.Path, ".txt") {
		page.Path = page.Path + ".txt"
		log.Printf("Fixed path to include .txt extension: %s", page.Path)
	}

	opts := &github.RepositoryContentFileOptions{
		Message: github.String(fmt.Sprintf("Create page: %s", page.Path)),
		Content: page.Body,
		Branch:  &g.branch,
	}

	log.Printf("Sending create request to GitHub API with content length: %d", len(page.Body))
	log.Printf("Repository: %s/%s, Branch: %s", g.owner, g.repository, g.branch)

	_, resp, err := g.client.Repositories.CreateFile(g.ctx, g.owner, g.repository, page.Path, opts)
	if err != nil {
		log.Printf("Error creating page: %v", err)
		if resp != nil {
			log.Printf("GitHub API response status: %s", resp.Status)
			log.Printf("GitHub API response body: %s", resp.Body)
		}
		// Check if it's an authentication error
		if strings.Contains(err.Error(), "401") {
			return fmt.Errorf("GitHub authentication failed. Please check your token permissions")
		}
		// Check if it's a permission error
		if strings.Contains(err.Error(), "403") {
			return fmt.Errorf("GitHub permission denied. Please check your token permissions")
		}
		// Check if it's a 404 error (repository not found)
		if strings.Contains(err.Error(), "404") {
			return fmt.Errorf("Repository not found. Please check repository name and owner")
		}
		return fmt.Errorf("failed to create page: %v", err)
	}

	if resp != nil {
		log.Printf("GitHub API response status: %s", resp.Status)
		log.Printf("GitHub API response body: %s", resp.Body)
	}
	log.Printf("=== CreatePage END: %s ===", page.Path)
	log.Printf("Successfully created page: %s", page.Path)
	return nil
}

// UpdatePage updates an existing page in GitHub
func (g *GitHubStorage) UpdatePage(page *types.Page) error {
	log.Printf("=== UpdatePage START: %s ===", page.Path)

	// Make sure we have a valid path ending with .txt
	if !strings.HasSuffix(page.Path, ".txt") {
		page.Path = page.Path + ".txt"
	}
	log.Printf("Using path: %s", page.Path)

	// First try to check if the file exists and get its SHA
	fileContent, _, resp, err := g.client.Repositories.GetContents(
		g.ctx,
		g.owner,
		g.repository,
		page.Path,
		&github.RepositoryContentGetOptions{Ref: g.branch},
	)

	if err != nil {
		log.Printf("Error checking if file exists: %v", err)
		if resp != nil && resp.StatusCode == 404 {
			// File doesn't exist, create it instead
			log.Printf("File doesn't exist, creating instead of updating")
			return g.CreatePage(page)
		}
		return fmt.Errorf("failed to check if file exists: %v", err)
	}

	if fileContent == nil {
		log.Printf("File content is nil, creating instead of updating")
		return g.CreatePage(page)
	}

	// File exists, update it with the SHA
	opts := &github.RepositoryContentFileOptions{
		Message: github.String(fmt.Sprintf("Update page: %s", page.Path)),
		Content: page.Body,
		SHA:     fileContent.SHA,
		Branch:  github.String(g.branch),
	}

	_, resp, err = g.client.Repositories.UpdateFile(
		g.ctx,
		g.owner,
		g.repository,
		page.Path,
		opts,
	)

	if err != nil {
		log.Printf("Error updating file: %v", err)
		if resp != nil {
			log.Printf("Response status: %s", resp.Status)
		}
		return fmt.Errorf("failed to update file: %v", err)
	}

	log.Printf("Successfully updated file: %s", page.Path)
	log.Printf("=== UpdatePage END ===")
	return nil
}

// DeletePage deletes a page from GitHub
func (g *GitHubStorage) DeletePage(path string) error {
	log.Printf("=== DeletePage START: %s ===", path)

	// If path doesn't end with .txt, add it
	if !strings.HasSuffix(path, ".txt") {
		// Check if this is a path with directories
		if strings.Contains(path, "/") {
			// Split the path to get the filename
			parts := strings.Split(path, "/")
			filename := parts[len(parts)-1]

			// If the filename doesn't have .txt, add it
			if !strings.HasSuffix(filename, ".txt") {
				parts[len(parts)-1] = filename + ".txt"
				path = strings.Join(parts, "/")
			}
		} else {
			// Simple filename without directories
			path = path + ".txt"
		}
		log.Printf("Fixed path to include .txt extension: %s", path)
	}

	// Get the file content first to get its SHA
	fileContent, _, resp, err := g.client.Repositories.GetContents(
		g.ctx,
		g.owner,
		g.repository,
		path,
		&github.RepositoryContentGetOptions{Ref: g.branch},
	)

	if err != nil {
		log.Printf("Error getting file: %v", err)
		if resp != nil && resp.StatusCode == 404 {
			log.Printf("File doesn't exist (404), nothing to delete")
			return fmt.Errorf("file not found: %s", path)
		}
		return fmt.Errorf("failed to get file: %v", err)
	}

	if fileContent == nil {
		log.Printf("File content is nil, nothing to delete")
		return fmt.Errorf("file not found: %s", path)
	}

	// File exists, delete it
	opts := &github.RepositoryContentFileOptions{
		Message: github.String(fmt.Sprintf("Delete page: %s", path)),
		SHA:     fileContent.SHA,
		Branch:  github.String(g.branch),
	}

	_, resp, err = g.client.Repositories.DeleteFile(
		g.ctx,
		g.owner,
		g.repository,
		path,
		opts,
	)

	if err != nil {
		log.Printf("Error deleting file: %v", err)
		if resp != nil {
			log.Printf("Response status: %s", resp.Status)
		}
		return fmt.Errorf("failed to delete file: %v", err)
	}

	log.Printf("Successfully deleted file: %s", path)
	log.Printf("=== DeletePage END ===")
	return nil
}

// ListFolders retrieves all folders from the GitHub repository
func (g *GitHubStorage) ListFolders() ([]string, error) {
	log.Printf("=== ListFolders START ===")

	// Get root folders first
	var allFolders []string

	// Call recursive helper to get all folders
	if err := g.getAllFolders("", &allFolders); err != nil {
		return nil, err
	}

	log.Printf("Found total of %d folders", len(allFolders))
	for _, folder := range allFolders {
		log.Printf("Folder: %s", folder)
	}

	log.Printf("=== ListFolders END ===")
	return allFolders, nil
}

// getAllFolders recursively gets all folders
func (g *GitHubStorage) getAllFolders(path string, allFolders *[]string) error {
	opts := &github.RepositoryContentGetOptions{
		Ref: g.branch,
	}

	_, contents, _, err := g.client.Repositories.GetContents(g.ctx, g.owner, g.repository, path, opts)
	if err != nil {
		return fmt.Errorf("failed to get repository contents: %v", err)
	}

	for _, content := range contents {
		if content.GetType() == "dir" {
			folderPath := content.GetPath()
			log.Printf("Found folder: %s", folderPath)
			*allFolders = append(*allFolders, folderPath)

			// Recursively get subfolders
			if err := g.getAllFolders(folderPath, allFolders); err != nil {
				log.Printf("Warning: Error getting subfolders of %s: %v", folderPath, err)
				// Continue even if there's an error with one subfolder
			}
		}
	}

	return nil
}

// CreateFolder creates a new folder in GitHub
func (g *GitHubStorage) CreateFolder(path string) error {
	log.Printf("=== CreateFolder START: %s ===", path)
	ctx := context.Background()

	// Create a .folder file to mark the directory
	folderPath := path + "/.folder"

	// GitHub requires content to be non-empty and properly encoded
	// This is a small text file with a message explaining its purpose
	content := []byte("This file marks the folder for the wiki system. Please do not delete.")

	opts := &github.RepositoryContentFileOptions{
		Message: github.String(fmt.Sprintf("Create folder: %s", path)),
		Content: content,
		Branch:  &g.branch,
	}

	log.Printf("Creating folder marker file at: %s", folderPath)
	_, resp, err := g.client.Repositories.CreateFile(ctx, g.owner, g.repository, folderPath, opts)
	if err != nil {
		log.Printf("Error creating folder: %v", err)
		if resp != nil {
			log.Printf("GitHub API response status: %s", resp.Status)
		}
		return fmt.Errorf("failed to create folder: %v", err)
	}

	log.Printf("Successfully created folder: %s", path)
	log.Printf("=== CreateFolder END ===")
	return nil
}

// DeleteFolder deletes a folder from GitHub
func (g *GitHubStorage) DeleteFolder(path string) error {
	ctx := context.Background()

	// Get all contents in the folder
	_, contents, _, err := g.client.Repositories.GetContents(ctx, g.owner, g.repository, path, nil)
	if err != nil {
		return fmt.Errorf("failed to get folder contents: %v", err)
	}

	// Delete all files in the folder
	for _, content := range contents {
		if err := g.DeletePage(content.GetPath()); err != nil {
			return fmt.Errorf("failed to delete file %s: %v", content.GetPath(), err)
		}
	}

	return nil
}

// GetPagesInFolder retrieves pages from a specific folder
func (g *GitHubStorage) GetPagesInFolder(folderPath string) ([]types.Page, error) {
	log.Printf("=== GetPagesInFolder START: %s ===", folderPath)

	// Handle empty or root path
	if folderPath == "" {
		log.Printf("Empty folder path, returning empty result")
		return []types.Page{}, nil
	}

	// Set up context and options
	ctx, cancel := context.WithTimeout(g.ctx, time.Second*30)
	defer cancel()

	opts := &github.RepositoryContentGetOptions{
		Ref: g.branch,
	}

	// Get contents of the folder
	_, dirContents, resp, err := g.client.Repositories.GetContents(
		ctx,
		g.owner,
		g.repository,
		folderPath,
		opts,
	)

	if err != nil {
		log.Printf("Error getting contents of folder: %v", err)
		if resp != nil && resp.StatusCode == 404 {
			log.Printf("Folder not found (404): %s", folderPath)
			return []types.Page{}, nil
		}
		return nil, fmt.Errorf("failed to get folder contents: %v", err)
	}

	log.Printf("Found %d items in folder: %s", len(dirContents), folderPath)

	// Filter for .txt files and create Page objects
	var pages []types.Page
	for _, content := range dirContents {
		if content == nil || content.Type == nil || content.Name == nil {
			log.Printf("Skipping nil content item")
			continue
		}

		if *content.Type == "file" && strings.HasSuffix(*content.Name, ".txt") {
			log.Printf("Processing file: %s", *content.Name)

			// Get the file content
			fileContent, _, _, err := g.client.Repositories.GetContents(
				ctx,
				g.owner,
				g.repository,
				*content.Path,
				opts,
			)

			if err != nil {
				log.Printf("Error getting file content: %v, skipping", err)
				continue
			}

			// Decode content
			decodedContent, err := fileContent.GetContent()
			if err != nil {
				log.Printf("Error decoding content: %v, skipping", err)
				continue
			}

			title := strings.TrimSuffix(*content.Name, ".txt")

			// Add page to results
			page := types.Page{
				Title:   title,
				Path:    *content.Path,
				Content: decodedContent,
				Body:    []byte(decodedContent),
			}

			// Add preview
			if len(decodedContent) > 150 {
				page.Preview = decodedContent[:150] + "..."
			} else {
				page.Preview = decodedContent
			}

			pages = append(pages, page)
			log.Printf("Added page: %s", title)
		}
	}

	log.Printf("=== GetPagesInFolder END: %s, found %d pages ===", folderPath, len(pages))
	return pages, nil
}
