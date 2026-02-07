package main

import (
	"fmt"
	"log"
	"wallet-point/config"
	"wallet-point/internal/auth"
	"wallet-point/internal/mission"
	"wallet-point/internal/wallet"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func main() {
	// Load config from .env
	cfg := config.LoadConfig()

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		cfg.DBUser, cfg.DBPassword, cfg.DBHost, cfg.DBPort, cfg.DBName)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("=== CEK SKEMA DATABASE TERKINI (GORM) ===")

	// Check Missions Table
	fmt.Println("\n[Tabel: missions]")
	var m mission.Mission
	cols, _ := db.Migrator().ColumnTypes(&m)
	for _, col := range cols {
		tn, _ := col.ColumnType()
		fmt.Printf("- %-18s | %s\n", col.Name(), tn)
	}

	// Check Mission Submissions Table
	fmt.Println("\n[Tabel: mission_submissions]")
	var ms mission.MissionSubmission
	cols, _ = db.Migrator().ColumnTypes(&ms)
	for _, col := range cols {
		tn, _ := col.ColumnType()
		fmt.Printf("- %-18s | %s\n", col.Name(), tn)
	}

	// Check Users Table
	fmt.Println("\n[Tabel: users]")
	var u auth.User
	cols, _ = db.Migrator().ColumnTypes(&u)
	for _, col := range cols {
		tn, _ := col.ColumnType()
		fmt.Printf("- %-18s | %s\n", col.Name(), tn)
	}

	// Check Wallets Table
	fmt.Println("\n[Tabel: wallets]")
	var w wallet.Wallet
	cols, _ = db.Migrator().ColumnTypes(&w)
	for _, col := range cols {
		tn, _ := col.ColumnType()
		fmt.Printf("- %-18s | %s\n", col.Name(), tn)
	}
}
