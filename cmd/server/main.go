package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/daniel-vuky/golang-my-wiki-v2/pkg/auth"
	"github.com/daniel-vuky/golang-my-wiki-v2/pkg/config"
	"github.com/daniel-vuky/golang-my-wiki-v2/pkg/handlers"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
)

func main() {
	// Load configuration
	if err := config.LoadConfig(); err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize Google OAuth
	auth.InitGoogleOAuth()

	r := gin.Default()

	// Set up session middleware
	store := cookie.NewStore([]byte(config.GetSessionSecret()))
	store.Options(sessions.Options{
		Path:     "/",
		MaxAge:   24 * 60 * 60, // 24 hours
		HttpOnly: true,
		Secure:   config.GetSessionSecure(),
		SameSite: http.SameSiteLaxMode,
	})
	r.Use(sessions.Sessions(config.GetSessionName(), store))

	// Load templates with custom functions
	handlers.LoadTemplates(r)

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
		protected.GET("/", handlers.HomeHandler)
		protected.GET("/view/:title", handlers.ViewHandler)
		protected.GET("/edit/:title", handlers.EditHandler)
		protected.POST("/save/:title", handlers.SaveHandler)
		protected.GET("/delete/:title", handlers.DeleteHandler)
	}

	// Start server
	addr := fmt.Sprintf("%s:%s", config.GetServerHost(), config.GetServerPort())
	if err := r.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
