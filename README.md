# Daniel's Wiki

[![Go](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat-square&logo=go)](https://golang.org/)
[![License](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

A modern, feature-rich personal wiki system built with Go and a beautiful user interface. This wiki system supports both local file storage and GitHub-based storage, with Redis caching for improved performance.

## 🛠️ Tech Stack

- **Backend**: Go 1.22+
- **Web Framework**: Gin
- **Authentication**: Google OAuth2
- **Storage Options**:
  - Local File System
  - GitHub Repository
- **Caching**: Redis
- **Frontend**:
  - HTML5
  - CSS3 (with CSS Variables for theming)
  - JavaScript (Vanilla)
- **Markdown Support**: Blackfriday v2
- **Configuration**: Viper

## 🌟 Features

- **Rich Text Editor**: TinyMCE integration for seamless content editing
- **Dark/Light Theme**: Automatic theme switching with CSS variables
- **Responsive Design**: Mobile-friendly interface
- **User Authentication**: 
  - Google OAuth2 integration
  - Email-based access control
  - Session management
- **Storage Options**:
  - Local file system storage
  - GitHub repository storage
  - Redis caching for improved performance
- **Category Management**:
  - Hierarchical folder structure
  - Dynamic folder tree navigation
  - Support for multiple category levels
- **Clean URLs**: SEO-friendly structure
- **Performance Optimizations**:
  - Redis caching layer
  - Efficient file system operations
  - Optimized GitHub API usage

## 🚀 Quick Start

### Prerequisites

1. Go 1.22 or higher
2. Redis (optional, for caching)
3. Google OAuth2 credentials (for authentication)
4. GitHub Personal Access Token (if using GitHub storage)

### Installation

1. Clone the repository:
   ```bash
   git clone https://github.com/daniel-vuky/golang-my-wiki-v2.git
   cd golang-my-wiki-v2
   ```

2. Install dependencies:
   ```bash
   go mod download
   ```

3. Create configuration file:
   ```bash
   cp env.yaml.example env.yaml
   ```

4. Update `env.yaml` with your settings:
   ```yaml
   google:
     client_id: your_client_id
     client_secret: your_client_secret
     redirect_url: {{url}}:8080/auth/google/callback
     allowed_emails:
       - your.email@example.com

   session:
     secret: your_session_secret
     name: wiki_session
     secure: false
     allowed_emails:
       - your.email@example.com

   server:
     port: 8080
     host: localhost
     data_dir: ./data

   storage_mode: github  # or "local"

   # If using GitHub storage
   github:
     token: your_github_token
     owner: your_username
     repository: your_wiki_repo
     branch: main

   # Redis cache settings (optional)
   redis:
     address: localhost:6379
     password: ""
     db: 0
     enabled: true
     expiration_seconds: 900
   ```

5. Create data directory:
   ```bash
   mkdir data
   ```

6. Run the application:
   ```bash
   go run cmd/server/main.go
   ```

7. Visit `http://localhost:8080` in your browser

## 📁 Project Structure

```
golang-my-wiki-v2/
├── cmd/
│   └── server/
│       └── main.go          # Application entry point
├── pkg/
│   ├── auth/               # Authentication package
│   ├── cache/              # Redis caching package
│   ├── config/             # Configuration management
│   ├── handlers/           # HTTP request handlers
│   ├── middleware/         # HTTP middleware
│   ├── models/             # Data models
│   └── storage/            # Storage implementations
├── static/                 # Static assets (CSS, JS)
├── templates/              # HTML templates
├── data/                   # Wiki page storage (local mode)
├── env.yaml               # Configuration file
└── go.mod                 # Go module file
```

## 🔧 Development

- Run tests:
  ```bash
  go test ./...
  ```

- Format code:
  ```bash
  go fmt ./...
  ```

## 🔒 Security

- All user authentication is handled through Google OAuth2
- Session management with secure cookie storage
- Email-based access control
- HTTPS support (configurable in production)

## 📝 License

MIT License - feel free to use and modify!

## 🤝 Contributing

Contributions are welcome! Please feel free to submit a Pull Request.
