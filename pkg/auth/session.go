package auth

import (
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

type UserSession struct {
	Email   string
	Name    string
	Picture string
}

func SaveUserSession(c *gin.Context, user *UserSession) {
	session := sessions.Default(c)
	session.Set("user_email", user.Email)
	session.Set("user_name", user.Name)
	session.Set("user_picture", user.Picture)
	session.Save()
}

func GetUserSession(c *gin.Context) *UserSession {
	session := sessions.Default(c)
	userEmail := session.Get("user_email")
	userName := session.Get("user_name")
	userPicture := session.Get("user_picture")

	if userEmail == nil || userName == nil {
		return nil
	}

	return &UserSession{
		Email:   userEmail.(string),
		Name:    userName.(string),
		Picture: userPicture.(string),
	}
}

func ClearUserSession(c *gin.Context) {
	session := sessions.Default(c)
	session.Clear()
	session.Save()
}
