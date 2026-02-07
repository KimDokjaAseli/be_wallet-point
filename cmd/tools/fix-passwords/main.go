package main

import (
	"log"
	"wallet-point/config"
	"wallet-point/internal/auth"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func main() {
	// Load config
	cfg := config.LoadConfig()

	// Direct connection setup to avoid circular deps or complex init
	dsn := cfg.DBUser + ":" + cfg.DBPassword + "@tcp(" + cfg.DBHost + ":" + cfg.DBPort + ")/" + cfg.DBName + "?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	log.Println("✅ Connected to database")

	// Plain text password for testing
	newPassword := "Password123!"

	// Update all users
	result := db.Model(&auth.User{}).Where("1=1").Update("password_hash", newPassword)
	if result.Error != nil {
		log.Fatal("Failed to update passwords:", result.Error)
	}

	log.Printf("✅ Successfully updated %d users to use plaintext password: '%s'\n", result.RowsAffected, newPassword)
	log.Println("👉 You can now login with this password.")
}
