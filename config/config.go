// config/config.go
package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	TelegramToken     string
	GoogleSheetID     string
	CredentialsJSON   string
	FirebaseProjectId string
}

func Load() Config {
	err := godotenv.Load()
	if err != nil {
		log.Println("⚠️ No .env file found, using system environment variables.")
	}
	cfg := Config{
		TelegramToken:     os.Getenv("TELEGRAM_TOKEN"),
		CredentialsJSON:   os.Getenv("GOOGLE_CREDENTIALS_FILE"),
		FirebaseProjectId: os.Getenv("FIREBASE_PROJECT_ID"),
	}
	if cfg.TelegramToken == "" || cfg.CredentialsJSON == "" {
		log.Fatal("❌ Missing required environment variables")
	}
	return cfg
}
