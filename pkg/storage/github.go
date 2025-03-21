package storage

import (
	"context"
	"fmt"
	"log"
	"path/filepath"
	"strings"

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

// GetPage retrieves a specific page from GitHub
func (g *GitHubStorage) GetPage(path string) (*types.Page, error) {
	log.Printf("=== GetPage START: %s ===", path)
	log.Printf("Repository: %s, Branch: %s", g.repository, g.branch)

	// Ensure path ends with .txt
	if !strings.HasSuffix(path, ".txt") {
		path = path + ".txt"
	}
	log.Printf("Final path: %s", path)

	// Get the file content
	content, _, _, err := g.client.Repositories.GetContents(g.ctx, g.owner, g.repository, path, &github.RepositoryContentGetOptions{
		Ref: g.branch,
	})
	if err != nil {
		log.Printf("Error getting content: %v", err)
		return nil, fmt.Errorf("failed to get content: %v", err)
	}

	// Get the file name without extension
	fileName := strings.TrimSuffix(filepath.Base(path), ".txt")
	log.Printf("File name without extension: %s", fileName)

	// Get the content string directly
	contentStr, err := content.GetContent()
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

	// Make sure we have a valid path ending with .txt
	if !strings.HasSuffix(path, ".txt") {
		path = path + ".txt"
	}
	log.Printf("Using path: %s", path)

	// Get the current file to get its SHA
	content, _, resp, err := g.client.Repositories.GetContents(
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

	if content == nil {
		log.Printf("File content is nil, nothing to delete")
		return fmt.Errorf("file not found: %s", path)
	}

	// File exists, delete it
	opts := &github.RepositoryContentFileOptions{
		Message: github.String(fmt.Sprintf("Delete page: %s", path)),
		SHA:     content.SHA,
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
	ctx := context.Background()
	opts := &github.RepositoryContentGetOptions{
		Ref: g.branch,
	}

	_, contents, _, err := g.client.Repositories.GetContents(ctx, g.owner, g.repository, "", opts)
	if err != nil {
		return nil, fmt.Errorf("failed to get repository contents: %v", err)
	}

	var folders []string
	for _, content := range contents {
		if content.GetType() == "dir" {
			folders = append(folders, content.GetPath())
		}
	}

	return folders, nil
}

// CreateFolder creates a new folder in GitHub
func (g *GitHubStorage) CreateFolder(path string) error {
	ctx := context.Background()

	// Create an empty .folder file to mark the directory
	folderPath := path + "/.folder"
	opts := &github.RepositoryContentFileOptions{
		Message: github.String(fmt.Sprintf("Create folder: %s", path)),
		Content: []byte(""), // Empty file to mark directory
		Branch:  &g.branch,
	}

	_, _, err := g.client.Repositories.CreateFile(ctx, g.owner, g.repository, folderPath, opts)
	if err != nil {
		return fmt.Errorf("failed to create folder: %v", err)
	}

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
