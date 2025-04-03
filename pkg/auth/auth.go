package auth

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/daniel-vuky/golang-my-wiki-v2/pkg/config"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

type Handler struct {
	oauthConfig *oauth2.Config
	cfg         *config.Config
}

func NewHandler(cfg *config.Config) *Handler {
	return &Handler{
		cfg: cfg,
		oauthConfig: &oauth2.Config{
			ClientID:     cfg.Google.ClientID,
			ClientSecret: cfg.Google.ClientSecret,
			RedirectURL:  cfg.Google.RedirectURL,
			Scopes: []string{
				"https://www.googleapis.com/auth/userinfo.email",
				"https://www.googleapis.com/auth/userinfo.profile",
			},
			Endpoint: google.Endpoint,
		},
	}
}

func InitGoogleOAuth(cfg *config.Config) {
	// Initialize the OAuth config
	_ = NewHandler(cfg)
}

// GenerateRandomState generates a random state string for OAuth
func GenerateRandomState() string {
	b := make([]byte, 32)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}

func (h *Handler) getOAuthConfig() *oauth2.Config {
	return h.oauthConfig
}

func (h *Handler) Login(c *gin.Context) {
	// Get the session
	session := sessions.Default(c)

	// Check if user is already logged in
	if email := session.Get("user_email"); email != nil {
		c.Redirect(http.StatusTemporaryRedirect, "/")
		return
	}

	// Get the OAuth2 config
	oauthConfig := h.getOAuthConfig()

	// Generate random state
	state := GenerateRandomState()
	session.Set("oauth_state", state)
	session.Save()

	// Redirect to Google's consent page
	url := oauthConfig.AuthCodeURL(state, oauth2.AccessTypeOffline)
	c.Redirect(http.StatusTemporaryRedirect, url)
}

func (h *Handler) Callback(c *gin.Context) {
	// Get the session
	session := sessions.Default(c)

	// Verify state
	state := session.Get("oauth_state")
	if state == nil || state != c.Query("state") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid state"})
		return
	}

	// Clear the state after verification
	session.Delete("oauth_state")
	session.Save()

	// Get the OAuth2 config
	oauthConfig := h.getOAuthConfig()

	// Exchange code for token
	code := c.Query("code")
	token, err := oauthConfig.Exchange(c, code)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to exchange token"})
		return
	}

	// Get user info
	client := oauthConfig.Client(c, token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user info"})
		return
	}
	defer resp.Body.Close()

	var userInfo struct {
		ID            string `json:"id"`
		Email         string `json:"email"`
		VerifiedEmail bool   `json:"verified_email"`
		Name          string `json:"name"`
		Picture       string `json:"picture"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode user info"})
		return
	}

	// Check if email is allowed
	allowedEmails := h.cfg.Session.AllowedEmails
	isAllowed := len(allowedEmails) == 0 // If no emails are specified, allow all
	if !isAllowed {
		for _, email := range allowedEmails {
			if email == userInfo.Email {
				isAllowed = true
				break
			}
		}
	}

	if !isAllowed {
		// Set error message in session
		session.Set("error", "Your email is not authorized to access this application")
		session.Save()
		c.Redirect(http.StatusTemporaryRedirect, "/login")
		return
	}

	// Clear any existing session data
	session.Clear()

	// Set new session data
	session.Set("user_email", userInfo.Email)
	session.Set("user_name", userInfo.Name)
	session.Set("user_picture", userInfo.Picture)
	session.Set("last_activity", time.Now().Unix()) // Store as Unix timestamp

	// Save the session
	if err := session.Save(); err != nil {
		log.Printf("Error saving session: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save session"})
		return
	}

	c.Redirect(http.StatusTemporaryRedirect, "/")
}
