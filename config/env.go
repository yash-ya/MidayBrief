package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

func LoadEnv() {
	if os.Getenv("RAILWAY_ENVIRONMENT") == "" {
		_ = godotenv.Load()
		err := godotenv.Load()
		if err != nil {
			log.Println("Warning: .env file not loaded")
		}
	}
}
