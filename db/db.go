package db

import (
	"log"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func Init() {
	dns := os.Getenv("DATABASE_URL")
	var err error
	DB, err = gorm.Open(postgres.Open(dns), &gorm.Config{})
	if err != nil {
        log.Fatal("❌ Failed to connect to DB:", err)
    }

    log.Println("✅ Connected to DB")
}