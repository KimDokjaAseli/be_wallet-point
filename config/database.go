package config

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func ConnectDB(cfg *Config) *gorm.DB {
	// Build DSN (Data Source Name)
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		cfg.DBUser,
		cfg.DBPassword,
		cfg.DBHost,
		cfg.DBPort,
		cfg.DBName,
	)

	// Configure GORM
	gormConfig := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
		NowFunc: func() time.Time {
			return time.Now().Local()
		},
	}

	// Retry loop for database connection
	var db *gorm.DB
	var err error
	maxRetries := 15

	for i := 1; i <= maxRetries; i++ {
		db, err = gorm.Open(mysql.Open(dsn), gormConfig)
		if err == nil {
			var sqlDB *sql.DB // We need database/sql imported, wait, let's just use db.DB()
			sqlDB, err = db.DB()
			if err == nil {
				err = sqlDB.Ping()
				if err == nil {
					break // Connection successful
				}
			}
		}

		log.Printf("⏳ Database connection attempt %d/%d failed: %v. Retrying in 5s...", i, maxRetries, err)
		if i < maxRetries {
			time.Sleep(5 * time.Second)
		}
	}

	if err != nil {
		log.Fatal("❌ Failed to connect to database after retries:", err)
	}

	// Get underlying SQL DB for connection pool configuration
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatal("Failed to get database instance:", err)
	}

	// Connection pool settings
	sqlDB.SetMaxIdleConns(10)                  // Maximum idle connections
	sqlDB.SetMaxOpenConns(100)                 // Maximum open connections
	sqlDB.SetConnMaxLifetime(5 * time.Minute)  // Connection lifetime (shortened to prevent proxy drop EOF)

	log.Println("✅ Database connected successfully")
	return db
}
