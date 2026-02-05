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

	var users []struct {
		ID       uint
		Email    string
		FullName string
		Role     string
		Status   string
	}

	db.Table("users").Select("id, email, full_name, role, status").Find(&users)

	fmt.Println("--- USERS DATA ---")
	for _, u := range users {
		fmt.Printf("ID: %v | Email: %-25s | Name: %-20s | Role: %-10s | Status: %s\n", u.ID, u.Email, u.FullName, u.Role, u.Status)
	}
}
