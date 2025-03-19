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
	userEmail := session.Get("user_email")
	userName := session.Get("user_name")

	if userEmail == nil || userName == nil {
		return nil
	}

	return &UserSession{
		ID:    "", // ID is optional
		Email: userEmail.(string),
		Name:  userName.(string),
	}
}

func ClearUserSession(c *gin.Context) {
	session := sessions.Default(c)
	session.Clear()
	session.Save()
}
