# Implementing Google OAuth Login in Go Wiki

This guide provides step-by-step instructions for implementing Google OAuth authentication in your Go Wiki project.

## Step 1: Set Up Google OAuth Credentials

1. Go to [Google Cloud Console](https://console.cloud.google.com/)
2. Create a new project or select an existing one
3. Navigate to "APIs & Services" > "Credentials"
4. Click "Create Credentials" > "OAuth client ID"
5. Select "Web application" as the application type
6. Add a name for your OAuth client (e.g., "Go Wiki")
7. Add authorized redirect URIs:
   - `http://localhost:8080/auth/google/callback` (for development)
   - Add your production URL when deploying
8. Click "Create"
9. Note down your Client ID and Client Secret

## Step 2: Install Required Dependencies

```bash
# OAuth and session management
go get golang.org/x/oauth2
go get github.com/gin-contrib/sessions

# Markdown rendering
go get github.com/russross/blackfriday/v2
```

The `blackfriday` package is used to convert Markdown content to HTML. This is essential for our wiki because:
- Users write content in Markdown format (which is easier to write than raw HTML)
- Content is stored as Markdown in our files
- When displaying content, we convert it to HTML for proper rendering

## Step 3: Create Environment Variables

Create a `.env` file in your project root:

```env
GOOGLE_CLIENT_ID=your_client_id_here
GOOGLE_CLIENT_SECRET=your_client_secret_here
SESSION_SECRET=your_random_secret_here
```

## Step 4: Create Authentication Models

Create `pkg/auth/google.go`:

```go
package auth

import (
    "encoding/json"
    "fmt"
    "net/http"
    "os"

    "golang.org/x/oauth2"
    "golang.org/x/oauth2/google"
)

type GoogleUserInfo struct {
    ID            string `json:"id"`
    Email         string `json:"email"`
    VerifiedEmail bool   `json:"verified_email"`
    Name          string `json:"name"`
    Picture       string `json:"picture"`
}

var (
    googleOauthConfig *oauth2.Config
)

func InitGoogleOAuth() {
    googleOauthConfig = &oauth2.Config{
        ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
        ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
        RedirectURL:  "http://localhost:8080/auth/google/callback",
        Scopes: []string{
            "https://www.googleapis.com/auth/userinfo.email",
            "https://www.googleapis.com/auth/userinfo.profile",
        },
        Endpoint: google.Endpoint,
    }
}

func GetGoogleOAuthConfig() *oauth2.Config {
    return googleOauthConfig
}

func GetGoogleUserInfo(token *oauth2.Token) (*GoogleUserInfo, error) {
    client := googleOauthConfig.Client(oauth2.NoContext, token)
    resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
    if err != nil {
        return nil, fmt.Errorf("failed getting user info: %s", err.Error())
    }
    defer resp.Body.Close()

    var userInfo GoogleUserInfo
    if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
        return nil, fmt.Errorf("failed decoding user info: %s", err.Error())
    }

    return &userInfo, nil
}
```

## Step 5: Create Session Management

Create `pkg/auth/session.go`:

```go
package auth

import (
    "github.com/gin-contrib/sessions"
    "github.com/gin-gonic/gin"
)

type UserSession struct {
    ID    string
    Email string
    Name  string
}

func SaveUserSession(c *gin.Context, user *UserSession) {
    session := sessions.Default(c)
    session.Set("user_id", user.ID)
    session.Set("user_email", user.Email)
    session.Set("user_name", user.Name)
    session.Save()
}

func GetUserSession(c *gin.Context) *UserSession {
    session := sessions.Default(c)
    userID := session.Get("user_id")
    userEmail := session.Get("user_email")
    userName := session.Get("user_name")

    if userID == nil || userEmail == nil || userName == nil {
        return nil
    }

    return &UserSession{
        ID:    userID.(string),
        Email: userEmail.(string),
        Name:  userName.(string),
    }
}

func ClearUserSession(c *gin.Context) {
    session := sessions.Default(c)
    session.Clear()
    session.Save()
}
```

## Step 6: Create Authentication Middleware

Create `pkg/auth/middleware.go`:

```go
package auth

import (
    "github.com/gin-gonic/gin"
    "net/http"
)

func AuthRequired() gin.HandlerFunc {
    return func(c *gin.Context) {
        user := GetUserSession(c)
        if user == nil {
            c.Redirect(http.StatusTemporaryRedirect, "/login")
            c.Abort()
            return
        }
        c.Set("user", user)
        c.Next()
    }
}
```

## Step 7: Create Authentication Handlers

Create `pkg/handlers/auth_handlers.go`:

```go
package handlers

import (
    "fmt"
    "net/http"
    "os"
    "path"
    "time"

    "github.com/gin-contrib/sessions"
    "github.com/gin-gonic/gin"
    "golang.org/x/oauth2"
    "your-project/pkg/auth"
)

func LoginHandler(c *gin.Context) {
    c.HTML(http.StatusOK, "login.html", gin.H{})
}

func GoogleLoginHandler(c *gin.Context) {
    url := auth.GetGoogleOAuthConfig().AuthCodeURL("state")
    c.Redirect(http.StatusTemporaryRedirect, url)
}

func GoogleCallbackHandler(c *gin.Context) {
    code := c.Query("code")
    token, err := auth.GetGoogleOAuthConfig().Exchange(c, code)
    if err != nil {
        c.String(http.StatusInternalServerError, "Failed to exchange token")
        return
    }

    userInfo, err := auth.GetGoogleUserInfo(token)
    if err != nil {
        c.String(http.StatusInternalServerError, "Failed to get user info")
        return
    }

    user := &auth.UserSession{
        ID:    userInfo.ID,
        Email: userInfo.Email,
        Name:  userInfo.Name,
    }

    auth.SaveUserSession(c, user)
    c.Redirect(http.StatusTemporaryRedirect, "/")
}

func LogoutHandler(c *gin.Context) {
    auth.ClearUserSession(c)
    c.Redirect(http.StatusTemporaryRedirect, "/login")
}
```

## Step 8: Create Login Template

Create `templates/login.html`:

```html
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Login - Go Wiki</title>
    <link rel="stylesheet" href="/static/css/styles.css">
</head>
<body>
    <div class="container">
        <div class="login-box">
            <h1>Welcome to Go Wiki</h1>
            <p>Please sign in to continue</p>
            <a href="/auth/google" class="google-btn">
                <i class="fab fa-google"></i>
                Sign in with Google
            </a>
        </div>
    </div>
</body>
</html>
```

## Step 9: Update Main Application

Update `cmd/server/main.go`:

```go
package main

import (
    "github.com/gin-contrib/sessions"
    "github.com/gin-contrib/sessions/cookie"
    "github.com/gin-gonic/gin"
    "github.com/joho/godotenv"
    "log"
    "os"
    "your-project/pkg/auth"
    "your-project/pkg/handlers"
)

func main() {
    if err := godotenv.Load(); err != nil {
        log.Fatal("Error loading .env file")
    }

    auth.InitGoogleOAuth()

    r := gin.Default()

    // Set up session middleware
    store := cookie.NewStore([]byte(os.Getenv("SESSION_SECRET")))
    r.Use(sessions.Sessions("wiki_session", store))

    // Load templates
    r.LoadHTMLGlob("templates/*")

    // Serve static files
    r.Static("/static", "./static")

    // Public routes
    r.GET("/login", handlers.LoginHandler)
    r.GET("/auth/google", handlers.GoogleLoginHandler)
    r.GET("/auth/google/callback", handlers.GoogleCallbackHandler)
    r.GET("/logout", handlers.LogoutHandler)

    // Protected routes
    protected := r.Group("/")
    protected.Use(auth.AuthRequired())
    {
        protected.GET("/", handlers.ViewHandler)
        protected.GET("/view/:title", handlers.ViewHandler)
        protected.GET("/edit/:title", handlers.EditHandler)
        protected.POST("/save/:title", handlers.SaveHandler)
    }

    r.Run(":8080")
}
```

## Step 10: Update Existing Templates

Update your existing templates to include user information and logout button. Add this to the navigation area in `view.html` and `edit.html`:

```html
<div class="user-info">
    <span>Welcome, {{ .User.Name }}</span>
    <a href="/logout" class="logout-btn">Logout</a>
</div>
```

## Step 11: Update Handlers

Update your existing handlers to include user information in the template context. For example, in `ViewHandler`:

```go
func ViewHandler(c *gin.Context) {
    title := c.Param("title")
    user := c.MustGet("user").(*auth.UserSession)
    
    // ... existing code ...

    c.HTML(http.StatusOK, "view.html", gin.H{
        "Title": title,
        "Body":  string(body),
        "User":  user,
    })
}
```

## Step 12: Test the Implementation

1. Start the application:
```bash
go run cmd/server/main.go
```

2. Visit `http://localhost:8080`
3. You should be redirected to the login page
4. Click "Sign in with Google"
5. Complete the Google authentication
6. You should be redirected back to the wiki
7. Try accessing protected routes
8. Test the logout functionality

## Troubleshooting

1. **White Screen After Login**
   - Check if all templates are properly loaded
   - Verify template paths and names
   - Check browser console for errors

2. **Session Not Persisting**
   - Verify SESSION_SECRET is set correctly
   - Check if cookies are enabled in the browser
   - Ensure HTTPS is used in production

3. **OAuth Errors**
   - Verify Google OAuth credentials
   - Check redirect URI matches exactly
   - Ensure all required scopes are included

## Security Notes

1. Always use HTTPS in production
2. Keep your OAuth credentials secure
3. Implement proper session timeout
4. Validate all user input
5. Use secure session configuration

## Additional Resources

- [Google OAuth2 Documentation](https://developers.google.com/identity/protocols/oauth2)
- [Gin Framework Documentation](https://gin-gonic.com/docs/)
- [Gin Sessions Documentation](https://github.com/gin-contrib/sessions) 