package main

import (
	"log"
	"wallet-point/config"
	"wallet-point/internal/database"
	"wallet-point/routes"
	"wallet-point/utils"

	"github.com/gin-gonic/gin"
)

// @title Wallet Point API
// @version 1.0
// @description Platform Wallet Point Gamifikasi Kampus - Backend API
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.email support@campus.edu

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host walletpoint.xeroon.my.id
// @schemes https
// @BasePath /api/v1

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

func main() {
	// Load configuration
	cfg := config.LoadConfig()

	// Set Gin mode
	gin.SetMode(cfg.GinMode)

	// Initialize JWT
	utils.InitJWT(cfg.JWTSecret)

	// Connect to database
	db := config.ConnectDB(cfg)

	// Run migrations
	database.Migrate(db)

	// Initialize Gin
	r := gin.Default()

	// Setup routes
	routes.SetupRoutes(r, db, cfg.AllowedOrigins, cfg.JWTExpiryHours)

	// Start server
	serverAddress := ":" + cfg.ServerPort
	log.Printf("🚀 Server starting on http://%s", cfg.ServerAddress)
	log.Printf("📚 API Documentation (if Swagger enabled): http://%s/swagger/index.html", cfg.ServerAddress)
	log.Printf("🏥 Health Check: http://%s/api/v1/health", cfg.ServerAddress)
	log.Println("✨ Press Ctrl+C to stop the server")

	if err := r.Run(serverAddress); err != nil {
		log.Fatal("❌ Failed to start server:", err)
	}
}
