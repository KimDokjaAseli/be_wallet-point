package main

import (
	"fmt"
	"log"
	"wallet-point/config"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func main() {
	cfg := config.LoadConfig()
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		cfg.DBUser, cfg.DBPassword, cfg.DBHost, cfg.DBPort, cfg.DBName)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Memastikan kolom submission_type ada di tabel missions...")

	// Add column submission_type if not exists
	err = db.Exec("ALTER TABLE missions ADD COLUMN IF NOT EXISTS submission_type ENUM('image', 'file', 'link', 'text') DEFAULT 'text' AFTER status").Error
	if err != nil {
		// If IF NOT EXISTS is not supported (older MariaDB/MySQL), try without it but ignore error if already exists
		fmt.Printf("Mencoba alternatif: %v\n", err)
		db.Exec("ALTER TABLE missions ADD COLUMN submission_type ENUM('image', 'file', 'link', 'text') DEFAULT 'text' AFTER status")
	}

	fmt.Println("Selesai.")
}
