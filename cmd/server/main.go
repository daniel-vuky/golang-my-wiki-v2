package main

import (
	"fmt"
	"html/template"
	"log"
	"os"

	"github.com/daniel-vuky/golang-my-wiki-v2/pkg/config"
	"github.com/daniel-vuky/golang-my-wiki-v2/pkg/handlers"
	"github.com/daniel-vuky/golang-my-wiki-v2/pkg/middleware"
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
	if cfg.StorageMode == "local" {
		if err := os.MkdirAll(cfg.Server.DataDir, 0755); err != nil {
			log.Fatalf("Failed to create data directory: %v", err)
		}
	}

	// Initialize Gin router
	router := gin.Default()

	// Set up session middleware
	sessionStore := cookie.NewStore([]byte(cfg.Session.Secret))
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

	// Add auth middleware
	router.Use(middleware.AuthMiddleware())

	// Public routes
	router.GET("/", handlers.HomeHandler)
	router.GET("/view/:title", handlers.ViewHandler)
	router.GET("/edit/:title", handlers.EditHandler)
	router.GET("/new", handlers.EditHandler)
	router.POST("/save", handlers.SaveHandler)
	router.POST("/delete/:title", handlers.DeleteHandler)
	router.GET("/delete/:title", handlers.DeleteHandler)

	// Category routes
	router.POST("/category/create", handlers.CategoryCreateHandler)
	router.GET("/category/*path", handlers.CategoryHandler)
	router.GET("/api/folders/children/*path", handlers.GetFolderChildrenHandler)
	router.DELETE("/api/folder/delete", handlers.DeleteFolderHandler)

	// Auth routes
	router.GET("/login", handlers.LoginHandler)
	router.GET("/auth/google", handlers.GoogleLoginHandler)
	router.GET("/auth/google/callback", handlers.GoogleCallbackHandler)
	router.GET("/logout", handlers.LogoutHandler)

	// Start server
	addr := fmt.Sprintf("%s:%s", cfg.Server.Host, cfg.Server.Port)
	log.Printf("Server starting on %s", addr)
	if err := router.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
