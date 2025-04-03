package handlers

import (
	"net/http"

	"github.com/daniel-vuky/golang-my-wiki-v2/pkg/auth"
	"github.com/daniel-vuky/golang-my-wiki-v2/pkg/config"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

var authHandler *auth.Handler

// InitAuthHandlers initializes the auth handlers with configuration
func InitAuthHandlers(cfg *config.Config) {
	authHandler = auth.NewHandler(cfg)
}

func LoginHandler(c *gin.Context) {
	session := sessions.Default(c)
	error := session.Get("error")
	if error != nil {
		// Clear the error message after retrieving it
		session.Delete("error")
		session.Save()
	}
	c.HTML(http.StatusOK, "login.html", gin.H{
		"Error": error,
	})
}

func GoogleLoginHandler(c *gin.Context) {
	authHandler.Login(c)
}

func GoogleCallbackHandler(c *gin.Context) {
	authHandler.Callback(c)
}

func LogoutHandler(c *gin.Context) {
	session := sessions.Default(c)

	// Clear all session data
	session.Clear()

	// Set cookie options to expire immediately and be secure
	session.Options(sessions.Options{
		Path:     "/",
		MaxAge:   -1, // Set to -1 to expire the cookie immediately
		HttpOnly: true,
		Secure:   config.GetConfig().Session.Secure,
		SameSite: http.SameSiteLaxMode,
	})

	// Save the session to ensure changes are applied
	if err := session.Save(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to clear session"})
		return
	}

	// Redirect to login page
	c.Redirect(http.StatusTemporaryRedirect, "/login")
}
