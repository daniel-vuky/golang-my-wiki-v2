# Daniel's Wiki

A personal wiki system built with Go, featuring GitHub storage integration and a modern web interface.

## Features

- 📝 Create and edit notes with rich text editing
- 📁 Hierarchical folder structure for organizing content
- 🌓 Dark/Light theme support
- 🔒 GitHub-based storage
- 📱 Responsive design
- 🔄 Real-time folder tree navigation
- 🎨 Modern UI with Font Awesome icons
- 🔐 Google OAuth authentication

## Technical Stack

- **Backend**: Go 1.21+
- **Frontend**:
  - HTML5
  - CSS3 (with CSS Variables for theming)
  - JavaScript (Vanilla)
  - TinyMCE for rich text editing
  - Font Awesome 6.0.0 for icons
- **Storage**: GitHub API
- **Caching**: Redis
- **Template Engine**: Go's html/template
- **Authentication**: Google OAuth2

## System Requirements

- Go 1.21 or higher
- Redis server
- GitHub account with appropriate permissions
- Google Cloud Platform account
- Modern web browser with JavaScript enabled

## Prerequisites

1. Install Go 1.21 or higher
2. Install and start Redis server
3. Create a GitHub Personal Access Token with `repo` scope
4. Set up Google OAuth credentials:
   - Go to [Google Cloud Console](https://console.cloud.google.com/)
   - Create a new project or select an existing one
   - Enable the Google+ API
   - Go to "Credentials" and create an OAuth 2.0 Client ID
   - Set the authorized redirect URI to: `{{url}}:8080/auth/google/callback`
   - Download the client credentials JSON file
5. Set up environment variables (see Configuration section)

## Installation

1. Clone the repository:
```bash
git clone https://github.com/daniel-vuky/golang-my-wiki-v2.git
cd golang-my-wiki-v2
```

2. Install dependencies:
```bash
go mod download
```

3. Create a `.env` file in the project root with the following variables:
```env
GITHUB_TOKEN=your_github_token
GITHUB_OWNER=your_github_username
GITHUB_REPO=wiki_storage
GITHUB_BRANCH=main
REDIS_URL=localhost:6379
GOOGLE_CLIENT_ID=your_google_client_id
GOOGLE_CLIENT_SECRET=your_google_client_secret
GOOGLE_REDIRECT_URL={{url}}:8080/auth/google/callback
```

4. Build the application:
```bash
go build -o wiki cmd/server/main.go
```

5. Run the server:
```bash
./wiki
```

The application will be available at `http://localhost:8080`

## Project Structure

```
.
├── cmd/
│   └── server/
│       └── main.go
├── internal/
│   ├── config/
│   │   └── config.go
│   ├── handlers/
│   │   └── handlers.go
│   ├── middleware/
│   │   └── auth.go
│   ├── models/
│   │   └── models.go
│   ├── storage/
│   │   └── github.go
│   └── utils/
│       └── utils.go
├── pkg/
│   └── handlers/
│       └── handlers.go
├── static/
│   ├── css/
│   │   ├── base.css
│   │   ├── components/
│   │   └── pages/
│   └── js/
│       ├── sidebar.js
│       ├── delete.js
│       └── theme.js
├── templates/
│   ├── base.html
│   ├── edit.html
│   ├── folder.html
│   ├── folder_sidebar.html
│   ├── login.html
│   ├── new.html
│   └── view.html
├── .env
├── .gitignore
├── go.mod
├── go.sum
└── README.md
```

## Configuration

### Environment Variables

- `GITHUB_TOKEN`: Your GitHub Personal Access Token
- `GITHUB_OWNER`: Your GitHub username
- `GITHUB_REPO`: Repository name for storing wiki content
- `GITHUB_BRANCH`: Branch to use for wiki content
- `REDIS_URL`: Redis server URL
- `GOOGLE_CLIENT_ID`: Your Google OAuth client ID
- `GOOGLE_CLIENT_SECRET`: Your Google OAuth client secret
- `GOOGLE_REDIRECT_URL`: OAuth callback URL (default: {{url}}:8080/auth/google/callback)

### Server Configuration

The server can be configured with the following command-line flags:

- `-port`: Port to run the server on (default: 8080)
- `-max-category-level`: Maximum depth for category nesting (default: 4)

## Usage

### Authentication

1. Click the "Login" button in the top-right corner
2. Select "Login with Google"
3. Choose your Google account
4. Grant the requested permissions
5. You'll be redirected back to the wiki

### Creating a New Note

1. Click the "New Page" button in the sidebar
2. Enter a title for your note
3. Use the rich text editor to write your content
4. Click "Create Note" to save

### Editing a Note

1. Navigate to the note you want to edit
2. Click the "Edit" button
3. Make your changes
4. Click "Save Changes"

### Organizing Notes

- Create folders by navigating to a category and clicking "New Folder"
- Move notes between folders using the folder selector when editing
- Use the sidebar tree to navigate between folders and notes

### Theme

- Click the theme toggle button to switch between light and dark modes
- Theme preference is saved in localStorage

## Development

### Adding New Features

1. Create a new branch for your feature
2. Make your changes
3. Test thoroughly
4. Submit a pull request

### Running Tests

```bash
go test ./...
```

### Code Style

- Follow Go standard formatting (`go fmt`)
- Use meaningful variable and function names
- Add comments for complex logic
- Keep functions focused and small

## Contributing

1. Fork the repository
2. Create a feature branch
3. Commit your changes
4. Push to the branch
5. Create a Pull Request

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Acknowledgments

- [TinyMCE](https://www.tiny.cloud/) for the rich text editor
- [Font Awesome](https://fontawesome.com/) for the icons
- [Go](https://golang.org/) for the programming language
- [GitHub](https://github.com/) for the storage solution
- [Google](https://developers.google.com/identity/protocols/oauth2) for OAuth authentication
