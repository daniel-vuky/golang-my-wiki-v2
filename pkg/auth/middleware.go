package auth

import (
	"log"
	"net/http"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

type User struct {
	Email string
	Name  string
}

func AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)

		// Check if session exists and is valid
		userEmail := session.Get("user_email")
		userName := session.Get("user_name")
		lastActivity := session.Get("last_activity")

		log.Printf("Auth check - Email: %v, Name: %v, Last Activity: %v", userEmail, userName, lastActivity)

		if userEmail == nil || userName == nil || lastActivity == nil {
			log.Printf("Session invalid or expired, redirecting to login")
			// Clear any existing session data
			session.Clear()
			session.Save()
			c.Redirect(http.StatusTemporaryRedirect, "/login")
			c.Abort()
			return
		}

		// Check if session has expired (60 seconds)
		lastActivityTime := time.Unix(lastActivity.(int64), 0)
		if time.Since(lastActivityTime) > 60*time.Second {
			log.Printf("Session expired, redirecting to login")
			session.Clear()
			session.Save()
			c.Redirect(http.StatusTemporaryRedirect, "/login")
			c.Abort()
			return
		}

		// Update last activity time
		session.Set("last_activity", time.Now().Unix())
		session.Save()

		// Set user info in context
		c.Set("user", User{
			Email: userEmail.(string),
			Name:  userName.(string),
		})
		c.Next()
	}
}
