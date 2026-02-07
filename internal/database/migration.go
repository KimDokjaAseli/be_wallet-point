package database

import (
	"log"
	"wallet-point/internal/audit"
	"wallet-point/internal/auth"
	"wallet-point/internal/marketplace"
	"wallet-point/internal/mission"
	"wallet-point/internal/wallet"

	"gorm.io/gorm"
)

func Migrate(db *gorm.DB) {
	log.Println("🔄 Running database connectivity check...")

	// We no longer run manual ALTER TABLEs here as they are moved to the SQL database files.
	// We only keep AutoMigrate for safety to ensure the structs match the DB,
	// but the heavy lifting and enums are handled in 01_tables.sql.

	err := db.AutoMigrate(
		&auth.User{},
		&wallet.Wallet{},
		&wallet.WalletTransaction{},
		&wallet.PaymentToken{},
		&marketplace.Product{},
		&marketplace.MarketplaceTransaction{},
		&marketplace.CartItem{},
		&audit.AuditLog{},
		&mission.Mission{},
		&mission.MissionQuestion{},
		&mission.MissionSubmission{},
	)

	if err != nil {
		log.Fatal("❌ Migration failed:", err)
	}

	log.Println("✅ Database migration/sync completed")
}
