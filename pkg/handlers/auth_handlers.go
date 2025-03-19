package handlers

import (
	"net/http"

	"github.com/daniel-vuky/golang-my-wiki-v2/pkg/auth"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

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
	session := sessions.Default(c)
	state := auth.GenerateRandomState()
	session.Set("oauth_state", state)
	session.Save()

	url := auth.GetGoogleOAuthConfig().AuthCodeURL(state)
	c.Redirect(http.StatusTemporaryRedirect, url)
}

func GoogleCallbackHandler(c *gin.Context) {
	// Use the auth package's Callback function which includes email validation
	auth.NewHandler().Callback(c)
}

func LogoutHandler(c *gin.Context) {
	auth.ClearUserSession(c)
	c.Redirect(http.StatusTemporaryRedirect, "/login")
}
