package middleware

import (
	"net/http"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)
		// Update last activity time
		session.Set("last_activity", time.Now().Unix())
		session.Options(sessions.Options{
			MaxAge: 3600, // 1 hour
		})
		if err := session.Save(); err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Failed to save session"})
			return
		}
		c.Next()
	}
}

func LogoutHandler(c *gin.Context) {
	session := sessions.Default(c)
	session.Clear()
	session.Options(sessions.Options{
		MaxAge: -1,
	})
	session.Save()
	c.Redirect(http.StatusSeeOther, "/")
}
