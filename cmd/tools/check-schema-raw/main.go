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

	tables := []string{"missions", "mission_submissions", "users", "wallets"}

	for _, table := range tables {
		fmt.Printf("\n--- KOLOM TABEL: %s ---\n", table)
		rows, err := db.Raw(fmt.Sprintf("SHOW COLUMNS FROM %s", table)).Rows()
		if err != nil {
			fmt.Printf("Gagal ambil data tabel %s: %v\n", table, err)
			continue
		}
		defer rows.Close()

		for rows.Next() {
			var field, typ, null, key, def, extra interface{}
			rows.Scan(&field, &typ, &null, &key, &def, &extra)
			fmt.Printf("- %-20s | %-20s | NULL:%-5s | KEY:%s\n", field, typ, null, key)
		}
	}
}
