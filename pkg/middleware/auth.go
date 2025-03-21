package middleware

import (
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

type User struct {
	Email string
	Name  string
}

// AuthMiddleware sets the user information in the context
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)
		email := session.Get("user_email")
		name := session.Get("user_name")

		if email != nil {
			c.Set("user", User{
				Email: email.(string),
				Name:  name.(string),
			})
		} else {
			c.Set("user", User{
				Email: "",
				Name:  "Guest",
			})
		}

		c.Next()
	}
}
