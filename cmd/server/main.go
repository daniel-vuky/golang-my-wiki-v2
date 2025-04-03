package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"

	"github.com/daniel-vuky/golang-my-wiki-v2/pkg/auth"
	"github.com/daniel-vuky/golang-my-wiki-v2/pkg/config"
	"github.com/daniel-vuky/golang-my-wiki-v2/pkg/handlers"
	"github.com/daniel-vuky/golang-my-wiki-v2/pkg/storage"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
)

func main() {
	// Load configuration
	if err := config.LoadConfig(); err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	cfg := config.GetConfig()

	// Initialize storage
	store, err := storage.NewStorage(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize storage: %v", err)
	}

	// Create data directory if it doesn't exist
	if err := os.MkdirAll(cfg.Server.DataDir, 0755); err != nil {
		log.Fatalf("Failed to create data directory: %v", err)
	}

	// Initialize Gin router
	router := gin.Default()

	// Set up session middleware
	sessionStore := cookie.NewStore([]byte(cfg.Session.Secret))
	sessionStore.Options(sessions.Options{
		Path:     "/",
		MaxAge:   3600, // 1 hour
		HttpOnly: true,
		Secure:   cfg.Session.Secure, // Use secure cookies in production
		SameSite: http.SameSiteLaxMode,
		Domain:   "", // Empty domain to ensure cookie is only sent to the exact domain
	})
	log.Printf("Initializing session store with MaxAge: 3600 seconds, Secure: %v", cfg.Session.Secure)
	router.Use(sessions.Sessions("wiki_session", sessionStore))

	// Set up template functions
	router.SetFuncMap(template.FuncMap{
		"add": func(a, b int) int {
			return a + b
		},
		"sub": func(a, b int) int {
			return a - b
		},
		"html": func(value interface{}) template.HTML {
			return template.HTML(fmt.Sprint(value))
		},
	})

	// Set up static files
	router.Static("/static", "./static")
	router.LoadHTMLGlob("templates/*.html")

	// Initialize handlers with storage
	handlers.InitHandlers(store)

	// Initialize auth handlers
	handlers.InitAuthHandlers(cfg)

	// Auth routes (no auth required)
	router.GET("/login", handlers.LoginHandler)
	router.GET("/auth/google", handlers.GoogleLoginHandler)
	router.GET("/auth/google/callback", handlers.GoogleCallbackHandler)
	router.GET("/logout", handlers.LogoutHandler)

	// Protected routes (auth required)
	protected := router.Group("/")
	protected.Use(auth.AuthRequired())
	{
		protected.GET("/", handlers.HomeHandler)
		protected.GET("/view/:title", handlers.ViewHandler)
		protected.GET("/edit/:title", handlers.EditHandler)
		protected.GET("/new", handlers.EditHandler)
		protected.POST("/save", handlers.SaveHandler)
		protected.POST("/delete/:title", handlers.DeleteHandler)
		protected.GET("/delete/:title", handlers.DeleteHandler)

		// Category routes
		protected.POST("/category/create", handlers.CategoryCreateHandler)
		protected.GET("/category/*path", handlers.CategoryHandler)
		protected.GET("/api/folders/children/*path", handlers.GetFolderChildrenHandler)
		protected.DELETE("/api/folder/delete", handlers.DeleteFolderHandler)

		// Sync route
		protected.POST("/api/sync", handlers.HandleSync)
	}

	// Start server
	addr := fmt.Sprintf("%s:%s", cfg.Server.Host, cfg.Server.Port)
	log.Printf("Server starting on %s", addr)
	if err := router.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
