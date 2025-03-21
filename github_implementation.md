# GitHub Integration Implementation Plan

## Overview
This document outlines the steps required to integrate GitHub as a storage backend for the wiki system, allowing seamless switching between local file storage and GitHub storage.

## Prerequisites
1. GitHub Personal Access Token with the following permissions:
   - `repo` (Full control of private repositories)
   - `workflow` (Optional, for GitHub Actions if needed)

2. Required Go packages:
   - `github.com/google/go-github/v45` (GitHub API client)
   - `golang.org/x/oauth2` (OAuth2 authentication)

## Implementation Steps

### 1. Configuration Setup
1. Add new configuration options in `config/config.go`:
   ```go
   type Config struct {
       // ... existing fields ...
       StorageMode string // "local" or "github"
       GitHubConfig struct {
           Token      string
           Owner      string
           Repository string
           Branch     string
       }
   }
   ```

2. Update configuration file structure to include GitHub settings:
   ```yaml
   storage_mode: "github"  # or "local"
   github:
     token: "your-github-token"
     owner: "your-username"
     repository: "your-wiki-repo"
     branch: "main"
   ```

### 2. Create Storage Interface
1. Create a new interface in `storage/storage.go`:
   ```go
   type Storage interface {
       ListPages() ([]Page, error)
       GetPage(path string) (*Page, error)
       CreatePage(page *Page) error
       UpdatePage(page *Page) error
       DeletePage(path string) error
       ListFolders() ([]string, error)
       CreateFolder(path string) error
       DeleteFolder(path string) error
   }
   ```

2. Implement the interface for both local and GitHub storage:
   - `storage/local.go` (existing implementation)
   - `storage/github.go` (new implementation)

### 3. GitHub Storage Implementation
1. Create `storage/github.go` with the following components:
   - GitHub client setup with authentication
   - CRUD operations implementation
   - Folder structure handling
   - Content encoding/decoding

2. Key functions to implement:
   ```go
   type GitHubStorage struct {
       client     *github.Client
       owner      string
       repository string
       branch     string
   }

   // Required methods:
   - NewGitHubStorage(config *config.Config) (*GitHubStorage, error)
   - ListPages() ([]Page, error)
   - GetPage(path string) (*Page, error)
   - CreatePage(page *Page) error
   - UpdatePage(page *Page) error
   - DeletePage(path string) error
   - ListFolders() ([]string, error)
   - CreateFolder(path string) error
   - DeleteFolder(path string) error
   ```

### 4. Storage Factory
1. Create a storage factory to handle storage mode switching:
   ```go
   func NewStorage(config *config.Config) (Storage, error) {
       switch config.StorageMode {
       case "github":
           return NewGitHubStorage(config)
       case "local":
           return NewLocalStorage(config)
       default:
           return nil, fmt.Errorf("unsupported storage mode: %s", config.StorageMode)
       }
   }
   ```

### 5. Update Handlers
1. Modify existing handlers to use the storage interface instead of direct file operations
2. Update all CRUD operations to work with the storage abstraction
3. Ensure proper error handling for both storage modes

### 6. Folder Structure Implementation
1. Implement folder creation/deletion in GitHub:
   - Use empty files with `.folder` extension to mark directories
   - Maintain folder hierarchy in GitHub repository
   - Handle nested folder structures

2. Update folder listing to work with GitHub structure:
   - Parse repository contents to identify folders
   - Handle nested folder navigation
   - Maintain folder metadata

### 7. Testing
1. Create test cases for GitHub storage implementation
2. Test both storage modes
3. Test folder operations
4. Test concurrent access handling

### 8. Documentation
1. Update README with GitHub integration instructions
2. Document configuration options
3. Add setup guide for GitHub token
4. Include examples for both storage modes

## Security Considerations
1. Secure storage of GitHub token
2. Rate limiting handling
3. Error handling for API limits
4. Proper access control

## Migration Guide
1. Create migration script from local to GitHub storage
2. Document backup procedures
3. Provide rollback instructions

## Future Enhancements
1. GitHub webhook support for real-time updates
2. Collaborative editing support
3. Version history integration
4. Branch-based content management

## Implementation Timeline
1. Configuration Setup: 1 day
2. Storage Interface: 1 day
3. GitHub Implementation: 2-3 days
4. Handler Updates: 1-2 days
5. Testing: 2 days
6. Documentation: 1 day
7. Migration Tools: 1 day

Total estimated time: 9-11 days

## Notes
- Ensure proper error handling for network issues
- Implement caching where appropriate
- Consider implementing retry mechanisms for failed API calls
- Add logging for GitHub operations
- Consider implementing a queue system for multiple updates 