package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"t3sesame/internal/handlers"
	"t3sesame/internal/models"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/gorilla/sessions"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	_ "github.com/lib/pq"
)

func main() {
	// Database connection
	dbHost := getEnv("DB_HOST", "localhost")
	dbPort := getEnv("DB_PORT", "5432")
	dbUser := getEnv("DB_USER", "t3sesame")
	dbPassword := getEnv("DB_PASSWORD", "password")
	dbName := getEnv("DB_NAME", "t3sesame")

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		dbHost, dbPort, dbUser, dbPassword, dbName)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	// Run migrations
	if err := runMigrations(dsn); err != nil {
		log.Fatal("Failed to run migrations:", err)
	}

	// Initialize Echo
	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())

	// Session middleware
	sessionSecret := getEnv("SESSION_SECRET", "your-super-secret-session-key")
	e.Use(session.Middleware(sessions.NewCookieStore([]byte(sessionSecret))))

	// Static files
	e.Static("/static", "static")

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(db)
	chatService := models.NewChatService(db)
	chatHandler := handlers.NewChatHandler(chatService)
	oauthHandler := handlers.NewOAuthHandler(models.NewUserService(db))

	// Routes
	// Guest routes (redirect to dashboard if authenticated)
	guest := e.Group("")
	guest.Use(handlers.GuestMiddleware)
	guest.GET("/", func(c echo.Context) error {
		return c.Redirect(302, "/login")
	})
	guest.GET("/login", authHandler.ShowLogin)
	guest.GET("/register", authHandler.ShowRegister)
	guest.POST("/login", authHandler.Login)
	guest.POST("/register", authHandler.Register)
	guest.GET("/auth/google", oauthHandler.GoogleLogin)
	guest.GET("/auth/google/callback", oauthHandler.GoogleCallback)

	// Protected routes (require authentication)
	protected := e.Group("")
	protected.Use(handlers.AuthMiddleware)
	protected.GET("/", chatHandler.ShowMainInterface)          // Main chat interface
	protected.GET("/dashboard", chatHandler.ShowMainInterface) // Redirect old dashboard
	protected.GET("/chat/:tree_id", chatHandler.GetChatMessages)
	protected.POST("/chat", chatHandler.CreateNewChat)
	protected.POST("/chat/:tree_id/message", chatHandler.SendMessage)
	protected.POST("/logout", authHandler.Logout)

	// Start server
	port := getEnv("PORT", "8080")
	log.Printf("Server starting on port %s", port)
	log.Fatal(e.Start(":" + port))
}

func runMigrations(dsn string) error {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return err
	}
	defer db.Close()

	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return err
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://migrations",
		"postgres", driver)
	if err != nil {
		return err
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return err
	}

	return nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
